package main

import (
	"os"

	"github.com/gtfierro/hod/version"
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
	app.Version = version.Release
	app.Usage = "BRICK database and query engine"

	app.Commands = []cli.Command{
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
			Name:   "rebuild",
			Usage:  "Delete old database and rebuild a-new, starting server afterwards",
			Action: rebuildServer,
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
		{
			Name:   "check",
			Usage:  "Check access to MDAL on behalf of some key",
			Action: doCheck,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "agent,a",
					Value:  "127.0.0.1:28589",
					Usage:  "Local BOSSWAVE Agent",
					EnvVar: "BW2_AGENT",
				},
				cli.StringFlag{
					Name:   "entity,e",
					Value:  "",
					Usage:  "The entity to use",
					EnvVar: "BW2_DEFAULT_ENTITY",
				},
				cli.StringFlag{
					Name:  "key, k",
					Usage: "The key or alias to check",
				},
				cli.StringFlag{
					Name:  "uri, u",
					Usage: "The base URI of MDAL",
				},
			},
		},
		{
			Name:   "grant",
			Usage:  "Grant access to MDAL to some key",
			Action: doGrant,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "agent,a",
					Value:  "127.0.0.1:28589",
					Usage:  "Local BOSSWAVE Agent",
					EnvVar: "BW2_AGENT",
				},
				cli.StringFlag{
					Name:   "entity,e",
					Value:  "",
					Usage:  "The entity to use",
					EnvVar: "BW2_DEFAULT_ENTITY",
				},
				cli.StringFlag{
					Name:   "bankroll, b",
					Value:  "",
					Usage:  "The entity to use for bankrolling",
					EnvVar: "BW2_DEFAULT_BANKROLL",
				},
				cli.StringFlag{
					Name:  "expiry",
					Usage: "Set the expiry on access to MDAL measured from now e.g. 3d7h20m",
				},
				cli.IntFlag{
					Name:  "ttl",
					Usage: "Set the TTL",
					Value: 0,
				},
				cli.StringFlag{
					Name:  "key, k",
					Usage: "The key or alias to check",
				},
				cli.StringFlag{
					Name:  "uri, u",
					Usage: "The base URI of MDAL",
				},
			},
		},
	}
	app.Run(os.Args)
}
