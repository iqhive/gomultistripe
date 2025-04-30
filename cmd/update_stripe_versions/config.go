package main

import "github.com/iqhive/cfggo"

type Config struct {
	cfggo.Structure
	Debug  func() bool `cfggo:"debug" default:"true" help:"Enable debug mode"`
	DryRun func() bool `cfggo:"dryrun" default:"true" help:"Enable dry run mode"`
}

var config Config

func loadConfig() {
	config.Init(&config)
}
