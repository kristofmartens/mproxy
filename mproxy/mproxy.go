package mproxy

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat/go-jwx/jwk"
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
	jwtKeys       *jwk.Set
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

// Returns function that will return the right key
func getKey(p *MProxy) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, &Error{
				Code: ErrorKeyError,
				Msg:  "No valid key id provided by discovery URL",
				Err:  nil,
			}
		}

		key := p.jwtKeys.LookupKeyID(kid)
		if len(key) == 0 {
			return nil, &Error{
				Code: ErrorKeyError,
				Msg:  "No valid keys provided by discovery URL",
				Err:  nil,
			}
		}

		return key[0].Materialize()
	}
}

func (p *MProxy) StartProxy() error {
	// Add proxy rule for each configured path
	for _, proxyRule := range p.config.ProxyRules {
		// Configure the different paths to proxy and their authorization rules
		http.HandleFunc(proxyRule.Pattern,
			func(writer http.ResponseWriter, request *http.Request) {
				// Check if there is an access token present
				token, ok := request.Header[p.config.AccessTokenHeader]
				if !ok || len(token) == 0 {
					defer request.Body.Close()
					log.Println("No access token present")
					writer.WriteHeader(403)
					return
				}

				// Check the validity of the JWT token
				tk, err := jwt.Parse(token[0], getKey(p))
				if err != nil {
					defer request.Body.Close()
					log.Println("Invalid token:", err)
					writer.WriteHeader(403)
					return
				}

				// Get the claims
				receivedClaims, ok := tk.Claims.(jwt.MapClaims)
				if !ok {
					defer request.Body.Close()
					log.Println("Invalid claims:", err)
					writer.WriteHeader(403)
					return
				}

				// Check the proxyRules for the right claims, if there aren't any, allow the request
				globalAllow := true
				for _, claim := range proxyRule.Claims {
					claimValues, ok := receivedClaims[claim.ClaimName]
					if !ok {
						defer request.Body.Close()
						log.Println("Mandatory claim not present:", claim.ClaimName)
						writer.WriteHeader(403)
						return
					}

					allowed := false
					for _, allowedClaim := range claim.AllowedClaims {
						for _, receivedClaim := range claimValues.([]interface{}) {
							if allowedClaim == receivedClaim.(string) {
								allowed = true
								break
							}
						}
						if allowed == true {
							break
						}
					}
					globalAllow = allowed && globalAllow
					if globalAllow == false {
						break
					}
				}

				if globalAllow == false {
					defer request.Body.Close()
					log.Println("Not Authorized to access this URL")
					writer.WriteHeader(403)
					return
				}

				// You are allowed to proxy this request
				// TODO: Add extra headers to identify the mproxy
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

	log.Println("Starting the mproxy server")

	// Start the actual proxy-ing
	if err := http.ListenAndServe(p.config.getListenAddress(), nil); err != nil {
		return &Error{
			Code: ErrorProxyInitError,
			Msg:  "Error running the proxy server",
			Err:  err,
		}
	}

	return nil
}

func (p *MProxy) setRunning(running bool) {
	p.running = running
}

func (p *MProxy) getJWTKeys() (*jwk.Set, error) {
	if p.jwtKeys == nil {
		var err error
		p.jwtKeys, err = p.config.getJWTKeys()
		if err != nil {
			return nil, err
		}
	}

	return p.jwtKeys, nil
}
