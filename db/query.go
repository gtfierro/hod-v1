package db

// make a set of structs that capture what these queries want to do

type Query struct {
	Select SelectClause
	Where  []Filter
}

type SelectClause struct {
	Variables []string
}

type Filter struct {
	Subject string
	Path    []PathPattern
	Object  string
}

type PathPattern struct {
	Predicate string
}

func (db *DB) RunQuery(q Query) {
}
