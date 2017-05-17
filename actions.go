package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/gtfierro/hod/config"
	hod "github.com/gtfierro/hod/db"
	"github.com/gtfierro/hod/goraptor"
	"github.com/gtfierro/hod/query"
	"github.com/gtfierro/hod/server"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/montanaflynn/stats"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"gopkg.in/immesys/bw2bind.v5"
)

func benchLoad(c *cli.Context) error {
	if c.NArg() == 0 {
		log.Fatal("Need to specify a turtle file to load")
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
		log.Fatal("Need to specify a turtle file to load")
	}
	filename := c.Args().Get(0)
	cfg, err := config.ReadConfig(c.String("config"))
	if err != nil {
		return err
	}
	p := turtle.GetParser()
	ds, duration := p.Parse(filename)
	rate := float64((float64(ds.NumTriples()) / float64(duration.Nanoseconds())) * 1e9)
	log.Infof("Loaded %d triples, %d namespaces in %s (%.0f/sec)", ds.NumTriples(), ds.NumNamespaces(), duration, rate)

	cfg.ReloadBrick = true

	db, err := hod.NewDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.LoadDataset(ds)
	if err != nil {
		return err
	}

	return nil
}

func loadLinks(c *cli.Context) error {
	if c.NArg() == 0 {
		return errors.New("Need to specify a JSON file to load")
	}
	filename := c.Args().Get(0)
	cfg, err := config.ReadConfig(c.String("config"))
	if err != nil {
		return err
	}
	cfg.ReloadBrick = false
	db, err := hod.NewDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	var updates = new(hod.LinkUpdates)
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(updates); err != nil {
		return err
	}
	log.Noticef("Adding %d links, Removing %d links", len(updates.Adding), len(updates.Removing))
	return db.UpdateLinks(updates)
}

func startCLI(c *cli.Context) error {
	cfg, err := config.ReadConfig(c.String("config"))
	if err != nil {
		return err
	}
	cfg.ReloadBrick = false
	db, err := hod.NewDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	return runInteractiveQuery(db)
}

func startHTTP(c *cli.Context) error {
	cfg, err := config.ReadConfig(c.String("config"))
	if err != nil {
		return err
	}
	cfg.ReloadBrick = false
	db, err := hod.NewDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	server.StartHodServer(db, cfg)
	return nil
}

func startBOSSWAVE(c *cli.Context) error {
	cfg, err := config.ReadConfig(c.String("config"))
	if err != nil {
		return err
	}
	cfg.ReloadBrick = false
	db, err := hod.NewDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	client := bw2bind.ConnectOrExit(c.String("agent"))
	client.SetEntityFileOrExit(c.String("entity"))
	client.OverrideAutoChainTo(true)

	svc := client.RegisterService(c.String("uri"), "s.hod")
	iface := svc.RegisterInterface("_", "i.hod")
	queryChan, err := client.Subscribe(&bw2bind.SubscribeParams{
		URI: iface.SlotURI("query"),
	})
	if err != nil {
		return errors.Wrap(err, "Could not subscribe to HodDB query slot URI")
	}

	log.Notice("Serving query URI", iface.SlotURI("query"))

	const QueryPIDString = "2.0.10.1"
	//var QueryPID = bw2bind.FromDotForm(QueryPIDString)
	const ResponsePIDString = "2.0.10.2"
	var ResponsePID = bw2bind.FromDotForm(ResponsePIDString)
	type hodQuery struct {
		Query string
		Nonce string
	}
	type hodResponse struct {
		Result hod.QueryResult
		Nonce  string
	}

	handleBOSSWAVEQuery := func(msg *bw2bind.SimpleMessage) {
		var inq hodQuery
		po := msg.GetOnePODF(QueryPIDString)
		if po == nil {
			return
		}
		if obj, ok := po.(bw2bind.MsgPackPayloadObject); !ok {
			log.Error("Payload 2.0.10.1 was not MsgPack")
			return
		} else if err := obj.ValueInto(&inq); err != nil {
			log.Error(errors.Wrap(err, "Could not unmarshal into a hod query"))
			return
		}
		log.Info("Serving query", inq.Query)
		q, err := query.Parse(strings.NewReader(inq.Query))
		if err != nil {
			log.Error(errors.Wrap(err, "Could not parse hod query"))
			return
		}
		result := db.RunQuery(q)
		response := hodResponse{
			Result: result,
			Nonce:  inq.Nonce,
		}
		responsePO, err := bw2bind.CreateMsgPackPayloadObject(ResponsePID, response)
		if err != nil {
			log.Error(errors.Wrap(err, "Could not serialize hod response"))
			return
		}
		if err = iface.PublishSignal("result", responsePO); err != nil {
			log.Error(errors.Wrap(err, "Could not send hod response"))
			return
		}
	}

	for msg := range queryChan {
		go handleBOSSWAVEQuery(msg)
	}
	return nil
}

