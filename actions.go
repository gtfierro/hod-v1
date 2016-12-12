package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	hod "github.com/gtfierro/hod/db"
	"github.com/gtfierro/hod/goraptor"
	query "github.com/gtfierro/hod/query"

	"github.com/chzyer/readline"
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
	fmt.Printf("Loaded %d triples, %d namespaces in %s (%.0f/sec)\n", ds.NumTriples(), ds.NumNamespaces(), duration, rate)
	return nil
}

func load(c *cli.Context) error {
	if c.NArg() == 0 {
		return errors.New("Need to specify a turtle file to load")
	}
	filename := c.Args().Get(0)
	path := c.String("path")
	p := turtle.GetParser()
	ds, duration := p.Parse(filename)
	rate := float64((float64(ds.NumTriples()) / float64(duration.Nanoseconds())) * 1e9)
	log.Infof("Loaded %d triples, %d namespaces in %s (%.0f/sec)", ds.NumTriples(), ds.NumNamespaces(), duration, rate)

	frame := c.String("frame")
	relships, _ := p.Parse(frame)

	class := c.String("class")
	classships, _ := p.Parse(class)

	db, err := hod.NewDB(path)
	if err != nil {
		return err
	}
	err = db.LoadRelationships(relships)
	if err != nil {
		return err
	}
	err = db.LoadDataset(classships)
	if err != nil {
		return err
	}
	err = db.LoadDataset(ds)
	if err != nil {
		return err
	}
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	fmt.Println("Successfully loaded dataset!")
	bufQuery := ""
	rl, err := readline.NewEx(&readline.Config{
		Prompt:                 "(hod)> ",
		HistoryFile:            currentUser.HomeDir + "/.hod-query-history",
		DisableAutoSaveHistory: true,
	})
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		if len(line) == 0 {
			continue
		}
		bufQuery += line + " "
		if !strings.HasSuffix(strings.TrimSpace(line), ";") {
			rl.SetPrompt(">>> ...")
			continue
		}
		rl.SetPrompt("(hod)> ")
		rl.SaveHistory(bufQuery)
		q, err := query.Parse(strings.NewReader(bufQuery))
		if err != nil {
			log.Error(err)
		} else {
			db.RunQuery(q)
		}
		bufQuery = ""
	}

	return nil
}

func dump(c *cli.Context) error {
	if c.NArg() == 0 {
		return errors.New("Need to specify a turtle file to load")
	}
	filename := c.Args().Get(0)
	p := turtle.GetParser()
	ds, _ := p.Parse(filename)
	for _, triple := range ds.Triples {
		var s = triple.Subject.Value
		var p = triple.Predicate.Value
		var o = triple.Object.Value
		for pfx, full := range ds.Namespaces {
			if triple.Subject.Namespace == full {
				s = pfx + ":" + s
			}
			if triple.Predicate.Namespace == full {
				p = pfx + ":" + p
			}
			if triple.Object.Namespace == full {
				o = pfx + ":" + o
			}
		}
		fmt.Printf("%s\t%s\t%s\n", s, p, o)
	}
	return nil
}

func classGraph(c *cli.Context) error {
	if c.NArg() == 0 {
		return errors.New("Need to specify a turtle file to load")
	}
	filename := c.Args().Get(0)
	p := turtle.GetParser()
	ds, _ := p.Parse(filename)

	name := gethash() + ".gv"
	f, err := os.Create(name)
	if err != nil {
		return err
	}

	nodes := make(map[string]struct{})
	edges := make(map[string]struct{})
	for _, triple := range ds.Triples {
		if triple.Predicate.String() == "http://www.w3.org/1999/02/22-rdf-syntax-ns#type" && triple.Object.String() == "http://www.w3.org/2002/07/owl#Class" {
			x := fmt.Sprintf("%s;\n", triple.Subject.Value)
			nodes[x] = struct{}{}
		} else if triple.Predicate.String() == "http://www.w3.org/2000/01/rdf-schema#subClassOf" {
			if strings.HasPrefix(triple.Object.Value, "genid") || strings.HasPrefix(triple.Subject.Value, "genid") {
				continue
			}
			x := fmt.Sprintf("%s -> %s [label=\"%s\"];\n", triple.Object.Value, triple.Subject.Value, "hasSubclass")
			edges[x] = struct{}{}
		}
	}

	fmt.Fprintln(f, "digraph G {")
	fmt.Fprintln(f, "ratio=\"auto\"")
	fmt.Fprintln(f, "rankdir=\"LR\"")
	fmt.Fprintln(f, "size=\"7.5,10\"")
	for node := range nodes {
		fmt.Fprintf(f, node)
	}
	for edge := range edges {
		fmt.Fprintf(f, edge)
	}
	fmt.Fprintln(f, "}")
	cmd := exec.Command("dot", "-Tpdf", name)
	pdf, err := cmd.Output()
	if err != nil {
		return err
	}
	f2, err := os.Create(filename + ".pdf")
	if err != nil {
		return err
	}
	_, err = f2.Write(pdf)
	if err != nil {
		return err
	}

	// remove DOT file
	//os.Remove(name)
	return nil
}

func gethash() string {
	h := md5.New()
	seed := make([]byte, 16)
	binary.PutVarint(seed, time.Now().UnixNano())
	h.Write(seed)
	return fmt.Sprintf("%x", h.Sum(nil))
}
