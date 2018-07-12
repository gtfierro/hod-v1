//go:generate statik -src=static
package server

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/gtfierro/hod/config"
	hod "github.com/gtfierro/hod/db"
	query "github.com/gtfierro/hod/lang"
	_ "github.com/gtfierro/hod/server/statik"
	"github.com/rakyll/statik/fs"

	"github.com/op/go-logging"
	//"github.com/pkg/profile"
	"golang.org/x/crypto/acme/autocert"
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
	db         *hod.HodDB
	port       string
	staticpath string
}

func StartHodServer(db *hod.HodDB, cfg *config.Config) *http.Server {
	server := &hodServer{
		db:         db,
		port:       cfg.ServerPort,
		staticpath: cfg.StaticPath,
	}
	log.Info("Static Path", cfg.StaticPath)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	// enable profiling if configured
	//	if cfg.EnableCPUProfile {
	//		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	//	} else if cfg.EnableMEMProfile {
	//		defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	//	} else if cfg.EnableBlockProfile {
	//		defer profile.Start(profile.BlockProfile, profile.ProfilePath(".")).Stop()
	//	}

	// configure server
	var (
		addrString string
		nettype    string
	)

	// check if ipv6
	if cfg.UseIPv6 {
		nettype = "tcp6"
		addrString = "[" + cfg.ListenAddress + "]:" + server.port
	} else {
		nettype = "tcp4"
		addrString = cfg.ListenAddress + ":" + server.port
	}

	address, err := net.ResolveTCPAddr(nettype, addrString)
	if err != nil {
		log.Fatalf("Error resolving address %s (%s)", addrString, err.Error())
	}

	http.Handle("/", http.FileServer(statikFS))
	http.HandleFunc("/api/query", server.handleQuery)
	http.HandleFunc("/api/querydot", server.handleQueryDot)
	http.HandleFunc("/api/queryclassdot", server.handleQueryClassDot)
	http.HandleFunc("/api/search", server.handleSearch)
	log.Notice("Starting HTTP Server on ", addrString)

	var srv *http.Server
	if cfg.TLSHost != "" {
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.TLSHost),
			Cache:      autocert.DirCache("certs"),
		}
		srv = &http.Server{
			Addr:      address.String(),
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}
		go func() {
			log.Warning(srv.ListenAndServeTLS("", ""))
		}()
	} else {
		srv = &http.Server{
			Addr: address.String(),
		}
		go func() {
			log.Warning(srv.ListenAndServe())
		}()
	}
	return srv
}

func (srv *hodServer) handleQuery(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	log.Infof("Query from %s", req.RemoteAddr)
	var querybytes = make([]byte, 2048)
	nbytes, err := req.Body.Read(querybytes)
	if err != nil && err != io.EOF {
		log.Error(err)
		rw.WriteHeader(400)
		rw.Write([]byte(err.Error()))
		return
	}
	querystring := string(querybytes[:nbytes])
	log.Debug(querystring)
	parsed, err := query.Parse(querystring)
	if err != nil {
		log.Error(err)
		rw.WriteHeader(400)
		rw.Write([]byte(err.Error()))
		return
	}

	// evaluate query
	res, err := srv.db.RunQuery(parsed)
	if err != nil {
		log.Error(err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	encoder := json.NewEncoder(rw)
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = encoder.Encode(res)
	if err != nil {
		log.Error(err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	return
}

func (srv *hodServer) handleSearch(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var search = struct {
		Query  string
		Number int
	}{}
	log.Infof("Query from %s", req.RemoteAddr)
	err := json.NewDecoder(req.Body).Decode(&search)
	if err != nil {
		log.Error(err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	res, err := srv.db.Search(search.Query, search.Number)
	encoder := json.NewEncoder(rw)
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = encoder.Encode(res)
	if err != nil {
		log.Error(err)
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}
	return
}

func (srv *hodServer) handleQueryDot(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	log.Infof("QueryDot from %s", req.RemoteAddr)

	var querybytes = make([]byte, 2048)
	nbytes, err := req.Body.Read(querybytes)
	if err != nil && err != io.EOF {
		log.Error(err)
		rw.WriteHeader(400)
		rw.Write([]byte(err.Error()))
		return
	}
	querystring := string(querybytes[:nbytes])
	log.Debug(querystring)
	dot, err := srv.db.QueryToDOT(querystring)
	if err != nil {
		log.Error(err)
		rw.WriteHeader(400)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Write([]byte(dot))
	return
}

func (srv *hodServer) handleQueryClassDot(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	log.Infof("QueryDot from %s", req.RemoteAddr)

	var querybytes = make([]byte, 2048)
	nbytes, err := req.Body.Read(querybytes)
	if err != nil && err != io.EOF {
		log.Error(err)
		rw.WriteHeader(400)
		rw.Write([]byte(err.Error()))
		return
	}
	querystring := string(querybytes[:nbytes])
	log.Debug(querystring)
	dot, err := srv.db.QueryToClassDOT(querystring)
	if err != nil {
		log.Error(err)
		rw.WriteHeader(400)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Write([]byte(dot))
	return
}
