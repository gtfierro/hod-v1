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
	app.Version = "0.1"
	app.Usage = "BRICK database and query engine"

	app.Commands = []cli.Command{
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
	}
	app.Run(os.Args)
}
