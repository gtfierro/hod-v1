package main

import (
	"os"

	"github.com/op/go-logging"
	"github.com/urfave/cli"
)

// logger
var log *logging.Logger

func init() {
	log = logging.MustGetLogger("hod")
	var format = "%{color}%{level} %{shortfile} %{time:Jan 02 15:04:05} %{color:reset} ▶ %{message}"
	var logBackend = logging.NewLogBackend(os.Stderr, "", 0)
	logBackendLeveled := logging.AddModuleLevel(logBackend)
	logging.SetBackend(logBackendLeveled)
	logging.SetFormatter(logging.MustStringFormatter(format))
}

func main() {
	app := cli.NewApp()
	app.Name = "hod"
	app.Version = "0.5.0"
	app.Usage = "BRICK database and query engine"

	app.Commands = []cli.Command{
		{
			Name:   "load",
			Usage:  "Load dataset into hoddb",
			Action: load,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Path to hoddb config file",
				},
			},
		},
		{
			Name:   "cli",
			Usage:  "Start hoddb from existing database",
			Action: startCLI,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Path to hoddb config file",
				},
			},
		},
		{
			Name:   "multi",
			Usage:  "Start multidb (BETA)",
			Action: startMultiDB,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Path to hoddb config file",
				},
			},
		},
		{
			Name:   "server",
			Usage:  "Start hoddb server from existing database. Default to HTTP server only, but can do both that and BOSSWAVE",
			Action: startServer,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Path to hoddb config file",
				},
			},
		},
		{
			Name:   "query",
			Usage:  "Query from command line (non-interactive)",
			Action: doQuery,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file, f",
					Usage: "Path to hoddb query file",
				},
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Path to hoddb config file",
				},
				cli.StringFlag{
					Name:  "query, q",
					Usage: "Query string",
				},
				cli.BoolFlag{
					Name:  "prefixes, p",
					Usage: "If true, abbreviate all namespaces. Else, just print the full URI",
				},
			},
		},
		{
			Name:   "search",
			Usage:  "Query for triples using just text matching",
			Action: doSearch,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Path to hoddb config file",
				},
				cli.StringFlag{
					Name:  "query, q",
					Usage: "Query string",
				},
				cli.IntFlag{
					Name:  "number, n",
					Usage: "Number of results to return",
					Value: 20,
				},
			},
		},
		{
			Name:   "benchload",
			Usage:  "Benchmark loading a turtle file",
			Action: benchLoad,
		},
		{
			Name:   "dump",
			Usage:  "Dump contents of turtle file",
			Action: dump,
		},
		{
			Name:   "viewclass",
			Usage:  "PDF visualization of class structure of file",
			Action: classGraph,
		},
		{
			Name:   "dumpgraph",
			Usage:  "PDF visualization of TTL file. WARNING this can get really big",
			Action: dumpGraph,
		},
		{
			Name:   "ttlstat",
			Usage:  "Outputs statistics on the provided TTL file. Loads all file provided as arguments",
			Action: ttlStat,
		},
	}
	app.Run(os.Args)
}
