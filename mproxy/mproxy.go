package mproxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Config struct {
	LocalPort   uint16
	Destination string
}

type MProxy struct {
	config        Config
	listenAddress string
	url           *url.URL
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
		url:           config.getURL(),
		running:       false,
	}

	return mp, ok
}

func (p *MProxy) StartProxy() int {
	ok := ErrorNoError

	http.HandleFunc("/",
		func(writer http.ResponseWriter, request *http.Request) {
			fmt.Println("headers:", request.Header)
			fmt.Println("Body:", request.Body)
			httputil.NewSingleHostReverseProxy(p.url).ServeHTTP(writer, request)
		})

	if err := http.ListenAndServe(p.config.getListenAddress(), nil); err != nil {
		ok = ErrorProxyInitError
	}

	return ok
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
	url, ok := url.Parse(c.Destination)
	if ok != nil {
		log.Fatal("Invalid destination", ok)
	}
	return url
}
