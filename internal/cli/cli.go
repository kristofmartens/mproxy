package cli

import (
	"github.com/kristofmartens/mproxy/internal/mproxy"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version = "0.1.0"
	cfg     = kingpin.Flag("config", "Path to config file defining authorization rules").Short('c').
		Required().PlaceHolder("filepath").Envar("MPROXY_CONFIG").String()
)

func ParseCli() mproxy.Config {
	kingpin.Version(version)
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	config := mproxy.GetDefaultConfig()

	err := mproxy.GetConfigFromFile(*cfg, &config)
	if err != nil {
		kingpin.FatalUsage("Could not read config file: %s", *cfg)
	}

	return config
}
