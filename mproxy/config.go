package mproxy

import (
	"fmt"
	"github.com/lestrrat/go-jwx/jwk"
	"net/url"
)

type OIDCClaim struct {
	ClaimName        string
	AllowedClaims    []string
	// TODO: Not yet implemented functionality
	RequireAllClaims bool
}

type ProxyRule struct {
	Pattern string
	Claims  []OIDCClaim
}

type Config struct {
	LocalPort         uint16
	Destination       string
	DiscoveryUrl      string
	AccessTokenHeader string
	ProxyRules        []ProxyRule
	livelinessPath    string
}

func IsValidConfig(config Config) bool {
	switch {
	case len(config.Destination) == 0:
		return false
	}

	return true
}

func GetDefaultConfig() Config {
	config := Config{
		LocalPort:      8080,
		livelinessPath: "/alive",
		ProxyRules: []ProxyRule{{
			Pattern: "/",
			Claims: []OIDCClaim{{
				ClaimName:        "cognito:groups",
				AllowedClaims:    []string{"admin", "s_ada_admin", "s_ada_devel"},
				RequireAllClaims: false,
			}},
		}},
	}

	return config
}

func (c Config) getListenAddress() string {
	return fmt.Sprintf(":%d", c.LocalPort)
}

func (c Config) getURL() (*url.URL, error) {
	destURL, err := url.Parse(c.Destination)
	if err != nil {
		return nil, &Error{
			Code: ErrorInvalidConfig,
			Msg:  "Could not parse URL",
			Err:  err,
		}
	}

	return destURL, err
}

func (c Config) getJWTKeys() (*jwk.Set, error) {
	oidcCfg, err := httpGetJson(fmt.Sprintf("%s/%s", c.DiscoveryUrl, OidcConfig))
	if err != nil {
		return nil, &Error{
			Code: ErrorHttpError,
			Msg:  "Could not retrieve discovery URL",
			Err:  err,
		}
	}

	jwtKeys, err := jwk.FetchHTTP(oidcCfg[OidcConfigJwksUri].(string))
	if err != nil {
		return nil, &Error{
			Code: ErrorHttpError,
			Msg:  "Could not retrieve JWT keys",
			Err:  err,
		}
	}

	return jwtKeys, nil
}
