package mproxy

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
