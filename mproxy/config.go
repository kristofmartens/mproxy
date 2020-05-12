package mproxy

import (
	"fmt"
	"log"
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

func (c Config) getURL() *url.URL {
	destURL, ok := url.Parse(c.Destination)
	if ok != nil {
		log.Fatal("Invalid destination", ok)
	}
	return destURL
}
