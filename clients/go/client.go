package hod

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	turtle "github.com/gtfierro/hod/goraptor"
	"github.com/pkg/error"
)

type ResultMap map[string]turtle.URI
type Result struct {
	Rows    []ResultMap
	Count   int
	Elapsed time.Duration
}

// Placeholder for now
type Options struct {
	ValueOnly bool
}

func DefaultOptions() *Options {
	return &Options{
		ValueOnly: false,
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
