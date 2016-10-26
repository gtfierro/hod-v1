package main

import (
	"fmt"
	"github.com/gtfierro/hod/goraptor"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func benchLoad(c *cli.Context) error {
	if c.NArg() == 0 {
		return errors.New("Need to specify a turtle file to load")
	}
	filename := c.Args().Get(0)
	p := turtle.GetParser()
	ds, duration := p.Parse(filename)
	fmt.Printf("Loaded %d triples, %d namespaces in %s\n", ds.NumTriples(), ds.NumNamespaces(), duration)
	return nil
}