func doQuery(c *cli.Context) error {
	cfg, err := config.ReadConfig(c.String("config"))
	if err != nil {
		return err
	}
	cfg.ReloadBrick = false
	db, err := hod.NewDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	var (
		q   query.Query
		res hod.QueryResult
	)
	if c.String("query") != "" {
		q, err = query.Parse(strings.NewReader(c.String("query")))
		if err != nil {
			log.Fatal(err)
		}
	} else if c.String("file") != "" {
		file, err := os.Open(c.String("file"))
		if err != nil {
			log.Fatal(err)
		}
		q, err = query.Parse(file)
		if err != nil {
			log.Fatal(err)
		}
	}
	res = db.RunQuery(q)
	return res.DumpToCSV(c.Bool("prefixes"), db, os.Stdout)
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
			x := fmt.Sprintf("\"%s\";\n", triple.Subject.Value)
			nodes[x] = struct{}{}
		} else if triple.Predicate.String() == "http://www.w3.org/2000/01/rdf-schema#subClassOf" {
			if strings.HasPrefix(triple.Object.Value, "genid") || strings.HasPrefix(triple.Subject.Value, "genid") {
				continue
			}
			x := fmt.Sprintf("\"%s\" -> \"%s\" [label=\"%s\"];\n", triple.Object.Value, triple.Subject.Value, "hasSubclass")
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
	os.Remove(name)
	return nil
}

