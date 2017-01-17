package db

import (
	"strings"
	"testing"

	"github.com/gtfierro/hod/config"
	"github.com/gtfierro/hod/query"
)

func BenchmarkExpandTuples(b *testing.B) {
	cfg, err := config.ReadConfig("testhodconfig.yaml")
	if err != nil {
		b.Error(err)
		return
	}
	cfg.DBPath = "test_databases/berkeleytestdb"
	db, err := NewDB(cfg)
	defer db.Close()
	if err != nil {
		b.Error(err)
		return
	}
	benchmarks := []struct {
		name  string
		query string
	}{
		{"SimpleSubjectVarTriple", "SELECT ?x WHERE { ?x rdf:type brick:Room . };"},
		{"LongerQuery1", "SELECT ?vav ?room WHERE { ?vav rdf:type brick:VAV . ?room rdf:type brick:Room . ?zone rdf:type brick:HVAC_Zone . ?vav bf:feeds+ ?zone . ?room bf:isPartOf ?zone . }; "},
		{"LooseQuery", "SELECT ?pred ?obj WHERE {   ?vav rdf:type brick:VAV .    ?vav ?pred ?obj .  } ;"},
	}

	b.ReportAllocs()
	for _, bm := range benchmarks {
		// setup the query
		q, e := query.Parse(strings.NewReader(bm.query))
		if e != nil {
			b.Error(e)
		}
		for idx, filter := range q.Where.Filters {
			q.Where.Filters[idx] = db.expandFilter(filter)
		}
		dg := db.sortQueryTerms(q)
		qp := db.formQueryPlan(dg, q)
		ctx := db.executeQueryPlan(qp)
		b.ResetTimer()
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ctx.expandTuples()
			}
		})
	}
}
