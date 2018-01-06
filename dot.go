package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/immesys/bw2/crypto"
	"github.com/immesys/bw2/objects"
	"github.com/immesys/bw2/util"
	bw2 "github.com/immesys/bw2bind"
	"github.com/mgutz/ansi"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func ns(uri string) string {
	return strings.Split(uri, "/")[0]
}

// going to check xbos/hod from namespace
func allowVK(uri string, fromVK, toVK []byte, client *bw2.BW2Client) bool {
	dots, validities, err := client.FindDOTsFromVK(crypto.FmtKey(fromVK))
	if err != nil {
		log.Error(err)
		return false
	}
	for idx, d := range dots {
		if validities[idx] != bw2.StateValid {
			continue
		}
		nskey, _ := crypto.UnFmtKey(ns(uri))
		if !bytes.Equal(d.GetAccessURIMVK(), nskey) {
			continue
		}
		if bytes.Equal(d.GetReceiverVK(), toVK) && d.GetAccessURISuffix() == "hod" {
			return true
		}
		if allowVK(uri, d.GetReceiverVK(), toVK, client) {
			return true
		}
	}
	return false
}

func resolveKey(client *bw2.BW2Client, key string) (string, error) {
	if _, err := os.Stat(key); err != nil && !os.IsNotExist(err) {
		return "", errors.Wrap(err, "Could not check key file")
	} else if err == nil {
		// have a file and load it!
		contents, err := ioutil.ReadFile(key)
		if err != nil {
			return "", errors.Wrap(err, "Could not read file")
		}
		entity, err := objects.NewEntity(int(contents[0]), contents[1:])
		if err != nil {
			return "", errors.Wrap(err, "Could not decode entity from file")
		}
		ent, ok := entity.(*objects.Entity)
		if !ok {
			return "", errors.New(fmt.Sprintf("File was not an entity: %s", key))
		}
		key_vk := objects.FmtKey(ent.GetVK())
		return key_vk, nil
	} else {
		// resolve key from registry
		a, b, err := client.ResolveRegistry(key)
		if err != nil {
			return "", errors.Wrapf(err, "Could not resolve key %s", key)
		}
		if b != bw2.StateValid {
			return "", errors.New(fmt.Sprintf("Key was not valid: %s", key))
		}
		ent, ok := a.(*objects.Entity)
		if !ok {
			return "", errors.New(fmt.Sprintf("Key was not an entity: %s", key))
		}
		key_vk := objects.FmtKey(ent.GetVK())
		return key_vk, nil
	}
}

func doCheck(c *cli.Context) error {
	bw2.SilenceLog()
	key := c.String("key")
	if key == "" {
		log.Fatal(errors.New("Need to specify key"))
	}
	uri := c.String("uri")
	if uri == "" {
		log.Fatal(errors.New("Need to specify uri"))
	}
	entity := c.String("entity")
	agent := c.String("agent")
	// connect
	bwclient := bw2.ConnectOrExit(agent)
	bwclient.SetEntityFileOrExit(entity)
	bwclient.OverrideAutoChainTo(true)
	_, _, err := checkAccess(bwclient, key, uri)
	if err != nil {
		log.Error("Likely that key does not have access to archiver")
		log.Fatal(err)
	}
	return nil
}

