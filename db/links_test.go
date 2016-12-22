package db

import (
	"encoding/json"
	"testing"

	"github.com/gtfierro/hod/goraptor"
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
				{URI: turtle.URI{"ex", "temp-sensor-1"}, Key: []byte("URI"), Value: []byte("ucberkeley/eecs/soda/sensors/etcetc/1")},
			}},
		},
		{
			`{"ex:temp-sensor-1": { "URI": "ucberkeley/eecs/soda/sensors/etcetc/1", "UUID": "abcdef" }}`,
			&LinkUpdates{Adding: []*Link{
				{URI: turtle.URI{"ex", "temp-sensor-1"}, Key: []byte("URI"), Value: []byte("ucberkeley/eecs/soda/sensors/etcetc/1")},
				{URI: turtle.URI{"ex", "temp-sensor-1"}, Key: []byte("UUID"), Value: []byte("abcdef")},
			}},
		},
		{
			`{"ex:temp-sensor-1": { "URI": "ucberkeley/eecs/soda/sensors/etcetc/1" }, "ex:temp-sensor-2": { "URI": "ucberkeley/eecs/soda/sensors/etcetc/2"}}`,
			&LinkUpdates{Adding: []*Link{
				{URI: turtle.URI{"ex", "temp-sensor-1"}, Key: []byte("URI"), Value: []byte("ucberkeley/eecs/soda/sensors/etcetc/1")},
				{URI: turtle.URI{"ex", "temp-sensor-2"}, Key: []byte("URI"), Value: []byte("ucberkeley/eecs/soda/sensors/etcetc/2")},
			}},
		},
		{
			`{"ex:temp-sensor-1": { "URI": "ucberkeley/eecs/soda/sensors/etcetc/1", "UUID": "" }}`,
			&LinkUpdates{
				Adding: []*Link{
					{URI: turtle.URI{"ex", "temp-sensor-1"}, Key: []byte("URI"), Value: []byte("ucberkeley/eecs/soda/sensors/etcetc/1")},
				},
				Removing: []*Link{
					{URI: turtle.URI{"ex", "temp-sensor-1"}, Key: []byte("UUID")},
				},
			},
		},
		{
			`{"ex:temp-sensor-1": {}}`,
			&LinkUpdates{
				Removing: []*Link{
					{URI: turtle.URI{"ex", "temp-sensor-1"}},
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
