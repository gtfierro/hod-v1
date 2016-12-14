package server

import (
	"encoding/json"
	"net"
	"net/http"
	"os"

	"github.com/gtfierro/hod/config"
	hod "github.com/gtfierro/hod/db"
	"github.com/gtfierro/hod/query"

	"github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
)

// logger
var log *logging.Logger

// set up logging facilities
func init() {
	log = logging.MustGetLogger("http")
	var format = "%{color}%{level} %{time:Jan 02 15:04:05} %{shortfile}%{color:reset} ▶ %{message}"
	var logBackend = logging.NewLogBackend(os.Stderr, "", 0)
	logBackendLeveled := logging.AddModuleLevel(logBackend)
	logging.SetBackend(logBackendLeveled)
	logging.SetFormatter(logging.MustStringFormatter(format))
}

type hodServer struct {
	db     *hod.DB
	port   string
	router *httprouter.Router
}

func StartHodServer(db *hod.DB, cfg *config.Config) {
	server := &hodServer{
		db:   db,
		port: cfg.ServerPort,
	}
	r := httprouter.New()

	// TODO: how do we handle loading in data? Need to have the multiple
	// concurrent buildings issue fixed first, but for now it is sufficient
	// to just have one server per building
	//r.POST("/load", server.handleLoad)
	r.POST("/query", server.handleQuery)
	server.router = r

	var (
		addrString string
		nettype    string
	)

	// check if ipv6
	if cfg.UseIPv6 {
		nettype = "tcp6"
	} else {
		nettype = "tcp4"
	}

	if cfg.Localhost {
		addrString = "localhost:" + server.port
	} else {
		addrString = "0.0.0.0:" + server.port
	}

	address, err := net.ResolveTCPAddr(nettype, addrString)
	if err != nil {
		log.Fatalf("Error resolving address %s (%s)", server.port, err.Error())
	}

	http.Handle("/", server.router)
	log.Notice("Starting HTTP Server on ", addrString)
	srv := &http.Server{
		Addr: address.String(),
	}
	srv.ListenAndServe()
}

func (srv *hodServer) handleQuery(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()

	parsed, err := query.Parse(req.Body)
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	// evaluate query
	res := srv.db.RunQuery(parsed)

	encoder := json.NewEncoder(rw)
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = encoder.Encode(res)
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	return
}
