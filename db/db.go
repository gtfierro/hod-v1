package db

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"

	turtle "github.com/gtfierro/hod/goraptor"
	query "github.com/gtfierro/hod/query"

	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tinylib/msgp/msgp"
	"github.com/zhangxinngang/murmur"
)

// logger
var log *logging.Logger
var emptyHash = [4]byte{0, 0, 0, 0}

func init() {
	log = logging.MustGetLogger("hod")
	var format = "%{color}%{level} %{shortfile} %{time:Jan 02 15:04:05} %{color:reset} ▶ %{message}"
	var logBackend = logging.NewLogBackend(os.Stderr, "", 0)
	logBackendLeveled := logging.AddModuleLevel(logBackend)
	logging.SetBackend(logBackendLeveled)
	logging.SetFormatter(logging.MustStringFormatter(format))
}

type DB struct {
	path string
	// store []byte(entity URI) => primary key
	entityDB *leveldb.DB
	// store primary key => [](entity URI)
	pkDB *leveldb.DB
	// predicate index: stores "children" of predicates
	predDB    *leveldb.DB
	predIndex map[turtle.URI]*PredicateEntity
	// graph structure
	graphDB *leveldb.DB
	// store relationships and their inverses
	relationships map[turtle.URI]turtle.URI
	// store the namespace prefixes as strings
	namespaces map[string]string
}

func NewDB(path string) (*DB, error) {
	path = strings.TrimSuffix(path, "/")

	// set up entity, pk databases
	entityDBPath := path + "/db-entities"
	entityDB, err := leveldb.OpenFile(entityDBPath, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open entityDB file %s", entityDBPath)
	}

	pkDBPath := path + "/db-pk"
	pkDB, err := leveldb.OpenFile(pkDBPath, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open pkDB file %s", pkDBPath)
	}

	graphDBPath := path + "/db-graph"
	graphDB, err := leveldb.OpenFile(graphDBPath, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open graphDB file %s", graphDBPath)
	}
	predDBPath := path + "/db-pred"
	predDB, err := leveldb.OpenFile(predDBPath, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open predDB file %s", predDBPath)
	}

	db := &DB{
		path:          path,
		entityDB:      entityDB,
		pkDB:          pkDB,
		graphDB:       graphDB,
		predDB:        predDB,
		predIndex:     make(map[turtle.URI]*PredicateEntity),
		relationships: make(map[turtle.URI]turtle.URI),
		namespaces:    make(map[string]string),
	}

	// TODO: load predIndex and relationships from database
	predIndexPath := path + "/predIndex"
	relshipIndexPath := path + "/relshipIndex"
	namespaceIndexPath := path + "/namespaceIndex"
	if _, err := os.Stat(predIndexPath); !os.IsNotExist(err) {
		f, err := os.Open(predIndexPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not open predIndex file %s", predIndexPath)
		}
		var pi = new(PredIndex)
		if err := msgp.Decode(f, pi); err != nil {
			return nil, err
		}
		for uri, pe := range *pi {
			db.predIndex[turtle.ParseURI(uri)] = pe
		}
	}
	if _, err := os.Stat(relshipIndexPath); !os.IsNotExist(err) {
		f, err := os.Open(relshipIndexPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not open relshipIndexPath file %s", relshipIndexPath)
		}
		var ri = new(RelshipIndex)
		if err := msgp.Decode(f, ri); err != nil {
			return nil, err
		}
		for uri, uri2 := range *ri {
			db.relationships[turtle.ParseURI(uri)] = turtle.ParseURI(uri2)
		}
	}
	if _, err := os.Stat(namespaceIndexPath); !os.IsNotExist(err) {
		f, err := os.Open(namespaceIndexPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not open namespaceIndexPath file %s", namespaceIndexPath)
		}
		var ni = new(NamespaceIndex)
		if err := msgp.Decode(f, ni); err != nil {
			return nil, err
		}
		for ns, full := range *ni {
			db.namespaces[ns] = full
		}
		log.Notice("loaded namespace index")
	}

	return db, nil
}

// hashes the given URI into the byte array
func (db *DB) hashURI(u turtle.URI, dest []byte, salt uint32) {
	var hash uint32
	if len(dest) < 4 {
		dest = make([]byte, 4)
	}
	if salt > 0 {
		saltbytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(saltbytes, salt)
		hash = murmur.Murmur3(append(u.Bytes(), saltbytes...))
	} else {
		hash = murmur.Murmur3(u.Bytes())
	}
	binary.LittleEndian.PutUint32(dest, hash)
}

