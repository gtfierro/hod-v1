package db

import (
	"testing"
	"time"

	"github.com/gtfierro/hod/config"
	turtle "github.com/gtfierro/hod/goraptor"
	query "github.com/gtfierro/hod/lang"
)

func TestDBQuery(t *testing.T) {
	cfg, err := config.ReadConfig("testhodconfig.yaml")
	if err != nil {
		t.Error(err)
		return
	}
	cfg.DBPath = "test_databases/testdb"
	db, err := NewDB(cfg)
	defer db.Close()
	if err != nil {
		t.Error(err)
		return
	}
	for _, test := range []struct {
		query   string
		results []ResultMap
	}{
		{
			"SELECT ?x WHERE { ?x rdf:type brick:Room };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#room_1")}},
		},
		{
			"SELECT ?x WHERE { bldg:room_1 rdf:type ?x };",
			[]ResultMap{{"?x": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#Room")}},
		},
		{
			"SELECT ?x WHERE { bldg:room_1 ?x brick:Room };",
			[]ResultMap{{"?x": turtle.ParseURI("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")}},
		},
		{
			"SELECT ?x ?y WHERE { ?x bf:feeds ?y };",
			[]ResultMap{
				{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1"), "?y": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")},
				{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#ahu_1"), "?y": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")},
			},
		},
		{
			"SELECT ?x ?y WHERE { bldg:room_1 ?x ?y };",
			[]ResultMap{
				{"?x": turtle.ParseURI("https://brickschema.org/schema/1.0.1/BrickFrame#isPartOf"), "?y": turtle.ParseURI("http://buildsys.org/ontologies/building_example#floor_1")},
				{"?x": turtle.ParseURI("https://brickschema.org/schema/1.0.1/BrickFrame#isPartOf"), "?y": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")},
				{"?x": turtle.ParseURI("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), "?y": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#Room")},
				{"?x": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#label"), "?y": turtle.URI{Value: "Room 1"}},
			},
		},
		{
			"SELECT ?x ?y WHERE { ?r rdf:type brick:Room . ?r ?x ?y };",
			[]ResultMap{
				{"?x": turtle.ParseURI("https://brickschema.org/schema/1.0.1/BrickFrame#isPartOf"), "?y": turtle.ParseURI("http://buildsys.org/ontologies/building_example#floor_1")},
				{"?x": turtle.ParseURI("https://brickschema.org/schema/1.0.1/BrickFrame#isPartOf"), "?y": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")},
				{"?x": turtle.ParseURI("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"), "?y": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#Room")},
				{"?x": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#label"), "?y": turtle.URI{Value: "Room 1"}},
			},
		},
		{
			"SELECT ?x ?y WHERE { ?r rdf:type brick:Room . ?x ?y ?r };",
			[]ResultMap{
				{"?y": turtle.ParseURI("https://brickschema.org/schema/1.0.1/BrickFrame#hasPart"), "?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#floor_1")},
				{"?y": turtle.ParseURI("https://brickschema.org/schema/1.0.1/BrickFrame#hasPart"), "?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")},
			},
		},
		//		{
		//			"SELECT ?x ?y WHERE { bldg:room_1 ?p bldg:floor_1 . ?x ?p ?y };",
		//			[]ResultMap{
		//				{"?y": turtle.ParseURI("http://buildsys.org/ontologies/BrickFrame#hasPart"), "?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#floor_1")},
		//				{"?y": turtle.ParseURI("http://buildsys.org/ontologies/BrickFrame#hasPart"), "?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")},
		//			},
		//		},
		{
			"SELECT ?x WHERE { ?x rdf:type <https://brickschema.org/schema/1.0.1/Brick#Room> };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#room_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds ?x };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds+ ?x };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy+ ?ahu };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds ?x };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds+ ?x };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds? ?x };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy/bf:isFedBy? ?ahu };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds* ?x };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#ahu_1")}},
		},
		{
			"SELECT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy* ?ahu };",
			[]ResultMap{{"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}, {"?x": turtle.ParseURI("http://buildsys.org/ontologies/building_example#ahu_1")}},
		},
		{
			"SELECT ?vav ?room WHERE { ?vav rdf:type brick:VAV . ?room rdf:type brick:Room . ?zone rdf:type brick:HVAC_Zone . ?vav bf:feeds+ ?zone . ?room bf:isPartOf ?zone }; ",
			[]ResultMap{{"?room": turtle.ParseURI("http://buildsys.org/ontologies/building_example#room_1"), "?vav": turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1")}},
		},
		{
			"SELECT ?sensor WHERE { ?sensor rdf:type/rdfs:subClassOf* brick:Zone_Temperature_Sensor };",
			[]ResultMap{{"?sensor": turtle.ParseURI("http://buildsys.org/ontologies/building_example#ztemp_1")}},
		},
		{
			"SELECT ?s ?p WHERE { ?s ?p brick:Zone_Temperature_Sensor . ?s rdfs:subClassOf brick:Zone_Temperature_Sensor };",
			[]ResultMap{
				{"?s": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#Average_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#Coldest_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#Highest_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#Lowest_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#Warmest_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#VAV_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#AHU_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#FCU_Zone_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
				{"?s": turtle.ParseURI("https://brickschema.org/schema/1.0.1/Brick#Zone_Air_Temperature_Sensor"), "?p": turtle.ParseURI("http://www.w3.org/2000/01/rdf-schema#subClassOf")},
			},
		},
	} {
		time.Sleep(100 * time.Millisecond)
		q, e := query.Parse(test.query)
		if e != nil {
			t.Error(test.query, e)
			continue
		}
		result, err := db.RunQuery(q)
		if err != nil {
			t.Error(err)
			return
		}
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
	defer db.Close()
	if err != nil {
		t.Error(err)
		return
	}
	for _, test := range []struct {
		query       string
		resultCount int
	}{
		{
			"COUNT ?x WHERE { ?x rdf:type brick:Room };",
			243,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds ?x };",
			240,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds+ ?x };",
			480,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy+ ?ahu };",
			480,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds ?x };",
			240,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds+ ?x };",
			240,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds/bf:feeds? ?x };",
			480,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy/bf:isFedBy? ?ahu };",
			480,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds* ?x };",
			485,
		},
		{
			"COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy* ?ahu };",
			485,
		},
		{
			"COUNT ?vav ?room WHERE { ?vav rdf:type brick:VAV . ?room rdf:type brick:Room . ?zone rdf:type brick:HVAC_Zone . ?vav bf:feeds+ ?zone . ?room bf:isPartOf ?zone }; ",
			243,
		},
		{
			"COUNT ?sensor WHERE { ?sensor rdf:type/rdfs:subClassOf* brick:Zone_Temperature_Sensor };",
			232,
		},
		{
			"COUNT ?sensor ?room WHERE { ?sensor rdf:type/rdfs:subClassOf* brick:Zone_Temperature_Sensor . ?room rdf:type brick:Room . ?vav rdf:type brick:VAV . ?zone rdf:type brick:HVAC_Zone . ?vav bf:feeds+ ?zone . ?zone bf:hasPart ?room . ?sensor bf:isPointOf ?vav };",
			232,
		},
		{
			"COUNT ?sensor ?room WHERE { ?sensor rdf:type/rdfs:subClassOf* brick:Zone_Temperature_Sensor . ?vav rdf:type brick:VAV . ?zone rdf:type brick:HVAC_Zone . ?room rdf:type brick:Room . ?vav bf:feeds+ ?zone . ?zone bf:hasPart ?room  { ?sensor bf:isPointOf ?vav } UNION { ?sensor bf:isPointOf ?room } };",
			232,
		},
		{
			"COUNT ?sensor ?room WHERE { ?sensor rdf:type/rdfs:subClassOf* brick:Zone_Temperature_Sensor . ?room rdf:type brick:Room . ?vav rdf:type brick:VAV . ?zone rdf:type brick:HVAC_Zone . ?vav bf:feeds+ ?zone . ?zone bf:hasPart ?room . ?sensor bf:isPointOf ?room };",
			0,
		},
		{
			"COUNT ?vav ?x ?y WHERE { ?vav rdf:type brick:VAV . ?vav bf:hasPoint ?x . ?vav bf:isFedBy ?y };",
			823,
		},
		{
			"COUNT ?ahu WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds soda_hall:vav_C711 };",
			1,
		},
		{
			"COUNT ?ahu WHERE { ?ahu bf:feeds soda_hall:vav_C711 . ?ahu rdf:type brick:AHU };",
			1,
		},
		{
			"COUNT ?vav ?x ?y ?z WHERE { ?vav rdf:type brick:VAV . ?vav bf:feeds+ ?x . ?vav bf:isFedBy+ ?y . ?vav bf:hasPoint+ ?z };",
			823,
		},
	} {
		time.Sleep(100 * time.Millisecond)
		q, e := query.Parse(test.query)
		if e != nil {
			t.Error(test.query, e)
			continue
		}
		result, err := db.RunQuery(q)
		if err != nil {
			t.Error(err)
			return
		}
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
	defer db.Close()
	if err != nil {
		b.Error(err)
		return
	}
	benchmarks := []struct {
		name  string
		query string
	}{
		{"SimpleSubjectVarTriple", "SELECT ?x WHERE { ?x rdf:type brick:Room };"},
		{"LongerQuery1", "SELECT ?vav ?room WHERE { ?vav rdf:type brick:VAV . ?room rdf:type brick:Room . ?zone rdf:type brick:HVAC_Zone . ?vav bf:feeds+ ?zone . ?room bf:isPartOf ?zone }; "},
		{"LooseQuery", "SELECT ?pred ?obj WHERE { ?vav rdf:type brick:VAV .  ?vav ?pred ?obj }; "},
		{"LocQuery", " SELECT ?sensor ?room WHERE { ?sensor rdf:type/rdfs:subClassOf* brick:Zone_Temperature_Sensor . ?room rdf:type brick:Room . ?vav rdf:type brick:VAV . ?zone rdf:type brick:HVAC_Zone . ?vav bf:feeds+ ?zone . ?zone bf:hasPart ?room { ?sensor bf:isPointOf ?vav } UNION { ?sensor bf:isPointOf ?room } };"},
		{"RoomEnum", "COUNT ?x WHERE { ?x rdf:type brick:Room };"},
		{"AHUFeed0", "COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds ?x };"},
		{"AHUFeed1", "COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds+ ?x };"},
		{"AHUFeed2", "COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?ahu bf:feeds* ?x };"},
		{"AHUFeed1Reverse", "COUNT ?x WHERE { ?ahu rdf:type brick:AHU . ?x bf:isFedBy+ ?ahu };"},
		{"SensorSubclass", "COUNT ?sensor WHERE { ?sensor rdf:type/rdfs:subClassOf* brick:Zone_Temperature_Sensor };"},
		{"VAVExplore", "COUNT ?vav ?x ?y WHERE { ?vav rdf:type brick:VAV . ?vav bf:hasPoint ?x . ?vav bf:isFedBy ?y };"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				q, e := query.Parse(bm.query)
				if e != nil {
					b.Error(e)
					continue
				}
				db.RunQuery(q)
			}
		})
	}
}
