package db

import (
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/turtle"
)

func (db *DB) handleInsert(q *sparql.Query, result QueryResult) error {
	var insert turtle.DataSet
	for _, insertTerm := range q.Insert.Terms {
		if result.Count == 0 {
			insert.AddTripleURIs(insertTerm.Subject, insertTerm.Predicates[0].Predicate, insertTerm.Object)
		} else {
			for _, row := range result.Rows {
				// replace all variables with content from query
				if insertTerm.Subject.IsVariable() {
					if value, found := row[insertTerm.Subject.Value]; found {
						insertTerm.Subject = value
					}
				}
				pred := insertTerm.Predicates[0].Predicate
				if pred.IsVariable() {
					if value, found := row[pred.Value]; found {
						insertTerm.Predicates[0].Predicate = value
					}
				}
				if insertTerm.Object.IsVariable() {
					if value, found := row[insertTerm.Object.Value]; found {
						insertTerm.Object = value
					}
				}
				insert.AddTripleURIs(insertTerm.Subject, insertTerm.Predicates[0].Predicate, insertTerm.Object)
			}
		}
	}

	tx, err := db.openTransaction()
	if err != nil {
		tx.discard()
		return err
	}
	if err := tx.addTriples(insert); err != nil {
		tx.discard()
		return err
	}
	if err := tx.commit(); err != nil {
		tx.discard()
		return err
	}
	err = db.saveIndexes()
	if err != nil {
		return err
	}
	return nil
}
