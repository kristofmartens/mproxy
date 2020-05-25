package mproxy

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

func GetConfig() Config {
	cfg := flag.String("config", "", "Path to config file defining authorization rules")
	flag.Parse()

	if len(*cfg) == 0 {
		fmt.Println("No configuration provided")
		os.Exit(ErrorInvalidConfig)
	}

	config := GetDefaultConfig()

	err := GetConfigFromFile(*cfg, &config)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(err.(*Error).Code)
	}

	return config
}

func httpGetJson(url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, &Error{
			Code: ErrorHttpError,
			Msg:  fmt.Sprintf("Could not access url: %v", url),
			Err:  err,
		}
	}

	defer resp.Body.Close()

	var output map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&output)

	return output, nil
}
