package umq

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"sort"
	"time"
)

var client *http.Client

func init() {
	client = newTimeoutHTTPClient(time.Duration(10) * time.Second)
}

func makeJsonReader(body interface{}) (io.Reader, error) {
	reader, writer := io.Pipe()
	err := json.NewEncoder(writer).Encode(body)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func sendHTTPRequest(url string, method string, body io.Reader, authToken string, output interface{}) (err error) {
	var reader io.Reader
	var writer io.Writer
	rqst, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}

	rqst.Header.Set("content-length", "application/json")
	rqst.Header.Set("Authorization", authToken)
	response, err := client.Do(rqst)
	if err != nil {
		return
	}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(output)
	if err != nil {
		return
	}

	switch response.StatusCode {
	case 200:
		return nil
	case 404:
		return ErrInvalidResource
	case 400:
		return ErrInvalidInput
	case 500:
		return ErrServerError
	}
	return
}

func sendAPIHttpRequest(params map[string]string, privateKey string, timeout uint32) (res []byte, err error) {
	return nil, nil
}

func sendUMQAPIHttpRequest(url string, params map[string]string, privateKey string, timeout uint32) (res []byte, err error) {
	return nil, nil
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
