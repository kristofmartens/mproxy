package main

import (
	"flag"
	"fmt"
	"math"
	"mproxy/mproxy"
	"os"
)

func getConfig() mproxy.Config {
	port := flag.Uint("port", 8080, "Port proxy server will listen to")
	dest := flag.String("dest", "", "The address to forward the requests to")
	dUrl := flag.String("discovery-url", "", "The OAUTH discovery url")
	ath := flag.String("access-token-header", "X-Amzn-Oidc-Accesstoken",
		"HTTP header that contains the access token to verify")
	flag.Parse()

	switch {
	case *port > math.MaxUint16:
		fmt.Println("Invalid port")
		os.Exit(mproxy.ErrorInvalidConfig)
	case len(*dest) == 0:
		fmt.Println("No destination provided")
		os.Exit(mproxy.ErrorInvalidConfig)
	case len(*dUrl) == 0:
		fmt.Println("No or invalid discovery-url provided")
		os.Exit(mproxy.ErrorInvalidConfig)
	case len(*ath) == 0:
		fmt.Println("No or invalid access-token-header provided")
		os.Exit(mproxy.ErrorInvalidConfig)
	}

	config := mproxy.GetDefaultConfig()
	config.LocalPort = uint16(*port)
	config.Destination = *dest
	config.DiscoveryUrl = *dUrl
	config.AccessTokenHeader = *ath

	return config
}

func main() {
	// Retrieve the configuration for the proxy server
	config := getConfig()

	// Create the proxy server based on the provided configuration
	proxyServer, ok := mproxy.CreateProxy(config)
	if ok != nil {
		fmt.Println(ok.Error())
		os.Exit(ok.(*mproxy.Error).Code)
	}

	// Start the proxy server
	ok = proxyServer.StartProxy()
	if ok != nil {
		fmt.Println(ok.Error())
		os.Exit(ok.(*mproxy.Error).Code)
	}
}
