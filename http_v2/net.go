package httpclient

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

func sendHttpRequest(url_path string, params map[string]interface{}, timeOut uint32) (res []byte, err error) {

	req_url, err := url.Parse(url_path)
	if err != nil {
		return
	}
	req_params := req_url.Query()
	for k, v := range params {
		req_params.Set(k, v.(string))
	}
	req_url.RawQuery = req_params.Encode()
	client := newTimeoutHTTPClient(time.Duration(timeOut) * time.Second)
	result, err := client.Get(req_url.String())
	if err != nil {
		return
	}
	defer result.Body.Close()
	res, err = ioutil.ReadAll(result.Body)
	return
}

func dialHTTPTimeout(timeOut time.Duration) func(net, addr string) (net.Conn, error) {
	return func(network, addr string) (c net.Conn, err error) {
		c, err = net.DialTimeout(network, addr, timeOut)
		if err != nil {
			return
		}
		if timeOut > 0 {
			c.SetDeadline(time.Now().Add(timeOut))
		}
		return
	}
}

func newTimeoutHTTPClient(timeOut time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: dialHTTPTimeout(timeOut),
		},
	}
}