func (db *DB) insertEntity(entity turtle.URI, hashdest []byte) error {
	// check if we've inserted Subject already
	if exists, err := db.entityDB.Has(entity.Bytes(), nil); err == nil && exists {
		// populate hash anyway
		hash, err := db.entityDB.Get(entity.Bytes(), nil)
		copy(hashdest, hash[:])
		return err
	} else if err != nil {
		return errors.Wrapf(err, "Error checking db membership for %s", entity.String())
	}
	// generate the hash
	var salt = uint32(0)
	db.hashURI(entity, hashdest, salt)
	for {
		if exists, err := db.pkDB.Has(hashdest, nil); err == nil && exists {
			log.Warning("hash exists")
			salt += 1
			db.hashURI(entity, hashdest, salt)
		} else if err != nil {
			return errors.Wrapf(err, "Error checking db membership for %v", hashdest)
		} else {
			break
		}
	}

	// insert the hash into the entity and prefix dbs
	if err := db.entityDB.Put(entity.Bytes(), hashdest, nil); err != nil {
		return errors.Wrapf(err, "Error inserting entity %s", entity.String())
	}
	if err := db.pkDB.Put(hashdest, entity.Bytes(), nil); err != nil {
		return errors.Wrapf(err, "Error inserting pk %s", hashdest)
	}
	return nil
}

// for each part of the triple (subject, predicate, object), we check if its already in the entity database.
// If it is, we can skip it. If not, we generate a murmur3 hash for the entity, and then
// 0. check if we've already inserted the entity (skip if we already have)
// 1. check if the hash is unique (check membership in pk db) - if it isn't then we add a salt and check again
// 2. insert hash => []byte(entity) into pk db
// 3. insert []byte(entity) => hash into entity db
func (db *DB) insertEntityTx(entity turtle.URI, hashdest []byte, enttx, pktx *leveldb.Transaction) error {
	// check if we've inserted Subject already
	if exists, err := enttx.Has(entity.Bytes(), nil); err == nil && exists {
		// populate hash anyway
		hash, err := enttx.Get(entity.Bytes(), nil)
		copy(hashdest, hash[:])
		return err
	} else if err != nil {
		return errors.Wrapf(err, "Error checking db membership for %s", entity.String())
	}
	// generate the hash
	var salt = uint32(0)
	db.hashURI(entity, hashdest, salt)
	for {
		if exists, err := pktx.Has(hashdest, nil); err == nil && exists {
			log.Warning("hash exists")
			salt += 1
			db.hashURI(entity, hashdest, salt)
		} else if err != nil {
			return errors.Wrapf(err, "Error checking db membership for %v", hashdest)
		} else {
			break
		}
	}

	// insert the hash into the entity and prefix dbs
	if err := enttx.Put(entity.Bytes(), hashdest, nil); err != nil {
		return errors.Wrapf(err, "Error inserting entity %s", entity.String())
	}
	if err := pktx.Put(hashdest, entity.Bytes(), nil); err != nil {
		return errors.Wrapf(err, "Error inserting pk %s", hashdest)
	}
	return nil
}

func (db *DB) loadPredicateEntity(predicate turtle.URI, _predicateHash, _subjectHash, _objectHash []byte, predtx *leveldb.Transaction) error {
	var (
		pred          *PredicateEntity
		found         bool
		predicateHash [4]byte
		subjectHash   [4]byte
		objectHash    [4]byte
	)
	copy(predicateHash[:], _predicateHash)
	copy(subjectHash[:], _subjectHash)
	copy(objectHash[:], _objectHash)

	if pred, found = db.predIndex[predicate]; !found {
		pred = NewPredicateEntity()
		pred.PK = predicateHash
	}

	pred.AddSubjectObject(subjectHash, objectHash)
	db.predIndex[predicate] = pred

	if reverse, found := db.relationships[predicate]; found {
		if pred, found = db.predIndex[reverse]; !found {
			pred = NewPredicateEntity()
			pred.PK = predicateHash
		}
		pred.AddSubjectObject(objectHash, subjectHash)
		db.predIndex[predicate] = pred
	}

	return nil
}

func (db *DB) SaveIndexes() error {
	f, err := os.Create(db.path + "/predIndex")
	if err != nil {
		return err
	}

	pi := make(PredIndex)
	for uri, pe := range db.predIndex {
		pi[uri.String()] = pe
	}

	if err := msgp.Encode(f, pi); err != nil {
		return err
	}

	f, err = os.Create(db.path + "/relshipIndex")
	if err != nil {
		return err
	}

	ri := make(RelshipIndex)
	for uri, uri2 := range db.relationships {
		ri[uri.String()] = uri2.String()
	}

	if err := msgp.Encode(f, ri); err != nil {
		return err
	}

	f, err = os.Create(db.path + "/namespaceIndex")
	if err != nil {
		return err
	}
	if err := msgp.Encode(f, NamespaceIndex(db.namespaces)); err != nil {
		return err
	}

	return nil
}

