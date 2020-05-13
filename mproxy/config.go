package mproxy

import (
	"fmt"
	"github.com/lestrrat/go-jwx/jwk"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
)

type OIDCClaim struct {
	ClaimName     string   `yaml:"claimName"`
	AllowedClaims []string `yaml:"allowedClaims"`
	// TODO: Not yet implemented functionality
	RequireAllClaims bool `yaml:"requireAllClaims"`
}

type ProxyRule struct {
	Pattern string      `yaml:"pattern"`
	Claims  []OIDCClaim `yaml:"claims"`
}

type Config struct {
	LocalPort         uint16      `yaml:"localPort"`
	Destination       string      `yaml:"destination"`
	DiscoveryUrl      string      `yaml:"discoveryUrl"`
	AccessTokenHeader string      `yaml:"accessTokenHeader"`
	ProxyRules        []ProxyRule `yaml:"proxyRules"`
	LivelinessPath    string      `yaml:"livelinessPath"`
}

func IsValidConfig(config Config) (bool, error) {
	return true, nil
}

func GetDefaultConfig() Config {
	config := Config{
		LocalPort:      8080,
		LivelinessPath: "/alive",
		// This default rule will result in authentication only
		ProxyRules: []ProxyRule{{
			Pattern: "/",
			Claims:  nil,
		}},
	}

	return config
}

func GetConfigFromFile(fileName string, config *Config) error {
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		return &Error{
			Code: ErrorInvalidConfig,
			Msg:  fmt.Sprintf("Could not open config file: %s", fileName),
			Err:  err,
		}
	}

	*config = GetDefaultConfig()

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return &Error{
			Code: ErrorInvalidConfig,
			Msg:  fmt.Sprintf("Could not parse yaml config file: %s", fileName),
			Err:  err,
		}
	}

	if valid, err := IsValidConfig(*config); !valid {
		return &Error{
			Code: ErrorInvalidConfig,
			Msg:  fmt.Sprintf("Provided configuration is not valid"),
			Err:  err,
		}
	}

	return nil
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
