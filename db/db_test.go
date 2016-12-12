package db

import (
	"strings"
	"testing"

	turtle "github.com/gtfierro/hod/goraptor"
	query "github.com/gtfierro/hod/query"
)

func TestDBQuery(t *testing.T) {
	db, err := NewDB("../testdb")
	if err != nil {
		t.Error(err)
		return
	}
	for _, test := range []struct {
		query   string
		results []ResultMap
	}{
		{
			"SELECT ?x WHERE { ?x rdf:type brick:Room . };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#room_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds ?x .};",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds+ ?x .};",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
	} {
		q, e := query.Parse(strings.NewReader(test.query))
		if e != nil {
			t.Error(e)
			continue
		}
		results := db.RunQuery(q)
		if !compareResultMapList(test.results, results) {
			t.Errorf("Results for %s had\n %+v\nexpected\n %+v", test.query, results, test.results)
		}
	}
}
