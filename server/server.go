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

	"github.com/julienschmidt/httprouter"
	"github.com/op/go-logging"
	"github.com/parnurzeal/gorequest"
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

type plotreq struct {
	URL   string
	UUIDs []string
}
type permalink struct {
	Autoupdate bool `json:"autoupdate"`
	Streams    []struct {
		Stream string `json:"stream"`
	} `json:"streams"`
	WindowType  string `json:"window_type"`
	WindowWidth int64  `json:"window_width"`
}

type hodServer struct {
	db         *hod.MultiDB
	port       string
	staticpath string
	router     *httprouter.Router
}

func StartHodServer(db *hod.MultiDB, cfg *config.Config) *http.Server {
	server := &hodServer{
		db:         db,
		port:       cfg.ServerPort,
		staticpath: cfg.StaticPath,
	}
	log.Info("Static Path", cfg.StaticPath)
	r := httprouter.New()

	r.POST("/api/query", server.handleQuery)
	r.POST("/api/querydot", server.handleQueryDot)
	r.POST("/api/queryclassdot", server.handleQueryClassDot)
	r.POST("/api/search", server.handleSearch)
	r.ServeFiles("/static/*filepath", http.Dir(cfg.StaticPath+"/static"))
	r.GET("/", server.serveQuery)
	r.GET("/query", server.serveQuery)
	r.GET("/help", server.serveHelp)
	r.GET("/plan", server.servePlanner)
	r.GET("/demo", server.serveDemo)
	r.POST("/permalink", server.permalink)
	//r.GET("/explore", server.serveExplorer)
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

func (srv *hodServer) handleQuery(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
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

func (srv *hodServer) serveDemo(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Infof("Serve demo from %s", req.RemoteAddr)
	defer req.Body.Close()
	http.ServeFile(rw, req, srv.staticpath+"/demo.html")
}

func (srv *hodServer) permalink(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Infof("Serve permalink from %s", req.RemoteAddr)
	defer req.Body.Close()
	var plot plotreq
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&plot); err != nil {
		log.Error(err)
		rw.WriteHeader(400)
		rw.Write([]byte(err.Error()))
		return
	}
	var pm = &permalink{
		Autoupdate:  true,
		WindowType:  "now",
		WindowWidth: 1e9 * 60 * 60 * 24,
	}
	for _, uuid := range plot.UUIDs {
		pm.Streams = append(pm.Streams, struct {
			Stream string `json:"stream"`
		}{uuid})
	}
	log.Debug(pm)
	r := gorequest.New()
	_, body, errs := r.Post(plot.URL + "/permalink").Send(pm).End()
	if len(errs) > 0 {
		log.Error(errs[0])
		rw.WriteHeader(500)
		rw.Write([]byte(errs[0].Error()))
		return
	}
	log.Debug(body)
	rw.Write([]byte(body))
}

func (srv *hodServer) serveExplorer(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Infof("Serve explorer from %s", req.RemoteAddr)
	defer req.Body.Close()
	http.ServeFile(rw, req, srv.staticpath+"/explore.html")
}

func (srv *hodServer) serveSearch(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	log.Infof("Serve search from %s", req.RemoteAddr)
	defer req.Body.Close()
	http.ServeFile(rw, req, srv.staticpath+"/search.html")
}

func (srv *hodServer) handleQueryDot(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
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

func (srv *hodServer) handleQueryClassDot(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
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
