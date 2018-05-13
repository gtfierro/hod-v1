package db

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sort"
	"time"

	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/turtle"
	"github.com/mitghi/btree"
	"github.com/zhangxinngang/murmur"
)

func hashURI(u turtle.URI, dest []byte, salt uint64) {
	var hash uint32
	if len(dest) < 8 {
		dest = make([]byte, 8)
	}
	if salt > 0 {
		saltbytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(saltbytes, salt)
		hash = murmur.Murmur3(append(u.Bytes(), saltbytes...))
	} else {
		hash = murmur.Murmur3(u.Bytes())
	}
	binary.LittleEndian.PutUint32(dest[:4], hash)
}

func mustGetURI(graph traversable, hash Key) turtle.URI {
	if uri, err := graph.getURI(hash); err != nil {
		panic(err)
	} else {
		return uri
	}
}

func dumpHashTree(tree *btree.BTree, db *DB, limit int) {
	max := tree.Max()
	iter := func(i btree.Item) bool {
		if limit == 0 {
			return false // stop iteration
		} else if limit > 0 {
			limit -= 1 //
		}
		fmt.Println(db.MustGetURI(i.(Key)))
		return i != max
	}
	tree.Ascend(iter)
}

func dumpEntityTree(tree *btree.BTree, db *DB, limit int) {
	max := tree.Max()
	iter := func(i btree.Item) bool {
		if limit == 0 {
			return false // stop iteration
		} else if limit > 0 {
			limit -= 1 //
		}
		fmt.Println(db.MustGetURI(i.(*Entity).PK))
		return i != max
	}
	tree.Ascend(iter)
}

func compareResultMapList(rml1, rml2 []ResultMap) bool {
	var (
		found bool
	)

	if len(rml1) != len(rml2) {
		return false
	}

	for _, val1 := range rml1 {
		found = false
		for _, val2 := range rml2 {
			if compareResultMap(val1, val2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func compareResultMap(rm1, rm2 ResultMap) bool {
	if len(rm1) != len(rm2) {
		return false
	}
	for k, v := range rm1 {
		if v2, found := rm2[k]; !found {
			return false
		} else if v2 != v {
			return false
		}
	}
	return true
}

func rowIsFull(row []Key) bool {
	for _, entry := range row {
		if entry == emptyKey {
			return false
		}
	}
	return true
}

func reversePath(path []sparql.PathPattern) []sparql.PathPattern {
	newpath := make([]sparql.PathPattern, len(path))
	// for in-place, replace newpath with path
	if len(newpath) == 1 {
		return path
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		newpath[i], newpath[j] = path[j], path[i]
	}
	return newpath
}

func hashQuery(q *sparql.Query) []byte {
	h := fnv.New64a()
	var selectVars = make(sort.StringSlice, len(q.Select.Vars))
	for idx, varname := range q.Select.Vars {
		selectVars[idx] = varname
	}
	for _, hv := range selectVars {
		h.Write([]byte(hv))
	}

	var triples []string
	q.IterTriples(func(triple sparql.Triple) sparql.Triple {
		triples = append(triples, triple.String())
		return triple
	})

	x := sort.StringSlice(triples)
	x.Sort()
	for _, hv := range x {
		h.Write([]byte(hv))
	}

	return h.Sum(nil)
}

type queryStats struct {
	ExecutionTime time.Duration
	ExpandTime    time.Duration
	NumResults    int
	NumInserted   int
	NumDeleted    int
}

func (mq *queryStats) merge(other queryStats) {
	if mq == nil {
		mq = &other
		return
	}
	if mq.ExecutionTime.Nanoseconds() < other.ExecutionTime.Nanoseconds() {
		mq.ExecutionTime = other.ExecutionTime
	}
	if mq.ExpandTime.Nanoseconds() < other.ExpandTime.Nanoseconds() {
		mq.ExpandTime = other.ExpandTime
	}
	mq.NumResults += other.NumResults
	mq.NumInserted += other.NumInserted
	mq.NumDeleted += other.NumDeleted
}
