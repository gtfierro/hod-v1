package db

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/gtfierro/hod/config"
	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/gtfierro/hod/query"
)

//func TestLinkdbkey(t *testing.T) {
//	var fetchkey [64]byte
//	for _, test := range []struct {
//		entity [4]byte
//		key    []byte
//		result [64]byte
//	}{
//		{
//			[4]byte{1, 2, 3, 4},
//			[]byte{1, 1, 1, 1},
//			[64]byte{1, 2, 3, 4, 1, 1, 1, 1},
//		},
//		{
//			[4]byte{1, 2, 3, 4},
//			[]byte{1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4},
//			[64]byte{1, 2, 3, 4, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4},
//		},
//	} {
//		getlinkdbkey(test.entity, test.key, &fetchkey)
//		if fetchkey != test.result {
//			t.Errorf("linkdbkey failed. Got\n%+v\nbut wanted\n%+v\n", fetchkey, test.result)
//		}
//	}
//}

func TestLinkUpdateUnmarshal(t *testing.T) {
	for _, test := range []struct {
		jsonString string
		result     *LinkUpdates
	}{
		{
			`{"ex:temp-sensor-1": { "URI": "ucberkeley/eecs/soda/sensors/etcetc/1" }}`,
			&LinkUpdates{Adding: []*Link{
				{URI: turtle.URI{Namespace: "ex", Value: "temp-sensor-1"}, Key: []byte("URI"), Value: []byte("ucberkeley/eecs/soda/sensors/etcetc/1")},
			}},
		},
		{
			`{"ex:temp-sensor-1": { "URI": "ucberkeley/eecs/soda/sensors/etcetc/1", "UUID": "abcdef" }}`,
			&LinkUpdates{Adding: []*Link{
				{URI: turtle.URI{Namespace: "ex", Value: "temp-sensor-1"}, Key: []byte("URI"), Value: []byte("ucberkeley/eecs/soda/sensors/etcetc/1")},
				{URI: turtle.URI{Namespace: "ex", Value: "temp-sensor-1"}, Key: []byte("UUID"), Value: []byte("abcdef")},
			}},
		},
		{
			`{"ex:temp-sensor-1": { "URI": "ucberkeley/eecs/soda/sensors/etcetc/1" }, "ex:temp-sensor-2": { "URI": "ucberkeley/eecs/soda/sensors/etcetc/2"}}`,
			&LinkUpdates{Adding: []*Link{
				{URI: turtle.URI{Namespace: "ex", Value: "temp-sensor-1"}, Key: []byte("URI"), Value: []byte("ucberkeley/eecs/soda/sensors/etcetc/1")},
				{URI: turtle.URI{Namespace: "ex", Value: "temp-sensor-2"}, Key: []byte("URI"), Value: []byte("ucberkeley/eecs/soda/sensors/etcetc/2")},
			}},
		},
		{
			`{"ex:temp-sensor-1": { "URI": "ucberkeley/eecs/soda/sensors/etcetc/1", "UUID": "" }}`,
			&LinkUpdates{
				Adding: []*Link{
					{URI: turtle.URI{Namespace: "ex", Value: "temp-sensor-1"}, Key: []byte("URI"), Value: []byte("ucberkeley/eecs/soda/sensors/etcetc/1")},
				},
				Removing: []*Link{
					{URI: turtle.URI{Namespace: "ex", Value: "temp-sensor-1"}, Key: []byte("UUID")},
				},
			},
		},
		{
			`{"ex:temp-sensor-1": {}}`,
			&LinkUpdates{
				Removing: []*Link{
					{URI: turtle.URI{Namespace: "ex", Value: "temp-sensor-1"}},
				},
			},
		},
	} {
		var updates = new(LinkUpdates)
		err := json.Unmarshal([]byte(test.jsonString), updates)
		if err != nil {
			t.Error(err)
			continue
		}
		if !compareLinkUpdates(test.result, updates) {
			t.Errorf("Expected\n%+v\nbut got\n%+v", test.result, updates)
		}
	}
}

func TestLinkQuery(t *testing.T) {
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
		results []LinkResultMap
	}{
		{
			"SELECT ?x[UUID] WHERE { ?x rdf:type brick:VAV . };",
			[]LinkResultMap{
				{
					turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1"): {
						"UUID": "427b8f7c-dc3a-11e6-8b12-1002b58053c7",
					},
				},
			},
		},
		{
			"SELECT ?f[Coords] WHERE { ?x rdf:type brick:VAV . ?x bf:feeds ?f . };",
			[]LinkResultMap{
				{
					turtle.ParseURI("http://buildsys.org/ontologies/building_example#hvaczone_1"): {
						"Coords": "[2, 3]",
					},
				},
			},
		},
		{
			"SELECT ?x[*] WHERE { ?x rdf:type brick:VAV . };",
			[]LinkResultMap{
				{
					turtle.ParseURI("http://buildsys.org/ontologies/building_example#vav_1"): {
						"UUID": "427b8f7c-dc3a-11e6-8b12-1002b58053c7",
					},
				},
			},
		},
	} {
		q, e := query.Parse(strings.NewReader(test.query))
		if e != nil {
			t.Error(test.query, e)
			continue
		}
		result := db.RunQuery(q)
		if !reflect.DeepEqual(result.Links, test.results) {
			t.Errorf("Results for %s had\n %+v\nexpected\n %+v", test.query, result.Links, test.results)
			return
		}
	}
}
