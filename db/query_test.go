package db

import (
	ll "log"
	"strings"
	"testing"

	"github.com/gtfierro/hod/config"
	"github.com/gtfierro/hod/query"
)

func TestQueryPlan(t *testing.T) {
	for _, test := range []struct {
		query string
	}{
		{
			"SELECT ?x WHERE { ?x rdf:type brick:Room . } ;",
		},
		{
			"SELECT ?x ?y WHERE { ?x ?y brick:Room . } ;",
		},
		{
			"SELECT ?x ?y WHERE { ?y rdf:type rdf:type . ?x ?y brick:Room . } ;",
		},
		{
			"SELECT ?a WHERE { ?a bf:feeds ?b . ?b bf:feeds ?c . ?c bf:feeds ?d . ?d bf:feeds ?e . ?e bf:feeds ?loc . ?loc bf:hasPoint brick:Power_Meter . };",
		},
		{
			"SELECT ?a WHERE { ?a bf:feeds ?b . ?b bf:feeds ?c . ?c bf:feeds ?d . ?d bf:feeds ?e . ?e bf:feeds ?loc . ?loc bf:hasPoint brick:Power_Meter . };",
		},
		{
			"SELECT ?meter WHERE { ?meter rdf:type brick:Power_Meter . ?room rdf:type brick:Room . ?meter bf:isPointOf ?equipment . ?equipment rdf:type ?class . ?class rdfs:subClassOf+ brick:Heating_Ventilation_Air_Conditioning_System . ?zone rdf:type/rdfs:subClassOf* brick:HVAC_Zone . ?equipment bf:feeds+ ?zone . ?zone bf:hasPart ?room . } ;",
		},
	} {
		ll.Println(test.query)
		q, e := query.Parse(strings.NewReader(test.query))
		if e != nil {
			t.Error(e)
			continue
		}
		//db := &DB{}
		makeDependencyGraph(q)
	}
}

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
		dg := db.formDependencyGraph(q)
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
