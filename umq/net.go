package umq

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	urlLib "net/url"
	"sort"
	"time"
)

var client *http.Client

func init() {
	client = newTimeoutHTTPClient(time.Duration(10) * time.Second)
}

func postHTTPRequest(url string, params map[string]string) (res []byte, err error) {
	if err != nil {
		return
	}
	bodyBuf, err := json.Marshal(params)
	if err != nil {
		return
	}
	outputBuf := bytes.NewReader(bodyBuf)
	result, err := client.Post(url, "application/json", outputBuf)
	if err != nil {
		return
	}
	defer result.Body.Close()
	res, err = ioutil.ReadAll(result.Body)
	return
}

func sendHTTPRequest(url string, params map[string]string, timeout uint32) (res []byte, err error) {
	req, err := urlLib.Parse(url)
	if err != nil {
		return
	}

	reqQuery := req.Query()
	for k, v := range params {
		reqQuery.Set(k, v)
	}
	req.RawQuery = reqQuery.Encode()
	result, err := client.Get(req.String())
	if err != nil {
		return
	}
	defer result.Body.Close()
	res, err = ioutil.ReadAll(result.Body)
	return
}

func sendAPIHttpRequest(params map[string]string, privateKey string, timeout uint32) (res []byte, err error) {
	sign := signParams(params, privateKey)
	params["Signature"] = sign
	return sendHTTPRequest("https://api.ucloud.cn", params, timeout)
}

func sendUMQAPIHttpRequest(url string, params map[string]string, privateKey string, timeout uint32) (res []byte, err error) {
	sign := signParams(params, privateKey)
	params["Signature"] = sign
	return sendHTTPRequest(url, params, timeout)
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

func signParams(params map[string]string, privateKey string) string {
	var source string
	keyList := make([]string, 0, len(params))
	for k := range params {
		keyList = append(keyList[:], k)
	}

	sort.Strings(keyList)
	for _, key := range keyList {
		source += key + params[key]
	}
	source += privateKey
	h := sha1.New()
	h.Write([]byte(source))
	sign := hex.EncodeToString(h.Sum(nil))
	return sign
}
