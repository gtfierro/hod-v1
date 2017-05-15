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
	app.Version = "0.3.4"
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
			Name:   "http",
			Usage:  "Start hoddb HTTP server from existing database",
			Action: startHTTP,
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
			Name:   "bosswave",
			Usage:  "Expose HodDB over BOSSWAVE",
			Action: startBOSSWAVE,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "uri, u",
					Usage: "Base URI to expose the HodDB service on",
				},
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Path to hoddb config file",
				},
				cli.StringFlag{
					Name:   "entity, e",
					Usage:  "Path to BOSSWAVE entity file to serve",
					Value:  "",
					EnvVar: "BW2_DEFAULT_ENTITY",
				},
				cli.StringFlag{
					Name:   "agent, a",
					Usage:  "Address of BOSSWAVE agent to use",
					Value:  "127.0.0.1:28589",
					EnvVar: "BW2_AGENT",
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
