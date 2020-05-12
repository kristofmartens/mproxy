package mproxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	OidcConfig        = ".well-known/openid-configuration"
	OidcConfigJwksUri = "jwks_uri"
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
		return MProxy{}, &Error{
			Code: ErrorInvalidConfig,
			Msg:  "Could not create proxy, invalid configuration provided",
		}
	}

	jwtKeys, err := config.getJWTKeys()
	if err != nil {
		return MProxy{}, err
	}
	destUrl, err := config.getURL()
	if err != nil {
		return MProxy{}, err
	}

	mp := MProxy{
		config:        config,
		listenAddress: config.getListenAddress(),
		destURL:       destUrl,
		running:       false,
		jwtKeys:       jwtKeys,
	}

	return mp, nil
}

func (p *MProxy) StartProxy() error {
	// Add proxy rule for each configured path
	for _, proxyRule := range p.config.ProxyRules {
		// Configure the different paths to proxy and their authorization rules
		http.HandleFunc(proxyRule.Pattern,
			func(writer http.ResponseWriter, request *http.Request) {
				if len(request.Header[p.config.AccessTokenHeader]) == 0 {
					// There is no access token
					log.Println("No access token present")
					writer.WriteHeader(403)
				} else {
					httputil.NewSingleHostReverseProxy(p.destURL).ServeHTTP(writer, request)
				}
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

	log.Println("Starting proxy server with config:\n", p)

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

func (p *MProxy) getJWTKeys() (string, error) {
	if p.jwtKeys == "" {
		var err error
		p.jwtKeys, err = p.config.getJWTKeys()
		if err != nil {
			return "", err
		}
	}

	return p.jwtKeys, nil
}
