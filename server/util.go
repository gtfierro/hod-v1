package server

import (
	"encoding/json"
	"github.com/gtfierro/hod/query"
	"io"
	"net/http"
	"strings"
)

func parseQueryInRequest(req *http.Request) (query.Query, error) {
	var queryreader io.Reader
	if req.Header.Get("Content-Type") == "application/json" {
		dec := json.NewDecoder(req.Body)
		var s string
		err := dec.Decode(&s)
		if err != nil {
			return query.Query{}, err
		}
		queryreader = strings.NewReader(s)
	} else {
		queryreader = req.Body
	}

	return query.Parse(queryreader)
}
