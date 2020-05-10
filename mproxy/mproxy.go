package mproxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type OIDCClaim struct {
	ClaimName     string
	AllowedClaims []string
}

type ProxyRule struct {
	Pattern string
	Claims  []OIDCClaim
}

type Config struct {
	LocalPort   uint16
	Destination string
	ProxyRules []ProxyRule
}

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

func GetDefaultConfig() Config {
	config := Config{
		LocalPort:   8080,
		Destination: "",
		ProxyRules: []ProxyRule{{
			Pattern: "/",
			Claims:  nil,
		}},
	}

	return config
}

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

	// Configure the different paths to proxy and their authorization rules
	http.HandleFunc("/",
		func(writer http.ResponseWriter, request *http.Request) {
			fmt.Println("headers:", request.Header)
			fmt.Println("Body:", request.Body)
			httputil.NewSingleHostReverseProxy(p.destURL).ServeHTTP(writer, request)
		})

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

func IsValidConfig(config Config) bool {
	switch {
	case len(config.Destination) == 0:
		return false
	}

	return true
}

func (c Config) getListenAddress() string {
	return fmt.Sprintf(":%d", c.LocalPort)
}

func (c Config) getURL() *url.URL {
	destURL, ok := url.Parse(c.Destination)
	if ok != nil {
		log.Fatal("Invalid destination", ok)
	}
	return destURL
}
