package main

import (
	"fmt"
	"github.com/kristofmartens/mproxy/internal/cli"
	"github.com/kristofmartens/mproxy/internal/mproxy"
	"os"
)

func main() {
	// Retrieve the configuration for the proxy server
	config := cli.ParseCli()

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
