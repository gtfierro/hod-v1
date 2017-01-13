package db

import (
	"strings"
	"testing"

	"github.com/gtfierro/hod/config"
	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/gtfierro/hod/query"
)

func TestDBQuery(t *testing.T) {
	cfg, err := config.ReadConfig("testhodconfig.yaml")
	if err != nil {
		t.Error(err)
		return
	}
	cfg.DBPath = "test_databases/testdb"
	db, err := NewDB(cfg)
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
			"SELECT ?x WHERE { ?x rdf:type <http://buildsys.org/ontologies/Brick#Room> . };",
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
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy+ ?ahu .};",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds ?x .};",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds+ ?x .};",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds? ?x .};",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy/bf:isFedBy? ?ahu .};",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds* ?x .};",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#ahu_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy* ?ahu .};",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#ahu_1")}},
		},
		{
			"SELECT ?vav ?room WHERE { ?vav rdf:type brick:VAV . ?room rdf:type brick:Room . ?zone rdf:type brick:HVAC_Zone . ?vav bf:feeds+ ?zone . ?room bf:isPartOf ?zone . }; ",
			[]ResultMap{{"?room": turtle.ParseURI("http://buildsys.org/ontologies/building_example#room_1"), "?vav": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
		{
			"SELECT ?sensor WHERE { ?sensor rdf:type/rdfs:subClassOf* brick:Zone_Temperature_Sensor . };",
			[]ResultMap{{"?sensor": turtle.ParseURI("http://buildsys.org/ontologies/building_example#ztemp_1")}},
		},
		{
			"SELECT ?s ?p WHERE { ?s ?p brick:Zone_Temperature_Sensor . ?s rdfs:subClassOf brick:Zone_Temperature_Sensor . };",
			[]ResultMap{
				{"?s": turtle.ParseURI("http://buildsys.org/ontologies/Brick#Average_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("http://buildsys.org/ontologies/Brick#Coldest_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("http://buildsys.org/ontologies/Brick#Highest_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("http://buildsys.org/ontologies/Brick#Lowest_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("http://buildsys.org/ontologies/Brick#Warmest_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
			},
		},
	} {
		q, e := query.Parse(strings.NewReader(test.query))
		if e != nil {
			t.Error(test.query, e)
			continue
		}
		result := db.RunQuery(q)
		if !compareResultMapList(test.results, result.Rows) {
			t.Errorf("Results for %s had\n %+v\nexpected\n %+v", test.query, result.Rows, test.results)
			return
		}
	}
}

func TestDBQueryBerkeley(t *testing.T) {
	cfg, err := config.ReadConfig("testhodconfig.yaml")
	if err != nil {
		t.Error(err)
		return
	}
	cfg.DBPath = "test_databases/berkeleytestdb"
	db, err := NewDB(cfg)
	if err != nil {
		t.Error(err)
		return
	}
	for _, test := range []struct {
		query       string
		resultCount int
	}{
		{
			"COUNT ?x WHERE { ?x rdf:type brick:Room . };",
			243,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds ?x .};",
			240,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds+ ?x .};",
			480,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy+ ?ahu .};",
			480,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds ?x .};",
			240,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds+ ?x .};",
			240,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds? ?x .};",
			480,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy/bf:isFedBy? ?ahu .};",
			480,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds* ?x .};",
			485,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy* ?ahu .};",
			485,
		},
		{
			"COUNT ?vav ?room WHERE { ?vav rdf:type brick:VAV . ?room rdf:type brick:Room . ?zone rdf:type brick:HVAC_Zone . ?vav bf:feeds+ ?zone . ?room bf:isPartOf ?zone . }; ",
			243,
		},
		{
			"COUNT ?sensor WHERE { ?sensor rdf:type/rdfs:subClassOf* brick:Zone_Temperature_Sensor . };",
			232,
		},
		{
			"COUNT ?sensor ?room WHERE { ?sensor rdf:type/rdfs:subClassOf* brick:Zone_Temperature_Sensor . ?vav rdf:type brick:VAV . ?zone rdf:type brick:HVAC_Zone . ?room rdf:type brick:Room . ?vav bf:feeds+ ?zone . ?zone bf:hasPart ?room . { ?sensor bf:isPointOf ?vav . OR ?sensor bf:isPointOf ?room .} };",
			232,
		},
	} {
		q, e := query.Parse(strings.NewReader(test.query))
		if e != nil {
			t.Error(test.query, e)
			continue
		}
		result := db.RunQuery(q)
		if result.Count != test.resultCount {
			t.Errorf("Results for %s had %d expected %d", test.query, result.Count, test.resultCount)
			return
		}
	}
}

func BenchmarkQueryPerformance1(b *testing.B) {
	cfg, err := config.ReadConfig("testhodconfig.yaml")
	if err != nil {
		b.Error(err)
		return
	}
	cfg.DBPath = "test_databases/berkeleytestdb"
	db, err := NewDB(cfg)
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

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				q, e := query.Parse(strings.NewReader(bm.query))
				if e != nil {
					b.Error(e)
				}
				db.RunQuery(q)
			}
		})
	}
}
