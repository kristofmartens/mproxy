package mproxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)


type MProxy struct {
	config        Config
	listenAddress string
	destURL       *url.URL
	running       bool
}

const (
	ErrorNoError = iota
	ErrorInvalidConfig
	ErrorProxyInitError
)

func CreateProxy(config Config) (MProxy, int) {
	ok := ErrorNoError

	if !IsValidConfig(config) {
		ok = ErrorInvalidConfig
	}

	mp := MProxy{
		config:        config,
		listenAddress: config.getListenAddress(),
		destURL:       config.getURL(),
		running:       false,
	}

	return mp, ok
}

func (p *MProxy) StartProxy() int {
	ok := ErrorNoError

	for _, proxyRule := range p.config.ProxyRules {
		// Configure the different paths to proxy and their authorization rules
		http.HandleFunc(proxyRule.Pattern,
			func(writer http.ResponseWriter, request *http.Request) {
				fmt.Println("claims config:", proxyRule.Claims)
				fmt.Println("headers:", request.Header)
				fmt.Println("Body:", request.Body)
				httputil.NewSingleHostReverseProxy(p.destURL).ServeHTTP(writer, request)
			})
	}

	// Set running state to true
	p.setRunning(true)
	defer p.setRunning(false)

	// Start the actual proxy-ing
	if err := http.ListenAndServe(p.config.getListenAddress(), nil); err != nil {
		ok = ErrorProxyInitError
		return ok
	}

	return ok
}

func (p *MProxy) setRunning(running bool) {
	p.running = running
}
