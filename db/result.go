package db

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/gtfierro/hod/query"

	"github.com/google/btree"
)

var emptyResultMapList = []ResultMap{}
var emptyLinkResultmapList = []LinkResultMap{}

type QueryResult struct {
	selectVars []query.SelectVar
	Rows       []ResultMap
	Links      []LinkResultMap
	Count      int
}

func newQueryResult() QueryResult {
	return QueryResult{
		Rows:  emptyResultMapList,
		Links: emptyLinkResultmapList,
	}
}

func (qr QueryResult) Dump() {
	if len(qr.Rows) > 0 {
		for _, row := range qr.Rows {
			fmt.Println(row)
		}
		return
	}
	if len(qr.Links) > 0 {
		for _, link := range qr.Links {
			fmt.Println(link)
		}
		return
	}
	fmt.Println(qr.Count)
}

func (qr QueryResult) DumpToCSV(usePrefixes bool, db *DB, w io.Writer) error {
	csvwriter := csv.NewWriter(w)
	if len(qr.Rows) > 0 {
		for _, row := range qr.Rows {
			var line = make([]string, len(qr.selectVars))
			for idx, varname := range qr.selectVars {
				if usePrefixes {
					line[idx] = db.Abbreviate(row[varname.Var.Value])
				} else {
					line[idx] = row[varname.Var.Value].String()
				}
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
	if len(qr.Links) > 0 {
		for _, link := range qr.Links {
			fmt.Println(link)
		}
		return nil
	}
	return nil
}

type ResultMap map[string]turtle.URI
type LinkResultMap map[turtle.URI]map[string]string

func (m LinkResultMap) MarshalJSON() ([]byte, error) {
	var n = make(map[string]map[string]string)
	for k, v := range m {
		n[k.String()] = v
	}
	return json.Marshal(n)
}

type ResultRow []turtle.URI

func (rr ResultRow) Less(than btree.Item) bool {
	row := than.(ResultRow)
	if len(rr) < len(row) {
		return true
	} else if len(row) < len(rr) {
		return false
	}
	before := false
	for idx, item := range rr {
		before = before || item.Value < row[idx].Value || item.Namespace < row[idx].Namespace
	}
	return before
}
