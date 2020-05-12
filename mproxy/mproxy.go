package mproxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type MProxy struct {
	config        Config
	listenAddress string
	destURL       *url.URL
	running       bool
	jwtKeys       string
}

func CreateProxy(config Config) (MProxy, error) {
	if !IsValidConfig(config) {
		ok := Error{
			Code: ErrorInvalidConfig,
			Msg:  "Could not create proxy, invalid configuration provided",
		}
		return MProxy{}, &ok
	}

	mp := MProxy{
		config:        config,
		listenAddress: config.getListenAddress(),
		destURL:       config.getURL(),
		running:       false,
	}

	return mp, nil
}

func (p *MProxy) StartProxy() error {
	// Add proxy rule for each configured path
	for _, proxyRule := range p.config.ProxyRules {
		// Configure the different paths to proxy and their authorization rules
		http.HandleFunc(proxyRule.Pattern,
			func(writer http.ResponseWriter, request *http.Request) {
				// TODO: verify the tokens running in the server
				fmt.Println("claims config:", proxyRule.Claims)
				fmt.Println("headers:", request.Header)
				fmt.Println("Body:", request.Body)
				httputil.NewSingleHostReverseProxy(p.destURL).ServeHTTP(writer, request)
			})
	}

	// Liveliness probe
	http.HandleFunc(p.config.livelinessPath, func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		if p.running {
			writer.WriteHeader(200)
		} else {
			writer.WriteHeader(500)
		}
	})

	// Set running state to true
	p.setRunning(true)
	defer p.setRunning(false)

	log.Println("Starting proxy server with config:\n", p.config)

	// Start the actual proxy-ing
	if ok := http.ListenAndServe(p.config.getListenAddress(), nil); ok != nil {
		return &Error{
			Code: ErrorProxyInitError,
			Msg:  "Error running the proxy server",
			Err:  ok,
		}
	}

	return nil
}

func (p *MProxy) setRunning(running bool) {
	p.running = running
}

func (p *MProxy) getJWTKeys() (string, error){
	if len(p.jwtKeys) == 0 {

	}
	return p.jwtKeys, nil
}
