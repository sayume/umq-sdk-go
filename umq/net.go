package umq

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	ErrInvalidResource  = errors.New("Invalid Resource")
	ErrInvalidInput     = errors.New("Invalid Arguments")
	ErrServerError      = errors.New("Server Error")
	ErrUnauthorizeError = errors.New("Unauthorized")
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

func sendHTTPRequest(url string, method string, body io.Reader, authToken string, output interface{}) (err error) {
	retryTimes := 0
	for {
		if retryTimes >= maxRetryTimes {
			return err
		}

		var rqst *http.Request
		var response *http.Response
		rqst, err = http.NewRequest(method, url, body)
		if err != nil {
			return
		}

		rqst.Header.Set("content-type", "application/json")
		rqst.Header.Set("Authorization", authToken)
		response, err = client.Do(rqst)
		if err != nil {
			return
		}
		defer response.Body.Close()

		fmt.Println(response.StatusCode)
		if response.StatusCode >= 500 {
			delay := rand.Int63n(int64(200 * math.Exp2(float64(retryTimes))))
			<-time.After(time.Duration(math.Min(float64(maxRetryTime), float64(delay))) * time.Millisecond)
			retryTimes++
			err = ErrServerError
			continue
		}

		switch response.StatusCode {
		case 200:
			err = json.NewDecoder(response.Body).Decode(output)
			if err != nil {
				return
			}
			return
		case 401:
			err = ErrUnauthorizeError
			return
		case 404:
			err = ErrInvalidResource
			return
		case 400:
			err = ErrInvalidInput
			return
		default:
			err = ErrServerError
			return
		}
	}
}

func sendHttpRequestForAPI(url string, params map[string]string, timeout uint32) (res []byte, err error) {
	retryTimes := 0
	for {
		if retryTimes >= maxRetryTimes {
			return nil, err
		}

		req, err := urlLib.Parse(url)
		if err != nil {
			return nil, err
		}
		reqQuery := req.Query()
		for k, v := range params {
			reqQuery.Set(k, v)
		}
		req.RawQuery = reqQuery.Encode()
		client := newTimeoutHTTPClient(time.Duration(timeout) * time.Second)
		result, err := client.Get(req.String())
		if err != nil {
			return nil, err
		}
		defer result.Body.Close()
		res, err = ioutil.ReadAll(result.Body)

		if result.StatusCode >= 500 {
			delay := rand.Int63n(int64(200 * math.Exp2(float64(retryTimes))))
			<-time.After(time.Duration(math.Min(float64(maxRetryTime), float64(delay))) * time.Millisecond)
			retryTimes++
			err = ErrServerError
			continue
		}
		switch result.StatusCode {
		case 200:
			return res, nil
		case 203:
			return nil, ErrUnauthorizeError
		case 404:
			return nil, ErrInvalidResource
		case 400:
			return nil, ErrInvalidInput
		default:
			return nil, ErrServerError
		}
	}
}

func sendAPIHttpRequest(params map[string]string, privateKey string, timeout uint32) (res []byte, err error) {
	sign := signParams(params, privateKey)
	params["Signature"] = sign
	return sendHttpRequestForAPI("https://api.ucloud.cn", params, timeout)
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
