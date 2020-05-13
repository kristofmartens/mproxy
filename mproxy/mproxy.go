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

// Use this token for test purposes
//const token = `eyJraWQiOiIxVTI3QU5cL1wvV3ZWWkszUGdKTlkwMGd3dFdXMWRTVjBJNjk5c05jVU5lbGM9IiwiYWxnIjoiUlMyNTYifQ.eyJzdWIiOiJlNDliYWQ2YS1jMGU2LTQ0N2YtYTgxZC04ZGEwYjUxZjFkNzQiLCJjb2duaXRvOmdyb3VwcyI6WyJzX2Rxc19kZXZlbCIsInNfeDBhX2FkbWluIiwic194MGFfYXVkaXQiLCJzX2FkYV9hdWRpdCIsInNfZHFzX2F1ZGl0Iiwic19kcXNfYWRtaW4iLCJhZG1pbiIsInNfYWRhX2RldmVsIiwic194MGFfZGV2ZWwiLCJzX2FkYV9hZG1pbiJdLCJ0b2tlbl91c2UiOiJhY2Nlc3MiLCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIiwiYXV0aF90aW1lIjoxNTg4Nzg0ODU0LCJpc3MiOiJodHRwczpcL1wvY29nbml0by1pZHAuZXUtd2VzdC0xLmFtYXpvbmF3cy5jb21cL2V1LXdlc3QtMV9KME1qUWV1YUMiLCJleHAiOjE1ODkzNjQyMzcsImlhdCI6MTU4OTM2MDYzNywidmVyc2lvbiI6MiwianRpIjoiMjE1NjVlMjgtMzk5Yi00NzA5LWI1NDQtYzlmMDE2ZmM3ZWQ5IiwiY2xpZW50X2lkIjoiMzlma2NoM2g0NWpoZ2xxMDhwNXBpdmdxdWwiLCJ1c2VybmFtZSI6IlNBTUxfSkU0MDU3NkBBQ0MtS0JDLUdST1VQLkNPTSJ9.Kr8WXt70DGio63wz_kLuNzjRgdU2zux5goPCyDaooRCBAFEeok9FrWR2WLOu3KOqzPt3MauXLRmvbeC1cFq5K97Orh5FAXjwC9XYs54YG3GUaMO3z4tueKY2MSR3z3BuOtA7ALs1dzKARyTb2WTvWEhKBprKluvgy-XieXqvIOwoO8BaNPh-rn3YMJwgCTpoyR3h1_T3VnBJ73IBIAJ5gdjKFD-liF3emouR9D43opVpwtjtn9pcDCn3Dckbd72clFAHT37hHT0SWzFdqbdzfwKLwB1PCaIWkmexjCMtv7l-cB1wrh9rQBAQ_y9_SqVumUDI_FZMP5q8B85agOkrdw`

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

	log.Println("Starting proxy server with config:\n", p)

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
