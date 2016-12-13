package db

import (
	"github.com/gtfierro/hod/query"
	ll "log"
	"strings"
	"testing"
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
