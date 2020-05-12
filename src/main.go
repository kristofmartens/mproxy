package main

import (
	"flag"
	"log"
	"math"
	"mproxy/mproxy"
)

func getConfig() mproxy.Config {
	port := flag.Uint("port", 8080, "Port proxy server will listen to")
	dest := flag.String("dest", "", "The address to forward the requests to")
	flag.Parse()

	switch {
	case *port > math.MaxUint16:
		log.Fatal("Invalid port")
	case len(*dest) == 0:
		log.Fatal("No destination provided")
	}

	config := mproxy.GetDefaultConfig()
	config.LocalPort = uint16(*port)
	config.Destination = *dest

	return config
}

func main() {
	// Retrieve the configuration for the proxy server
	config := getConfig()

	// Create the proxy server based on the provided configuration
	proxyServer, ok := mproxy.CreateProxy(config)
	if ok != mproxy.ErrorNoError {
		log.Fatal("Could not create proxy server: ", ok)
	}

	// Start the proxy server
	ok = proxyServer.StartProxy()
	if ok != mproxy.ErrorNoError {
		log.Fatal("Could not start proxy server: ", ok)
	}
}
