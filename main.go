package main

import (
	"flag"
	"fmt"
	"mproxy/mproxy"
	"os"
)

func getConfig() mproxy.Config {
	cfg := flag.String("config", "", "Path to config file defining authorization rules")
	flag.Parse()

	if len(*cfg) == 0 {
		fmt.Println("No or configuration provided")
		os.Exit(mproxy.ErrorInvalidConfig)
	}

	config := mproxy.GetDefaultConfig()

	err := mproxy.GetConfigFromFile(*cfg, &config)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(err.(*mproxy.Error).Code)
	}

	return config
}

func main() {
	// Retrieve the configuration for the proxy server
	config := getConfig()

	// Create the proxy server based on the provided configuration
	proxyServer, err := mproxy.CreateProxy(config)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(err.(*mproxy.Error).Code)
	}

	// Start the proxy server
	err = proxyServer.StartProxy()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(err.(*mproxy.Error).Code)
	}
}