func dumpGraph(c *cli.Context) error {
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
		x := fmt.Sprintf("\"%s\";\n", triple.Subject.Value)
		nodes[x] = struct{}{}
		x = fmt.Sprintf("\"%s\" -> \"%s\" [label=\"%s\"];\n", triple.Subject.Value, triple.Object.Value, triple.Predicate.Value)
		edges[x] = struct{}{}
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
	cmd := exec.Command("sfdp", "-Tpdf", name)
	pdf, err := cmd.Output()
	if err != nil {
		// try graphviz dot then
		cmd = exec.Command("dot", "-Tpdf", name)
		pdf, err = cmd.Output()
		if err != nil {
			return err
		}
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
	os.Remove(name)
	return nil
}

func gethash() string {
	h := md5.New()
	seed := make([]byte, 16)
	binary.PutVarint(seed, time.Now().UnixNano())
	h.Write(seed)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func runInteractiveQuery(db *hod.DB) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	fmt.Println("Successfully loaded dataset!")
	bufQuery := ""

	//setup color for prompt
	c := color.New(color.FgCyan)
	c.Add(color.Bold)
	cyan := c.SprintFunc()

	rl, err := readline.NewEx(&readline.Config{
		Prompt:                 cyan("(hod)> "),
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
		rl.SetPrompt(cyan("(hod)> "))
		rl.SaveHistory(bufQuery)
		q, err := query.Parse(strings.NewReader(bufQuery))
		if err != nil {
			log.Error(err)
		} else {
			res := db.RunQuery(q)
			res.Dump()
		}
		bufQuery = ""
	}
	return nil
}

type uniqCounter map[string]struct{}

func (uc *uniqCounter) Add(item string) {
	(*uc)[item] = struct{}{}
}
func (uc *uniqCounter) GetCount() float64 {
	return float64(len(*uc))
}

// We load all turtle files into a unified graph and compute
// the following statistics, recording the following
// - in-degree per node
// - out-degree per node
// - # of nodes
// - # of edges
// - # of triples
func ttlStat(c *cli.Context) error {
	if c.NArg() == 0 {
		log.Fatal("Need to provide at least one TTL file")
	}
	numTriples := 0
	uniqueEdges := make(uniqCounter)
	uniqueNodes := make(uniqCounter)
	degree := make(map[string]int)
	predFreq := make(map[string]int)
	for fileidx := 0; fileidx < c.NArg(); fileidx++ {
		filename := c.Args().Get(fileidx)
		p := turtle.GetParser()
		ds, _ := p.Parse(filename)
		numTriples += ds.NumTriples()
		for _, triple := range ds.Triples {
			uniqueEdges.Add(triple.Predicate.String())
			uniqueNodes.Add(triple.Subject.String())
			uniqueNodes.Add(triple.Object.String())

			if cur, found := degree[triple.Subject.String()]; found {
				degree[triple.Subject.String()] = cur + 1
			} else {
				degree[triple.Subject.String()] = 1
			}
			if cur, found := degree[triple.Object.String()]; found {
				degree[triple.Object.String()] = cur + 1
			} else {
				degree[triple.Object.String()] = 1
			}

			if cur, found := predFreq[triple.Predicate.String()]; found {
				predFreq[triple.Predicate.String()] = cur + 1
			} else {
				predFreq[triple.Predicate.String()] = 1
			}
		}
	}
	// load into arrays so we can do stats
	var degreeCounts stats.Float64Data
	for _, count := range degree {
		degreeCounts = append(degreeCounts, float64(count))
	}
	var predFrequencyCounts stats.Float64Data
	for _, count := range predFreq {
		predFrequencyCounts = append(predFrequencyCounts, float64(count))
	}
	fmt.Printf("# Triples: %d\n", numTriples)
	fmt.Printf("# Unique Nodes: %0.0f\n", uniqueNodes.GetCount())
	fmt.Printf("# Unique Edges: %0.0f\n", uniqueEdges.GetCount())
	// compute density
	// for N nodes in the graph, max edges is N(N-1) for a directed graph with 1 type of edge.
	// For M types of edges, this is M*N*(N-1)
	// density is the number of edges we have out of this theoretical maximum. Each triple corresponds to 1 edge
	maxEdges := big.NewFloat(uniqueNodes.GetCount() * (uniqueNodes.GetCount() - 1) * uniqueEdges.GetCount())
	ntFloat := big.NewFloat(float64(numTriples))
	density := new(big.Float)
	density.Quo(ntFloat, maxEdges)
	fmt.Printf("Density: %s\n", density.String())
	sum_deg, _ := degreeCounts.Sum()
	min_deg, _ := degreeCounts.Min()
	max_deg, _ := degreeCounts.Max()
	mean_deg := sum_deg / uniqueNodes.GetCount()
	std_deg, _ := degreeCounts.StandardDeviation()
	fmt.Printf("Degree: Min %0.2f, Max %0.2f, Mean %0.2f, Std Dev %0.2f\n", min_deg, max_deg, mean_deg, std_deg)
	sum_pred, _ := predFrequencyCounts.Sum()
	min_pred, _ := predFrequencyCounts.Min()
	max_pred, _ := predFrequencyCounts.Max()
	mean_pred := sum_pred / uniqueEdges.GetCount()
	std_pred, _ := predFrequencyCounts.StandardDeviation()
	fmt.Printf("Pred Frequencies: Min %0.2f, Max %0.2f, Mean %0.2f, Std Dev %0.2f\n", min_pred, max_pred, mean_pred, std_pred)
	return nil
}
