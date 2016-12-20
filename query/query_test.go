package query

import (
	"reflect"
	"strings"
	"testing"

	turtle "github.com/gtfierro/hod/goraptor"
)

func TestOrClauseFlatten(t *testing.T) {
	for _, test := range []struct {
		orclause  OrClause
		flattened [][]Filter
	}{
		{
			OrClause{
				LeftTerms:  []Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")}},
				RightTerms: []Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Floor")}},
			},
			[][]Filter{
				{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")}},
				{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Floor")}},
			},
		},
	} {
		flattened := test.orclause.Flatten()
		if len(flattened) != len(test.flattened) {
			t.Errorf("Flatten() failed. Wanted\n%v\nbut got\n%v", test.flattened, flattened)
		}
		for idx := 0; idx < len(flattened); idx++ {
			if !compareFilterSliceAsSet(flattened[idx], test.flattened[idx]) {
				t.Errorf("Flatten() failed. Wanted\n%v\nbut got\n%v", test.flattened, flattened)
			}
		}

	}
}

func TestQueryParse(t *testing.T) {
	for _, test := range []struct {
		str          string
		selectClause SelectClause
		whereClause  WhereClause
	}{
		{
			"SELECT ?x WHERE { ?x rdf:type brick:Room . } ;",
			SelectClause{Variables: []SelectVar{{Var: turtle.ParseURI("?x")}}},
			WhereClause{
				[]Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")}},
				[]OrClause{},
			},
		},
		{
			"SELECT ?x ?y WHERE { ?x ?y brick:Room . } ;",
			SelectClause{Variables: []SelectVar{{Var: turtle.ParseURI("?x")}, {Var: turtle.ParseURI("?y")}}},
			WhereClause{
				[]Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("?y"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")}},
				[]OrClause{},
			},
		},
		{
			"SELECT ?x ?y WHERE { ?y rdf:type rdf:type . ?x ?y brick:Room . } ;",
			SelectClause{Variables: []SelectVar{{Var: turtle.ParseURI("?x")}, {Var: turtle.ParseURI("?y")}}},
			WhereClause{
				[]Filter{
					{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("?y"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")},
					{Subject: turtle.ParseURI("?y"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("rdf:type")},
				},
				[]OrClause{},
			},
		},
		{
			"SELECT ?x WHERE { ?x rdf:type+ brick:Room . } ;",
			SelectClause{Variables: []SelectVar{{Var: turtle.ParseURI("?x")}}},
			WhereClause{
				[]Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_ONE_PLUS}}, Object: turtle.ParseURI("brick:Room")}},
				[]OrClause{},
			},
		},
		{
			"SELECT ?x ?y WHERE { ?y rdf:type|rdfs:subClassOf ?x .} ;",
			SelectClause{Variables: []SelectVar{{Var: turtle.ParseURI("?x")}, {Var: turtle.ParseURI("?y")}}},
			WhereClause{
				[]Filter{},
				[]OrClause{
					{
						RightTerms: []Filter{{Subject: turtle.ParseURI("?y"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("?x")}},
						LeftOr: []OrClause{
							{
								RightTerms: []Filter{{Subject: turtle.ParseURI("?y"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdfs:subClassOf"), PATTERN_SINGLE}}, Object: turtle.ParseURI("?x")}},
							},
						},
					},
				},
			},
		},
		{
			"SELECT ?x ?y WHERE { ?y rdf:type|rdfs:subClassOf|rdf:isa ?x .} ;",
			SelectClause{Variables: []SelectVar{{Var: turtle.ParseURI("?x")}, {Var: turtle.ParseURI("?y")}}},
			WhereClause{
				[]Filter{},
				[]OrClause{
					{
						RightTerms: []Filter{{Subject: turtle.ParseURI("?y"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("?x")}},
						LeftOr: []OrClause{
							{
								RightTerms: []Filter{{Subject: turtle.ParseURI("?y"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdfs:subClassOf"), PATTERN_SINGLE}}, Object: turtle.ParseURI("?x")}},
								LeftOr: []OrClause{
									{
										RightTerms: []Filter{{Subject: turtle.ParseURI("?y"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:isa"), PATTERN_SINGLE}}, Object: turtle.ParseURI("?x")}},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			"SELECT ?x WHERE { {?x rdf:type brick:Room . OR ?x rdf:type brick:Floor .} };",
			SelectClause{Variables: []SelectVar{{Var: turtle.ParseURI("?x")}}},
			WhereClause{
				[]Filter{},
				[]OrClause{{
					LeftTerms:  []Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Floor")}},
					RightTerms: []Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")}},
				}},
			},
		},
		{
			"SELECT ?x WHERE { ?y rdf:type brick:Building . { { ?x rdf:type brick:Room . ?x bf:isPartOf+ ?y .} OR { ?x rdf:type brick:Floor . ?x bf:isPartOf+ ?y .} } };",
			SelectClause{Variables: []SelectVar{{Var: turtle.ParseURI("?x")}}},
			WhereClause{
				[]Filter{{Subject: turtle.ParseURI("?y"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Building")}},
				[]OrClause{{
					LeftTerms: []Filter{
						{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Floor")},
						{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("bf:isPartOf"), PATTERN_ONE_PLUS}}, Object: turtle.ParseURI("?y")},
					},
					RightTerms: []Filter{
						{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")},
						{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("bf:isPartOf"), PATTERN_ONE_PLUS}}, Object: turtle.ParseURI("?y")},
					},
				}},
			},
		},
		{
			"SELECT ?x[link1] WHERE { ?x rdf:type brick:Room . } ;",
			SelectClause{Variables: []SelectVar{{Var: turtle.ParseURI("?x"), Links: []Link{{Name: "link1"}}}}},
			WhereClause{
				[]Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")}},
				[]OrClause{},
			},
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
		if !compareFilterSliceAsSet(q.Where.Filters, test.whereClause.Filters) {
			t.Errorf("Query %s got where Filters\n %v\nexpected\n %v", test.str, q.Where.Filters, test.whereClause.Filters)
			return
		}
		if !compareOrClauseLists(q.Where.Ors, test.whereClause.Ors) {
			t.Errorf("Query %s got where Ors\n %+v\nexpected\n %+v", test.str, q.Where.Ors, test.whereClause.Ors)
			return
		}
	}
}

func BenchmarkQueryParseShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := strings.NewReader("SELECT ?x WHERE { ?x rdf:type brick:Room . } ;")
		Parse(r)
	}
}

func BenchmarkQueryParseLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r := strings.NewReader("SELECT ?x WHERE { ?y rdf:type brick:Building . { { ?x rdf:type brick:Room . ?x bf:isPartOf+ ?y .} OR { ?x rdf:type brick:Floor . ?x bf:isPartOf+ ?y .} } };")
		Parse(r)
	}
}

func BenchmarkOrClauseFlattenShort(b *testing.B) {
	oc := OrClause{
		LeftTerms:  []Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Room")}},
		RightTerms: []Filter{{Subject: turtle.ParseURI("?x"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("brick:Floor")}},
	}
	for i := 0; i < b.N; i++ {
		oc.Flatten()
	}
}
func BenchmarkOrClauseFlattenLong(b *testing.B) {
	oc := OrClause{
		RightTerms: []Filter{{Subject: turtle.ParseURI("?y"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:type"), PATTERN_SINGLE}}, Object: turtle.ParseURI("?x")}},
		LeftOr: []OrClause{
			{
				RightTerms: []Filter{{Subject: turtle.ParseURI("?y"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdfs:subClassOf"), PATTERN_SINGLE}}, Object: turtle.ParseURI("?x")}},
				LeftOr: []OrClause{
					{
						RightTerms: []Filter{{Subject: turtle.ParseURI("?y"), Path: []PathPattern{PathPattern{turtle.ParseURI("rdf:isa"), PATTERN_SINGLE}}, Object: turtle.ParseURI("?x")}},
					},
				},
			},
		},
	}
	for i := 0; i < b.N; i++ {
		oc.Flatten()
	}
}
