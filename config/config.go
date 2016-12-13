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
	}
	return c, nil
}
