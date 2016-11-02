package main

import (
	"fmt"
	"github.com/gtfierro/hod/db"
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
	rate := float64((float64(ds.NumTriples()) / float64(duration.Nanoseconds())) * 1e9)
	fmt.Printf("Loaded %d triples, %d namespaces in %s (%f/sec)\n", ds.NumTriples(), ds.NumNamespaces(), duration, rate)
	return nil
}

func load(c *cli.Context) error {
	if c.NArg() == 0 {
		return errors.New("Need to specify a turtle file to load")
	}
	filename := c.Args().Get(0)
	path := c.String("path")
	db, err := db.NewDB(path)
	if err != nil {
		return err
	}

	p := turtle.GetParser()
	ds, duration := p.Parse(filename)
	rate := float64((float64(ds.NumTriples()) / float64(duration.Nanoseconds())) * 1e9)
	fmt.Printf("Loaded %d triples, %d namespaces in %s (%f/sec)\n", ds.NumTriples(), ds.NumNamespaces(), duration, rate)

	err = db.LoadDataset(ds)
	if err != nil {
		return err
	}
	fmt.Println("Successfully loaded dataset!")

	return nil
}
