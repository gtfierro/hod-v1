// This package provide a couple simple HodDB clients
// HTTP Client:
//
//
//    package main
//
//    import (
//    	"fmt"
//    	"github.com/gtfierro/hod/clients/go"
//    )
//
//    func main() {
//    	c := hod.NewHTTPClient("http://ciee.cal-sdb.org/api/query")
//
//    	query := `SELECT ?x ?r WHERE {
//        ?x rdf:type/rdfs:subClassOf* brick:Temperature_Sensor .
//        ?x bf:isLocatedIn ?r .
//      };`
//
//    	res, err := c.DoQuery(query, nil)
//    	fmt.Println(err)
//    	fmt.Printf("%+v", res)
//
//    }
//
//
// BOSSWAVE client
//
//    package main
//
//    import (
//    	"fmt"
//    	"github.com/gtfierro/hod/clients/go"
//	    bw2 "gopkg.in/immesys/bw2bind.v5"
//    )
//
//    func main() {
//
//	    client := bw2.ConnectOrExit("")
//	    client.OverrideAutoChainTo(true)
//	    client.SetEntityFromEnvironOrExit()
//	    bc, err := hod.NewBW2Client(client, "ciee/hod")
//	    if err != nil {
//          panic(err)
//      }
//
//    	query := `SELECT ?x ?r WHERE {
//        ?x rdf:type/rdfs:subClassOf* brick:Temperature_Sensor .
//        ?x bf:isLocatedIn ?r .
//      };`
//
//	    res, err := bc.DoQuery(query, nil)
//	    fmt.Println(err)
//	    fmt.Printf("%+v", res)
//
//    }
package hod

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/pkg/errors"
	bw2 "gopkg.in/immesys/bw2bind.v5"
)

var ErrNoResponse = errors.New("No response from archiver")

const RESULT_PONUM = `2.0.10.2/32`

var QUERY_PONUM = bw2.FromDotForm(`2.0.10.1`)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type ResultMap map[string]turtle.URI
type Result struct {
	Rows    []ResultMap
	Count   int
	Elapsed time.Duration
}

type Options struct {
	Timeout time.Duration
}

func DefaultOptions() *Options {
	return &Options{
		30 * time.Second,
	}
}

type HodClientHTTP struct {
	url string
}

func NewHTTPClient(url string) *HodClientHTTP {
	c := &HodClientHTTP{
		url: url,
	}

	return c
}

func (c *HodClientHTTP) DoQuery(query string, options *Options) (res Result, err error) {
	b := new(bytes.Buffer)
	b.WriteString(query)

	resp, err := http.Post(c.url, "application/json", b)
	defer resp.Body.Close()
	if err != nil {
		err = errors.Wrap(err, "Problem posting to HodDB")
		return
	}

	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		err = errors.Wrap(err, "Problem decoding response body")
		return
	}

	return
}

type HodClientBW2 struct {
	client  *bw2.BW2Client
	uri     string
	waiting map[string]chan BW2Result
	sync.RWMutex
}

type BW2Result struct {
	Nonce   string        `msgpack:"Nonce"`
	Error   error         `msgpack:-`
	err     string        `msgpack:"Error"`
	Rows    []ResultMap   `msgpack:"Rows"`
	Count   int           `msgpack:"Count"`
	Elapsed time.Duration `msgpack:-`
	elapsed int64         `msgpack:"Elapsed"`
}

type querymsg struct {
	Nonce string
	Query string
}

func NewBW2Client(client *bw2.BW2Client, uri string) (*HodClientBW2, error) {
	c := &HodClientBW2{
		client:  client,
		uri:     strings.TrimSuffix(uri, "/") + "/s.hod/_/i.hod",
		waiting: make(map[string]chan BW2Result),
	}

	sub, err := c.client.Subscribe(&bw2.SubscribeParams{
		URI: c.uri + "/signal/result",
	})
	if err != nil {
		return c, errors.Wrapf(err, "Could not subscribe to HodDB (%s)", uri)
	}

	go func() {
		for msg := range sub {
			po := msg.GetOnePODF(RESULT_PONUM)
			if po == nil {
				continue
			}

			var res BW2Result

			if err := po.(bw2.MsgPackPayloadObject).ValueInto(&res); err != nil {
				res.Error = err
			} else {
				res.Error = nil
			}

			c.Lock()
			if replyChan, found := c.waiting[res.Nonce]; found {
				select {
				case replyChan <- res:
				default:
				}
			}
			c.Unlock()
		}
	}()
	return c, nil
}

func (c *HodClientBW2) markWaitFor(nonce string, replyChan chan BW2Result) {
	c.Lock()
	c.waiting[nonce] = replyChan
	c.Unlock()
}

func (bc *HodClientBW2) DoQuery(query string, opts *Options) (res BW2Result, err error) {
	nonce := strconv.FormatUint(rand.Uint64(), 10)
	msg := querymsg{nonce, query}
	po, _ := bw2.CreateMsgPackPayloadObject(QUERY_PONUM, msg)

	if err = bc.client.Publish(&bw2.PublishParams{
		URI:            bc.uri + "/slot/query",
		PayloadObjects: []bw2.PayloadObject{po},
	}); err != nil {
		err = err
		return
	}

	if opts == nil {
		opts = DefaultOptions()
	}

	timeoutChan := time.After(opts.Timeout)

	replyChan := make(chan BW2Result, 1)
	bc.markWaitFor(nonce, replyChan)

	for {
		select {
		case <-timeoutChan:
			err = ErrNoResponse
			return
		case res = <-replyChan:
			bc.Lock()
			delete(bc.waiting, nonce)
			bc.Unlock()
			if res.Error != nil {
				err = res.Error
			}
			return

		}
	}
	err = nil

	return
}
