package server

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"os"

	"github.com/gtfierro/hod/config"
	hod "github.com/gtfierro/hod/db"
	"github.com/gtfierro/hod/query"

	"github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
	"github.com/pkg/profile"
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
	db         *hod.DB
	port       string
	staticpath string
	router     *httprouter.Router
}

func StartHodServer(db *hod.DB, cfg *config.Config) {
	server := &hodServer{
		db:         db,
		port:       cfg.ServerPort,
		staticpath: cfg.StaticPath,
	}
	r := httprouter.New()

	// TODO: how do we handle loading in data? Need to have the multiple
	// concurrent buildings issue fixed first, but for now it is sufficient
	// to just have one server per building
	r.POST("/api/query", server.handleQuery)
	r.POST("/api/querydot", server.handleQueryDot)
	r.POST("/api/queryclassdot", server.handleQueryClassDot)
	r.POST("/api/search", server.handleSearch)
	r.ServeFiles("/static/*filepath", http.Dir(cfg.StaticPath+"/static"))
	r.GET("/", server.serveQuery)
	r.GET("/query", server.serveQuery)
	r.GET("/help", server.serveHelp)
	r.GET("/plan", server.servePlanner)
	r.GET("/explore", server.serveExplorer)
	r.GET("/explore2", server.serveExplorer2)
	r.GET("/search", server.serveSearch)
	server.router = r

	// enable profiling if configured
	if cfg.EnableCPUProfile {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	} else if cfg.EnableMEMProfile {
		defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	} else if cfg.EnableBlockProfile {
		defer profile.Start(profile.BlockProfile, profile.ProfilePath(".")).Stop()
	}

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

	http.Handle("/", server.router)
	log.Notice("Starting HTTP Server on ", addrString)

	if cfg.TLSHost != "" {
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.TLSHost),
			Cache:      autocert.DirCache("certs"),
		}
		s := &http.Server{
			Addr:      address.String(),
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}
		log.Fatal(s.ListenAndServeTLS("", ""))
	} else {
		srv := &http.Server{
			Addr: address.String(),
		}
		log.Fatal(srv.ListenAndServe())
	}
}

func (srv *hodServer) handleQuery(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()

	log.Infof("Query from %s", req.RemoteAddr)
	parsed, err := query.Parse(req.Body)
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

func (srv *hodServer) handleSearch(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
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

func (srv *hodServer) serveHelp(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Infof("Serve help from %s", req.RemoteAddr)
	defer req.Body.Close()
	http.ServeFile(rw, req, srv.staticpath+"/help.html")
}

func (srv *hodServer) serveQuery(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Infof("Serve query from %s", req.RemoteAddr)
	defer req.Body.Close()
	http.ServeFile(rw, req, srv.staticpath+"/query.html")
}

func (srv *hodServer) servePlanner(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Infof("Serve planner from %s", req.RemoteAddr)
	defer req.Body.Close()
	http.ServeFile(rw, req, srv.staticpath+"/plan.html")
}

func (srv *hodServer) serveExplorer(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Infof("Serve explorer from %s", req.RemoteAddr)
	defer req.Body.Close()
	http.ServeFile(rw, req, srv.staticpath+"/explore.html")
}
func (srv *hodServer) serveExplorer2(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Infof("Serve explorer from %s", req.RemoteAddr)
	defer req.Body.Close()
	http.ServeFile(rw, req, srv.staticpath+"/explore2.html")
}

func (srv *hodServer) serveSearch(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Infof("Serve search from %s", req.RemoteAddr)
	defer req.Body.Close()
	http.ServeFile(rw, req, srv.staticpath+"/search.html")
}

func (srv *hodServer) handleQueryDot(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()
	log.Infof("QueryDot from %s", req.RemoteAddr)

	dot, err := srv.db.QueryToDOT(req.Body)
	if err != nil {
		log.Error(err)
		rw.WriteHeader(400)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Write([]byte(dot))
	return
}

func (srv *hodServer) handleQueryClassDot(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer req.Body.Close()
	log.Infof("QueryDot from %s", req.RemoteAddr)

	dot, err := srv.db.QueryToClassDOT(req.Body)
	if err != nil {
		log.Error(err)
		rw.WriteHeader(400)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Write([]byte(dot))
	return
}