func (db *DB) LoadRelationships(dataset turtle.DataSet) error {
	// iterate through dataset, and pull out all that have a "rdf:type" of "owl:ObjectProperty"
	// then we want to find the mapping that has "owl:inverseOf"
	var relationships = make(map[turtle.URI]struct{})

	rdf_namespace, found := dataset.Namespaces["rdf"]
	if !found {
		return errors.New("Relationships has no rdf namespace")
	}
	owl_namespace, found := dataset.Namespaces["owl"]
	if !found {
		return errors.New("Relationships has no owl namespace")
	}

	for _, triple := range dataset.Triples {
		if triple.Predicate.Namespace == rdf_namespace &&
			triple.Predicate.Value == "type" &&
			triple.Object.Namespace == owl_namespace &&
			triple.Object.Value == "ObjectProperty" {
			relationships[triple.Subject] = struct{}{}
		}
	}

	for _, triple := range dataset.Triples {
		if triple.Predicate.Namespace == owl_namespace && triple.Predicate.Value == "inverseOf" {
			// check that the subject/object of the inverseOf relationships are both actually relationships
			if _, found := relationships[triple.Subject]; !found {
				continue
			}
			if _, found := relationships[triple.Object]; !found {
				continue
			}
			db.relationships[triple.Subject] = triple.Object
			db.relationships[triple.Object] = triple.Subject
		}
	}

	return nil
}

func (db *DB) LoadDataset(dataset turtle.DataSet) error {
	start := time.Now()
	// merge, don't set outright
	for abbr, full := range dataset.Namespaces {
		db.namespaces[abbr] = full
	}
	// start transactions
	enttx, err := db.entityDB.OpenTransaction()
	if err != nil {
		return errors.Wrap(err, "Could not open transaction on entity dataset")
	}
	pktx, err := db.pkDB.OpenTransaction()
	if err != nil {
		return errors.Wrap(err, "Could not open transaction on pk dataset")
	}
	predtx, err := db.predDB.OpenTransaction()
	if err != nil {
		return errors.Wrap(err, "Could not open transaction on pred dataset")
	}
	// load triples and primary keys
	var (
		subjectHash   = make([]byte, 4)
		predicateHash = make([]byte, 4)
		objectHash    = make([]byte, 4)
	)
	for _, triple := range dataset.Triples {
		if err := db.insertEntityTx(triple.Subject, subjectHash, enttx, pktx); err != nil {
			return err
		}
		if err := db.insertEntityTx(db.relationships[triple.Predicate], predicateHash, enttx, pktx); err != nil {
			return err
		}
		if err := db.insertEntityTx(triple.Predicate, predicateHash, enttx, pktx); err != nil {
			return err
		}
		if err := db.insertEntityTx(triple.Object, objectHash, enttx, pktx); err != nil {
			return err
		}
		if err := db.loadPredicateEntity(triple.Predicate, predicateHash, subjectHash, objectHash, predtx); err != nil {
			return err
		}
	}

	for pred, _ := range db.relationships {
		if err := db.insertEntityTx(pred, predicateHash, enttx, pktx); err != nil {
			return err
		}
	}

	// finish those transactions
	if err := enttx.Commit(); err != nil {
		return errors.Wrap(err, "Could not commit transaction")
	}
	if err := pktx.Commit(); err != nil {
		return errors.Wrap(err, "Could not commit transaction")
	}
	if err := predtx.Commit(); err != nil {
		return errors.Wrap(err, "Could not commit transaction")
	}
	log.Infof("Built lookup tables in %s", time.Since(start))

	start = time.Now()
	if err := db.buildGraph(dataset); err != nil {
		return errors.Wrap(err, "Could not build graph")
	}
	log.Infof("Built graph in %s", time.Since(start))

	for pfx, uri := range db.namespaces {
		fmt.Printf("%s => %s\n", pfx, uri)
	}
	return nil
}

// returns the uint32 hash of the given URI (this is adjusted for uniqueness)
func (db *DB) GetHash(entity turtle.URI) ([4]byte, error) {
	var hash [4]byte
	val, err := db.entityDB.Get(entity.Bytes(), nil)
	if err != nil {
		return emptyHash, err
	}
	copy(hash[:], val)
	if hash == emptyHash {
		return emptyHash, errors.New("Got bad hash")
	}
	return hash, nil
}

func (db *DB) MustGetHash(entity turtle.URI) [4]byte {
	val, err := db.GetHash(entity)
	if err != nil {
		panic(err)
	}
	return val
}

