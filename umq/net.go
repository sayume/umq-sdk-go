package umq

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	urlLib "net/url"
	"sort"
	"time"
)

var client *http.Client

var (
	ErrInvalidResource  = errors.New("")
	ErrInvalidInput     = errors.New("")
	ErrServerError      = errors.New("")
	ErrUnauthorizeError = errors.New("")
)

const maxRetryTime = 30 * 1000 //milisecond
const maxRetryTimes = 8
const retryBase = 200 //milisecond

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

func sendHTTPRequest(url string, method string, body io.Reader, authToken string, output interface{}, retryTimes int) (err error) {
	rqst, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}

	rqst.Header.Set("content-type", "application/json")
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
	case 203:
		return ErrUnauthorizeError
	case 404:
		return ErrInvalidResource
	case 400:
		return ErrInvalidInput
	case 500:
		if retryTimes > maxRetryTimes {
			return ErrServerError
		}
		delay := rand.Int63n(int64(200 * math.Exp2(float64(retryTimes))))
		<-time.After(time.Duration(math.Min(float64(maxRetryTime), float64(delay))) * time.Millisecond)
		retryTimes++
		return sendHTTPRequest(url, method, body, authToken, output, retryTimes)
	}
	return
}

func sendHttpRequestForAPI(url string, params map[string]string, timeout uint32, retryTimes int) (res []byte, err error) {
	req, err := urlLib.Parse(url)
	if err != nil {
		return
	}
	reqQuery := req.Query()
	for k, v := range params {
		reqQuery.Set(k, v)
	}
	req.RawQuery = reqQuery.Encode()
	client := newTimeoutHTTPClient(time.Duration(timeout) * time.Second)
	result, err := client.Get(req.String())
	if err != nil {
		return
	}
	defer result.Body.Close()
	res, err = ioutil.ReadAll(result.Body)

	switch result.StatusCode {
	case 200:
		return nil, nil
	case 203:
		return nil, ErrUnauthorizeError
	case 404:
		return nil, ErrInvalidResource
	case 400:
		return nil, ErrInvalidInput
	case 500:
		delay := rand.Int63n(int64(200 * math.Exp2(float64(retryTimes))))
		<-time.After(time.Duration(math.Min(float64(maxRetryTime), float64(delay))) * time.Millisecond)
		retryTimes++
		return sendHttpRequestForAPI(url, params, timeout, retryTimes)
	}
	return
}

func sendAPIHttpRequest(params map[string]string, privateKey string, timeout uint32) (res []byte, err error) {
	sign := signParams(params, privateKey)
	params["Signature"] = sign
	return sendHttpRequestForAPI("https://api.ucloud.cn", params, timeout, 0)
}

func sendHTTPRequestWithStringOutput(url string, method string, body io.Reader, authToken string, retryTimes int) (output string, err error) {
	rqst, err := http.NewRequest(method, url, body)
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
	var buf bytes.Buffer
	buf.ReadFrom(response.Body)
	output = buf.String()

	switch response.StatusCode {
	case 200:
		return "", nil
	case 203:
		return "", ErrUnauthorizeError
	case 404:
		return "", ErrInvalidResource
	case 400:
		return "", ErrInvalidInput
	case 500:
		delay := rand.Int63n(int64(200 * math.Exp2(float64(retryTimes))))
		<-time.After(time.Duration(math.Min(float64(maxRetryTime), float64(delay))) * time.Millisecond)
		retryTimes++
		return sendHTTPRequestWithStringOutput(url, method, body, authToken, retryTimes)
	}
	return
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
