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
			[]Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")}},
		},
		{
			"SELECT ?x ?y WHERE { ?x ?y brick:Room . } ;",
			SelectClause{Variables: []turtle.URI{turtle.ParseURI("?x"), turtle.ParseURI("?y")}},
			[]Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("?y"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")}},
		},
		{
			"SELECT ?x ?y WHERE { ?y rdf:type rdf:type . ?x ?y brick:Room . } ;",
			SelectClause{Variables: []turtle.URI{turtle.ParseURI("?x"), turtle.ParseURI("?y")}},
			[]Filter{
				{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("?y"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")},
				{Subject: turtle.ParseURI("?y"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("rdf:type")},
			},
		},
		{
			"SELECT ?x WHERE { ?x rdf:type+ brick:Room . } ;",
			SelectClause{Variables: []turtle.URI{turtle.ParseURI("?x")}},
			[]Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_ONE_PLUS}}, Object: turtle.ParseURI("brick:Room")}},
		},
	} {
		r := strings.NewReader(test.str)
		q, e := Parse(r)
		if e != nil {
			t.Errorf("Error on query: %s", test.str, e)
			continue
		}
		if !reflect.DeepEqual(q.Select, test.selectClause) {
			t.Errorf("Query %s got select\n %v\nexpected\n %v", test.str, q.Select, test.selectClause)
			return
		}
		if !reflect.DeepEqual(q.Where, test.whereClause) {
			t.Errorf("Query %s got where\n %v\nexpected\n %v", test.str, q.Where, test.whereClause)
			return
		}
	}
}