func (db *DB) GetURI(hash [4]byte) (turtle.URI, error) {
	val, err := db.pkDB.Get(hash[:], nil)
	if err != nil {
		return turtle.URI{}, err
	}
	return turtle.ParseURI(string(val)), nil
}

func (db *DB) MustGetURI(hash [4]byte) turtle.URI {
	val, err := db.pkDB.Get(hash[:], nil)
	if err != nil {
		panic(err)
	}
	return turtle.ParseURI(string(val))
}

func (db *DB) MustGetURIStringHash(hash string) turtle.URI {
	var c [4]byte
	copy(c[:], []byte(hash))
	val, err := db.pkDB.Get(c[:], nil)
	if err != nil {
		panic(err)
	}
	return turtle.ParseURI(string(val))
}

func (db *DB) GetEntity(uri turtle.URI) (*Entity, error) {
	var entity = NewEntity()
	hash, err := db.GetHash(uri)
	if err != nil {
		return nil, err
	}
	bytes, err := db.graphDB.Get(hash[:], nil)
	if err != nil {
		return nil, err
	}
	_, err = entity.UnmarshalMsg(bytes)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (db *DB) GetEntityFromHash(hash [4]byte) (*Entity, error) {
	bytes, err := db.graphDB.Get(hash[:], nil)
	if err != nil {
		return nil, err
	}
	ent := NewEntity()
	_, err = ent.UnmarshalMsg(bytes)
	return ent, err
}

func (db *DB) MustGetEntityFromHash(hash [4]byte) *Entity {
	e, err := db.GetEntityFromHash(hash)
	if err != nil {
		panic(err)
	}
	return e
}

func (db *DB) DumpEntity(ent *Entity) {
	fmt.Println("DUMPING", db.MustGetURI(ent.PK))
	for edge, list := range ent.OutEdges {
		fmt.Printf(" OUT: %s \n", db.MustGetURIStringHash(edge).Value)
		for _, l := range list {
			fmt.Printf("     -> %s\n", db.MustGetURI(l).Value)
		}
	}
	for edge, list := range ent.InEdges {
		fmt.Printf(" In: %s \n", db.MustGetURIStringHash(edge).Value)
		for _, l := range list {
			fmt.Printf("     <- %s\n", db.MustGetURI(l).Value)
		}
	}
}

func (db *DB) GetEntityTx(graphtx *leveldb.Transaction, uri turtle.URI) (*Entity, error) {
	var entity = NewEntity()
	hash, err := db.GetHash(uri)
	if err != nil {
		return nil, err
	}
	bytes, err := graphtx.Get(hash[:], nil)
	if err != nil {
		return nil, err
	}
	_, err = entity.UnmarshalMsg(bytes)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (db *DB) GetEntityFromHashTx(graphtx *leveldb.Transaction, hash [4]byte) (*Entity, error) {
	bytes, err := graphtx.Get(hash[:], nil)
	if err != nil {
		return nil, err
	}
	ent := NewEntity()
	_, err = ent.UnmarshalMsg(bytes)
	return ent, err
}

func (db *DB) expandFilter(filter query.Filter) query.Filter {
	if !strings.HasPrefix(filter.Subject.Value, "?") {
		if full, found := db.namespaces[filter.Subject.Namespace]; found {
			filter.Subject.Namespace = full
		}
	}
	if !strings.HasPrefix(filter.Object.Value, "?") {
		if full, found := db.namespaces[filter.Object.Namespace]; found {
			filter.Object.Namespace = full
		}
	}
	for idx2, pred := range filter.Path {
		if !strings.HasPrefix(pred.Predicate.Value, "?") {
			if full, found := db.namespaces[pred.Predicate.Namespace]; found {
				pred.Predicate.Namespace = full
			}
			filter.Path[idx2] = pred
		}
	}
	return filter
}

func (db *DB) expandOrClauseFilters(orc query.OrClause) query.OrClause {
	for fidx, filter := range orc.Terms {
		orc.Terms[fidx] = db.expandFilter(filter)
	}
	for fidx, filter := range orc.LeftTerms {
		orc.LeftTerms[fidx] = db.expandFilter(filter)
	}
	for fidx, filter := range orc.RightTerms {
		orc.RightTerms[fidx] = db.expandFilter(filter)
	}
	for oidx, oc := range orc.LeftOr {
		orc.LeftOr[oidx] = db.expandOrClauseFilters(oc)
	}
	for oidx, oc := range orc.RightOr {
		orc.RightOr[oidx] = db.expandOrClauseFilters(oc)
	}
	return orc
}
