//go:generate msgp
package db

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gtfierro/hod/turtle"

	"github.com/gtfierro/btree"
)

var emptyResultMapList = []ResultMap{}

// QueryResult represents the result of a query execution
type QueryResult struct {
	selectVars []string
	// rows returned from the query
	Rows []ResultMap
	// number of results returned from the query
	Count int
	// amount of time elapsed for execution of the query
	Elapsed time.Duration
	// errors incurred during the execution of the query
	Errors []string
}

func (qr *QueryResult) fromRows(rows []*resultRow, vars []string, toMap bool) {
	qr.Count = len(rows)
	qr.selectVars = vars
	if toMap {
		for _, row := range rows {
			m := make(ResultMap)
			for idx, vname := range vars {
				m[vname] = row.row[idx]
			}
			qr.Rows = append(qr.Rows, m)
			finishResultRow(row)
		}
	}
}

// Dump writes a tabular representation of the query results to stdout
func (qr QueryResult) Dump() {
	if qr.Count == 0 {
		fmt.Println("No results")
		return
	}
	var dmp strings.Builder

	rowlens := make(map[string]int, len(qr.selectVars))

	for _, varname := range qr.selectVars {
		rowlens[varname] = len(varname)
	}

	for _, row := range qr.Rows {
		for varname, value := range row {
			if rowlens[varname] < len(value.String()) {
				rowlens[varname] = len(value.String())
			}
		}
	}

	totallen := 0
	for _, length := range rowlens {
		totallen += length + 2
	}

	fmt.Fprintf(&dmp, "+%s+\n", strings.Repeat("-", totallen+len(rowlens)-1))
	// header
	fmt.Fprintf(&dmp, "|")
	for _, varname := range qr.selectVars {
		fmt.Fprintf(&dmp, " %s%s|", varname, strings.Repeat(" ", rowlens[varname]-len(varname)+1))
	}
	fmt.Fprintf(&dmp, "\n")
	fmt.Fprintf(&dmp, "+%s+\n", strings.Repeat("-", totallen+len(rowlens)-1))

	for _, row := range qr.Rows {
		fmt.Fprintf(&dmp, "|")
		for _, varname := range qr.selectVars {
			valuelen := len(row[varname].String())
			fmt.Fprintf(&dmp, " %s%s |", row[varname], strings.Repeat(" ", rowlens[varname]-valuelen))
		}
		fmt.Fprintf(&dmp, "\n")
	}
	fmt.Fprintf(&dmp, "+%s+\n", strings.Repeat("-", totallen+len(rowlens)-1))
	fmt.Println(dmp.String())
	fmt.Println("Count:", qr.Count)
}

// DumpToCSV writes the query results to the provided io.Writer
func (qr QueryResult) DumpToCSV(usePrefixes bool, db *HodDB, w io.Writer) error {
	csvwriter := csv.NewWriter(w)
	if len(qr.Rows) > 0 {
		for _, row := range qr.Rows {
			var line = make([]string, len(qr.selectVars))
			for idx, varname := range qr.selectVars {
				//if usePrefixes {
				//	line[idx] = db.abbreviate(row[varname])
				//} else {
				line[idx] = row[varname].String()
				//}
			}
			if err := csvwriter.Write(line); err != nil {
				return err
			}
			csvwriter.Flush()
			if err := csvwriter.Error(); err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

// ResultMap maps variable names to their values
type ResultMap map[string]turtle.URI

type resultRow struct {
	row   []turtle.URI
	count int
}

func (rr resultRow) Less(than btree.Item, ctx interface{}) bool {
	row := than.(*resultRow)
	before := false
	for idx, item := range rr.row[:rr.count] {
		if before {
			return before
		}
		before = item.Value < row.row[idx].Value || item.Namespace < row.row[idx].Namespace
	}
	return before
}

var resultRowPool = sync.Pool{
	New: func() interface{} {
		return &resultRow{
			row:   make([]turtle.URI, 16),
			count: 0,
		}
	},
}

func getResultRow(num int) *resultRow {
	r := resultRowPool.Get().(*resultRow)
	r.count = num
	return r
}

func finishResultRow(r *resultRow) {
	r.count = 0
	resultRowPool.Put(r)
}

//func cleanResultRows(b *btree.BTree) {
//	i := b.DeleteMax()
//	for i != nil {
//		row := i.(*ResultRow)
//		finishResultRow(row)
//		i = b.DeleteMax()
//	}
//}
