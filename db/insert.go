package db

import (
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/turtle"
	"time"
)

func (db *DB) handleInsert(insert sparql.InsertClause, result QueryResult) (queryStats, error) {
	insertStart := time.Now()
	var additions turtle.DataSet
	var stats queryStats
	for _, insertTerm := range insert.Terms {
		if result.Count == 0 {
			additions.AddTripleURIs(insertTerm.Subject, insertTerm.Predicates[0].Predicate, insertTerm.Object)
		} else {
			for _, row := range result.Rows {
				newterm := insertTerm.Copy()
				// replace all variables with content from query
				if newterm.Subject.IsVariable() {
					if value, found := row[newterm.Subject.Value]; found {
						newterm.Subject = value
					}
				}
				pred := newterm.Predicates[0].Predicate
				if pred.IsVariable() {
					if value, found := row[pred.Value]; found {
						newterm.Predicates[0].Predicate = value
					}
				}
				if newterm.Object.IsVariable() {
					if value, found := row[newterm.Object.Value]; found {
						newterm.Object = value
					}
				}
				additions.AddTripleURIs(newterm.Subject, newterm.Predicates[0].Predicate, newterm.Object)
				stats.NumInserted += 1
			}
		}
	}

	tx, err := db.openTransaction()
	if err != nil {
		tx.discard()
		return stats, err
	}
	if err := tx.addTriples(additions); err != nil {
		tx.discard()
		return stats, err
	}
	if err := tx.commit(); err != nil {
		tx.discard()
		return stats, err
	}
	if err := db.buildTextIndex(additions); err != nil {
		return stats, err
	}
	err = db.saveIndexes()
	if err != nil {
		return stats, err
	}
	stats.InsertTime = time.Since(insertStart)
	return stats, nil
}
