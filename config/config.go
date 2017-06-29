package config

import (
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

type Config struct {
	DBPath            string
	BrickFrameTTL     string
	BrickClassTTL     string
	ReloadBrick       bool
	DisableQueryCache bool

	EnableHTTP     bool
	EnableBOSSWAVE bool

	ShowNamespaces         bool
	ShowDependencyGraph    bool
	ShowQueryPlan          bool
	ShowQueryPlanLatencies bool
	ShowOperationLatencies bool
	ShowQueryLatencies     bool
	LogLevel               logging.Level

	ServerPort    string
	UseIPv6       bool
	ListenAddress string
	StaticPath    string
	TLSHost       string

	BW2_AGENT          string
	BW2_DEFAULT_ENTITY string
	HodURI             string

	EnableCPUProfile   bool
	EnableMEMProfile   bool
	EnableBlockProfile bool
}

func init() {
	prefix := os.Getenv("GOPATH")
	// switch prefix to default GOPATH /home/{user}/go
	if prefix == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		prefix = filepath.Join(u.HomeDir, "go")
	}
	// set defaults for config
	viper.SetDefault("DBPath", "_hoddb")
	viper.SetDefault("BrickFrameTTL", prefix+"/src/github.com/gtfierro/hod/BrickFrame.ttl")
	viper.SetDefault("BrickClassTTL", prefix+"/src/github.com/gtfierro/hod/Brick.ttl")
	viper.SetDefault("ReloadBrick", true)
	viper.SetDefault("DisableQueryCache", false)

	viper.SetDefault("EnableHTTP", true)
	viper.SetDefault("EnableBOSSWAVE", false)

	viper.SetDefault("ShowNamespaces", true)
	viper.SetDefault("ShowDependencyGraph", false)
	viper.SetDefault("ShowQueryPlan", false)
	viper.SetDefault("ShowQueryPlanLatencies", false)
	viper.SetDefault("ShowOperationLatencies", false)
	viper.SetDefault("ShowQueryLatencies", true)
	viper.SetDefault("LogLevel", "notice")

	viper.SetDefault("ServerPort", "47808")
	viper.SetDefault("UseIPv6", false)
	viper.SetDefault("ListenAddress", "127.0.0.1")
	viper.SetDefault("StaticPath", prefix+"/src/github.com/gtfierro/hod/server")
	viper.SetDefault("TLSHost", "") // disabled

	viper.SetDefault("HodURI", "scratch.ns/hod")

	viper.SetDefault("EnableCPUProfile", false)
	viper.SetDefault("EnableMEMProfile", false)
	viper.SetDefault("EnableBlockProfile", false)

	viper.SetConfigName("hodconfig")
	// set search paths for config
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/hoddb/")
	viper.AddConfigPath(prefix + "/src/github.com/gtfierro/hod")
}

func ReadConfig(file string) (*Config, error) {
	if len(file) > 0 {
		viper.SetConfigFile(file)
	}
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	viper.AutomaticEnv()

	level, err := logging.LogLevel(viper.GetString("LogLevel"))
	if err != nil {
		level = logging.DEBUG
	}

	c := &Config{
		DBPath:                 viper.GetString("DBPath"),
		BrickFrameTTL:          viper.GetString("BrickFrameTTL"),
		BrickClassTTL:          viper.GetString("BrickClassTTL"),
		ReloadBrick:            viper.GetBool("ReloadBrick"),
		EnableHTTP:             viper.GetBool("EnableHTTP"),
		EnableBOSSWAVE:         viper.GetBool("EnableBOSSWAVE"),
		DisableQueryCache:      viper.GetBool("DisableQueryCache"),
		ShowNamespaces:         viper.GetBool("ShowNamespaces"),
		ShowDependencyGraph:    viper.GetBool("ShowDependencyGraph"),
		ShowQueryPlan:          viper.GetBool("ShowQueryPlan"),
		ShowQueryPlanLatencies: viper.GetBool("ShowQueryPlanLatencies"),
		ShowOperationLatencies: viper.GetBool("ShowOperationLatencies"),
		ShowQueryLatencies:     viper.GetBool("ShowQueryLatencies"),
		LogLevel:               level,
		ServerPort:             viper.GetString("ServerPort"),
		UseIPv6:                viper.GetBool("UseIPv6"),
		ListenAddress:          viper.GetString("ListenAddress"),
		StaticPath:             viper.GetString("StaticPath"),
		TLSHost:                viper.GetString("TLSHost"),
		BW2_AGENT:              viper.GetString("BW2_AGENT"),
		BW2_DEFAULT_ENTITY:     viper.GetString("BW2_DEFAULT_ENTITY"),
		HodURI:                 viper.GetString("HodURI"),
		EnableCPUProfile:       viper.GetBool("EnableCPUProfile"),
		EnableMEMProfile:       viper.GetBool("EnableMEMProfile"),
		EnableBlockProfile:     viper.GetBool("EnableBlockProfile"),
	}
	return c, nil
}
