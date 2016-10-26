package turtle

type DataSet struct {
	triplecount int
	nscount     int
	Namespaces  map[string]string
	Triples     []Triple
}

func newDataSet() *DataSet {
	return &DataSet{
		triplecount: 0,
		nscount:     0,
		Namespaces:  make(map[string]string),
		Triples:     []Triple{},
	}
}

func (d *DataSet) addTriple(subject, predicate, object string) {
	d.triplecount += 1
	d.Triples = append(d.Triples, MakeTriple(subject, predicate, object))
}

func (d *DataSet) addNamespace(prefix, namespace string) {
	d.nscount += 1
	d.Namespaces[prefix] = namespace
}

func (d *DataSet) NumTriples() int {
	return d.triplecount
}

func (d *DataSet) NumNamespaces() int {
	return d.nscount
}
