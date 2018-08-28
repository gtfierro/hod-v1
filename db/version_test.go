package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/gtfierro/hod/config"
	"github.com/stretchr/testify/require"
)

func TestVersionQuery(t *testing.T) {
	require := require.New(t)

	triples1 := generateTriples(map[string]int{
		"Room": 1,
		"AHU":  3,
	})
	require.Equal(4, len(triples1))
	path1, err := triplesToFile("gentrip1", triples1)
	require.NoError(err)
	defer os.Remove(path1)

	triples2 := generateTriples(map[string]int{
		"Room": 10,
		"AHU":  3,
	})
	require.Equal(13, len(triples2))
	path2, err := triplesToFile("gentrip2", triples2)
	require.NoError(err)
	defer os.Remove(path2)

	cfgStr := fmt.Sprintf(`Buildings:
    %s: %s
    %s: %s
Ontologies: []
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
EnableHTTP: false`, "gentrip1", path1, "gentrip2", path2)
	cfg, err := config.ReadConfigFromString(cfgStr)
	require.NoError(err, cfgStr)
	db, err := NewHodDB(cfg)
	defer db.Close()
	require.NoError(err)

	_, err = db.RunQueryString("LIST NAMES;")
	require.NoError(err)
}
