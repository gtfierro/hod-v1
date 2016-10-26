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
	}
	app.Run(os.Args)
}