func doGrant(c *cli.Context) error {
	bw2.SilenceLog()
	key := c.String("key")
	if key == "" {
		log.Fatal(errors.New("Need to specify key"))
	}
	uri := c.String("uri")
	if uri == "" {
		log.Fatal(errors.New("Need to specify uri"))
	}
	entity := c.String("entity")
	bankroll := c.String("bankroll")
	agent := c.String("agent")
	if c.String("expiry") == "" {
		log.Fatal(errors.New("Need to specify expiry"))
	}
	expiry, err := util.ParseDuration(c.String("expiry"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "Could not parse expiry"))
	}
	// connect
	bwclient := bw2.ConnectOrExit(agent)
	bwclient.SetEntityFileOrExit(entity)
	bwclient.OverrideAutoChainTo(true)

	uris, access, err := checkAccess(bwclient, key, uri)
	if err != nil {
		log.Error(err, "(This is probably OK)")
	}

	key_vk, err := resolveKey(bwclient, key)
	if err != nil {
		log.Fatal(err)
	}

	datmoney := bw2.ConnectOrExit(agent)
	datmoney.SetEntityFileOrExit(bankroll)
	datmoney.OverrideAutoChainTo(true)

	var dotsToPublish [][]byte
	var hashToPublish []string
	var urisToPublish []string
	successcolor := ansi.ColorFunc("green")

	// scan URI
	scanAccess := access[0]
	if !scanAccess {
		// grant dot
		scanURI := uris[0]
		params := &bw2.CreateDOTParams{
			To:                key_vk,
			TTL:               0,
			Comment:           fmt.Sprintf("Access to Hod on URI %s", uri),
			URI:               scanURI,
			ExpiryDelta:       expiry,
			AccessPermissions: "C*",
		}
		hash, blob, err := bwclient.CreateDOT(params)
		if err != nil {
			log.Fatal(errors.Wrap(err, fmt.Sprintf("Could not grant DOT to %s on %s with permissions C*", key_vk, scanURI)))
		}
		log.Info("Granting DOT", hash)
		dotsToPublish = append(dotsToPublish, blob)
		hashToPublish = append(hashToPublish, hash)
		urisToPublish = append(urisToPublish, scanURI)
	}

	// query URI
	queryAccess := access[1]
	if !queryAccess {
		// grant dot
		queryURI := uris[1]
		params := &bw2.CreateDOTParams{
			To:                key_vk,
			TTL:               0,
			Comment:           fmt.Sprintf("Access to archiver on URI %s", uri),
			URI:               queryURI,
			ExpiryDelta:       expiry,
			AccessPermissions: "P",
		}
		hash, blob, err := bwclient.CreateDOT(params)
		if err != nil {
			log.Fatal(errors.Wrap(err, fmt.Sprintf("Could not grant DOT to %s on %s with permissions P", key_vk, queryURI)))
		}
		log.Info("Granting DOT", hash)
		dotsToPublish = append(dotsToPublish, blob)
		hashToPublish = append(hashToPublish, hash)
		urisToPublish = append(urisToPublish, queryURI)
	}

	// response URI
	responseAccess := access[2]
	if !responseAccess {
		// grant dot
		responseURI := uris[2]
		params := &bw2.CreateDOTParams{
			To:                key_vk,
			TTL:               0,
			Comment:           fmt.Sprintf("Access to archiver on URI %s", uri),
			URI:               responseURI,
			ExpiryDelta:       expiry,
			AccessPermissions: "C",
		}
		hash, blob, err := bwclient.CreateDOT(params)
		if err != nil {
			log.Fatal(errors.Wrap(err, fmt.Sprintf("Could not grant DOT to %s on %s with permissions C", key_vk, responseURI)))
		}
		log.Info("Granting DOT", hash)
		dotsToPublish = append(dotsToPublish, blob)
		hashToPublish = append(hashToPublish, hash)
		urisToPublish = append(urisToPublish, responseURI)
	}

	var wg sync.WaitGroup
	wg.Add(len(dotsToPublish))
	quit := make(chan bool)

	for idx, blob := range dotsToPublish {
		blob := blob
		hash := hashToPublish[idx]
		uri := urisToPublish[idx]
		go func(blob []byte, hash string) {
			log.Info("Publishing DOT", hash)
			defer wg.Done()
			a, err := datmoney.PublishDOT(blob)
			if err != nil {
				log.Error(errors.Wrap(err, fmt.Sprintf("Could not publish DOT with hash %s (%s)", hash, uri)))
			} else {
				log.Info(successcolor(fmt.Sprintf("Successfully published DOT %s (%s)", a, uri)))
			}
		}(blob, hash)
	}
	// "status bar"
	go func() {
		tick := time.Tick(2 * time.Second)
		for {
			select {
			case <-quit:
				return
			case <-tick:
				fmt.Print(".")
			}
		}
	}()
	wg.Wait()
	quit <- true // quit the progress bar

	return nil
}

func checkAccess(bwclient *bw2.BW2Client, key, uri string) (uris []string, hasPermission []bool, err error) {

	successcolor := ansi.ColorFunc("green")
	//foundcolor := ansi.ColorFunc("blue+h")
	//badcolor := ansi.ColorFunc("yellow+b")

	key_vk, err := resolveKey(bwclient, key)
	if err != nil {
		return
	}

	scanURI := uri + "/*/s.hod/!meta/lastalive"         // (C*)
	queryURI := uri + "/s.hod/_/i.hod/slot/query"       // (P)
	responseURI := uri + "/s.hod/_/i.hod/signal/result" // (C)
	uris = []string{scanURI, queryURI, responseURI}
	hasPermission = []bool{false, false, false}

	// now check access
	chain, err := bwclient.BuildAnyChain(scanURI, "C*", key_vk)
	if err != nil {
		err = errors.Wrapf(err, "Could not build chain on %s to %s", scanURI, key_vk)
		return
	}
	if chain == nil {
		err = errors.New(fmt.Sprintf("Key %s does not have a chain to find archivers (%s)", key_vk, scanURI))
		return
	} else {
		hasPermission[0] = true
		fmt.Printf("Hash: %s  Permissions: %s%s URI: %s\n", chain.Hash, chain.Permissions, strings.Repeat(" ", 5-len(chain.Permissions)), chain.URI)
	}

	chain, err = bwclient.BuildAnyChain(queryURI, "P", key_vk)
	if err != nil {
		err = errors.Wrapf(err, "Could not build chain on %s to %s", queryURI, key_vk)
		return
	}
	if chain == nil {
		err = errors.New(fmt.Sprintf("Key %s does not have a chain to publish to the archiver (%s)", key_vk, queryURI))
		return
	} else {
		hasPermission[1] = true
		fmt.Printf("Hash: %s  Permissions: %s%s URI: %s\n", chain.Hash, chain.Permissions, strings.Repeat(" ", 5-len(chain.Permissions)), chain.URI)
	}

	chain, err = bwclient.BuildAnyChain(responseURI, "C", key_vk)
	if err != nil {
		err = errors.Wrapf(err, "Could not build chain on %s to %s", responseURI, key_vk)
		return
	}
	if chain == nil {
		err = errors.New(fmt.Sprintf("Key %s does not have a chain to consume from the archiver (%s)", key_vk, responseURI))
		return
	} else {
		hasPermission[2] = true
		fmt.Printf("Hash: %s  Permissions: %s%s URI: %s\n", chain.Hash, chain.Permissions, strings.Repeat(" ", 5-len(chain.Permissions)), chain.URI)
	}

	fmt.Println(successcolor(fmt.Sprintf("Key %s has access to archiver at %s\n", key_vk, uri)))

	return
}
