package db

import (
	"fmt"
	"github.com/google/btree"
	turtle "github.com/gtfierro/hod/goraptor"
)

type QueryResult struct {
	Rows  []ResultMap
	Links []LinkResultMap
	Count int
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

type ResultMap map[string]turtle.URI
type LinkResultMap map[turtle.URI]map[string]string

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
		before = before || item.Value[0] < row[idx].Value[0]
	}
	return before
}
