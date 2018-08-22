package db

import (
	"fmt"
	"github.com/gtfierro/hod/config"
	"github.com/gtfierro/hod/turtle"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func nameToClass(name string) turtle.URI {
	return turtle.ParseURI(fmt.Sprintf("https://brickschema.org/schema/1.0.3/Brick#%s", name))
}

// spec is of form "Brick class" -> number o those entities
func generateTriples(spec map[string]int) (triples []turtle.Triple) {
	isType := turtle.ParseURI("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")
	for class, num := range spec {
		obj := nameToClass(class)
		for i := 0; i < num; i++ {
			sub := turtle.ParseURI(fmt.Sprintf("http://buildsys.org/ontologies/generated#%s_%d", class, i))
			triples = append(triples, turtle.Triple{sub, isType, obj})
		}
	}
	return triples
}

// writes the triples to the file
func triplesToFile(name string, triples []turtle.Triple) (path string, err error) {
	fileContents := `
@prefix bf: <https://brickschema.org/schema/1.0.3/BrickFrame#> .
@prefix bldg: <http://buildsys.org/ontologies/building_example#> .
@prefix brick: <https://brickschema.org/schema/1.0.3/Brick#> .
@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
`
	for _, t := range triples {
		fileContents += fmt.Sprintf("<%s> <%s> <%s> .\n", t.Subject, t.Predicate, t.Object)
	}

	f, err := ioutil.TempFile(".", "gentriple")
	if err != nil {
		return "", err
	}
	_, err = f.Write([]byte(fileContents))
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func TestGenerateTriples(t *testing.T) {
	require := require.New(t)

	triples := generateTriples(map[string]int{
		"Room": 1,
		"AHU":  3,
	})
	require.Equal(4, len(triples))
	path, err := triplesToFile("gentrip1", triples)
	require.NoError(err)
	defer os.Remove(path)

	cfgStr := fmt.Sprintf(`Buildings:
    %s: %s
Ontologies:
    - testbuildings/BrickFrame.ttl
    - testbuildings/Brick.ttl
StorageEngine: memory
ReloadOntologies: false
DisableQueryCache: true
ShowNamespaces: false
ShowQueryPlan: false
ShowDependencyGraph: false
ShowQueryPlanLatencies: false
ShowOperationLatencies: false
ShowQueryLatencies: false
LogLevel: Critical
EnableBOSSWAVE: false
EnableHTTP: false`, "gentrip1", path)
	cfg, err := config.ReadConfigFromString(cfgStr)
	require.NoError(err, cfgStr)
	db, err := NewHodDB(cfg)
	defer db.Close()
	require.NoError(err)

	result, err := db.RunQueryString("SELECT ?r WHERE { ?r rdf:type brick:Room };")
	require.NoError(err)
	require.Equal(1, len(result.Rows))

	result, err = db.RunQueryString("SELECT ?r WHERE { ?r rdf:type brick:AHU };")
	require.NoError(err)
	require.Equal(3, len(result.Rows))

	result, err = db.RunQueryString("SELECT ?r WHERE { ?r rdf:type brick:INSERTED };")
	require.NoError(err)
	require.Equal(0, len(result.Rows))

	// insert
	_, err = db.RunQueryString("INSERT { ?r rdf:type brick:INSERTED } WHERE { ?r rdf:type brick:Room };")
	require.NoError(err)

	// query new version
	result, err = db.RunQueryString("SELECT ?r WHERE { ?r rdf:type brick:INSERTED };")
	require.NoError(err)
	require.Equal(1, len(result.Rows))
	result, err = db.RunQueryString("SELECT ?r WHERE { ?r rdf:type brick:AHU };")
	require.NoError(err)
	require.Equal(3, len(result.Rows))
}
