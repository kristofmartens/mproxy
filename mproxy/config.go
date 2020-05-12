package mproxy

import (
	"fmt"
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
			Claims:  nil,
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

func (c Config) getJWTKeys() (string, error) {
	oidcCfg, err := httpGetJson(fmt.Sprintf("%s/%s", c.DiscoveryUrl, OidcConfig))
	if err != nil {
		return "", &Error{
			Code: ErrorHttpError,
			Msg:  "Could not retrieve discovery URL",
			Err:  err,
		}
	}

	jwtKeys, err := httpGetJson(oidcCfg[OidcConfigJwksUri].(string))
	if err != nil {
		return "", &Error{
			Code: ErrorHttpError,
			Msg:  "Could not retrieve JWT keys",
			Err:  err,
		}
	}

	return fmt.Sprintf("%v", jwtKeys), nil
}
