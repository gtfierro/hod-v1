package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBPath        string
	BrickFrameTTL string
	BrickClassTTL string
	ReloadBrick   bool

	ShowDependencyGraph    bool
	ShowQueryPlan          bool
	ShowQueryPlanLatencies bool
	ShowOperationLatencies bool
	ShowQueryLatencies     bool

	ServerPort string
	UseIPv6    bool
	Localhost  bool

	EnableCPUProfile   bool
	EnableMEMProfile   bool
	EnableBlockProfile bool
}

func init() {
	// set defaults for config
	viper.SetDefault("DBPath", "_hoddb")
	viper.SetDefault("BrickFrameTTL", "BrickFrame.ttl")
	viper.SetDefault("BrickClassTTL", "Brick.ttl")
	viper.SetDefault("ReloadBrick", true)

	viper.SetDefault("ShowDependencyGraph", false)
	viper.SetDefault("ShowQueryPlan", false)
	viper.SetDefault("ShowQueryPlanLatencies", false)
	viper.SetDefault("ShowOperationLatencies", false)
	viper.SetDefault("ShowQueryLatencies", true)

	viper.SetDefault("ServerPort", "47808")
	viper.SetDefault("UseIPv6", false)
	viper.SetDefault("Localhost", true)

	viper.SetDefault("EnableCPUProfile", false)
	viper.SetDefault("EnableMEMProfile", false)
	viper.SetDefault("EnableBlockProfile", false)

	viper.SetConfigName("hodconfig")
	// set search paths for config
	viper.AddConfigPath("/etc/hoddb/")
	viper.AddConfigPath(".")
}

func ReadConfig(file string) (*Config, error) {
	if len(file) > 0 {
		viper.SetConfigFile(file)
	}
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	c := &Config{
		DBPath:                 viper.GetString("DBPath"),
		BrickFrameTTL:          viper.GetString("BrickFrameTTL"),
		BrickClassTTL:          viper.GetString("BrickClassTTL"),
		ReloadBrick:            viper.GetBool("ReloadBrick"),
		ShowDependencyGraph:    viper.GetBool("ShowDependencyGraph"),
		ShowQueryPlan:          viper.GetBool("ShowQueryPlan"),
		ShowQueryPlanLatencies: viper.GetBool("ShowQueryPlanLatencies"),
		ShowOperationLatencies: viper.GetBool("ShowOperationLatencies"),
		ShowQueryLatencies:     viper.GetBool("ShowQueryLatencies"),
		ServerPort:             viper.GetString("ServerPort"),
		UseIPv6:                viper.GetBool("UseIPv6"),
		Localhost:              viper.GetBool("Localhost"),
		EnableCPUProfile:       viper.GetBool("EnableCPUProfile"),
		EnableMEMProfile:       viper.GetBool("EnableMEMProfile"),
		EnableBlockProfile:     viper.GetBool("EnableBlockProfile"),
	}
	return c, nil
}
