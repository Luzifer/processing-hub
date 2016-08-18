package scripthelpers

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

var defaultHttpRequestSettings = map[string]interface{}{
	"method":      "GET",
	"timeout":     0,
	"contentType": "application/x-www-form-urlencoded; charset=UTF-8",
	"headers":     map[string]string{},
}

func httpRequest(u string, settings map[string]interface{}) *Response {
	for k, v := range defaultHttpRequestSettings {
		if _, ok := settings[k]; !ok {
			settings[k] = v
		}
	}

	var data io.Reader
	if d, ok := settings["data"]; ok {
		data = bytes.NewBufferString(d.(string))
	}

	req, err := http.NewRequest(settings["method"].(string), u, data)
	if err != nil {
		return &Response{Error: err.Error()}
	}

	if settings["username"] != nil && settings["password"] != nil {
		req.SetBasicAuth(settings["username"].(string), settings["password"].(string))
	}

	ctx := context.Background()
	if settings["timeout"].(int) != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(settings["timeout"].(int))*time.Second)
		defer cancel()
	}

	req.Header.Set("Content-Type", settings["contentType"].(string))

	for k, v := range settings["headers"].(map[string]string) {
		req.Header.Set(k, v)
	}

	resp, err := ctxhttp.Do(ctx, http.DefaultClient, req)
	if err != nil {
		return &Response{Error: err.Error()}
	}
	defer resp.Body.Close()

	rawResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &Response{Error: err.Error()}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       string(rawResponse),
		Error:      "",
	}
}

type Response struct {
	StatusCode int
	Body       string
	Error      string
}
