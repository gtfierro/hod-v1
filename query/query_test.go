package lang

import (
	"reflect"
	"strings"
	"testing"

	turtle "github.com/gtfierro/hod/goraptor"
)

func TestQueryParse(t *testing.T) {
	for _, test := range []struct {
		str          string
		selectClause SelectClause
		whereClause  []Filter
	}{
		{
			"SELECT ?x WHERE { ?x rdf:type brick:Room . } ;",
			SelectClause{Variables: []turtle.URI{turtle.ParseURI("?x")}},
			[]Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type")}}, Object: turtle.ParseURI("brick:Room")}},
		},
		{
			"SELECT ?x ?y WHERE { ?x ?y brick:Room . } ;",
			SelectClause{Variables: []turtle.URI{turtle.URI{Value: "x"}, turtle.URI{Value: "y"}}},
			[]Filter{{Subject: turtle.URI{Value: "x"}, Path: []PathPattern{PathPattern{turtle.ParseURI("?y")}}, Object: turtle.ParseURI("brick:Room")}},
		},
		{
			"SELECT ?x ?y WHERE { ?y rdf:type rdf:type . ?x ?y brick:Room . } ;",
			SelectClause{Variables: []turtle.URI{turtle.ParseURI("?x"), turtle.ParseURI("?y")}},
			[]Filter{
				{Subject: turtle.URI{Value: "x"}, Path: []PathPattern{PathPattern{turtle.ParseURI("?y")}}, Object: turtle.ParseURI("brick:Room")},
				{Subject: turtle.URI{Value: "y"}, Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type")}}, Object: turtle.ParseURI("rdf:type")},
			},
		},
	} {
		r := strings.NewReader(test.str)
		q, e := Parse(r)
		if e != nil {
			t.Error(e)
			continue
		}
		if !reflect.DeepEqual(q.Select, test.selectClause) {
			t.Errorf("Query %s got select\n %s\n, expected %s", test.str, q.Select, test.selectClause)
		}
		if !reflect.DeepEqual(q.Where, test.whereClause) {
			t.Errorf("Query %s got where\n %s\n, expected %s", test.str, q.Where, test.whereClause)
		}
	}
}
