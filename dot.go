package main

import (
	"bytes"
	"strings"

	"github.com/immesys/bw2/crypto"
	"github.com/immesys/bw2bind"
)

func ns(uri string) string {
	return strings.Split(uri, "/")[0]
}

// going to check xbos/hod from namespace
func allowVK(uri string, fromVK, toVK []byte, client *bw2bind.BW2Client) bool {
	dots, validities, err := client.FindDOTsFromVK(crypto.FmtKey(fromVK))
	if err != nil {
		log.Error(err)
		return false
	}
	for idx, d := range dots {
		if validities[idx] != bw2bind.StateValid {
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
