package go_app_push

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var MissingUrlErr = errors.New("missing url")
var HttpServerErr = errors.New("http server err")

type PushReq struct {
	Method   string
	Url      string
	Body     []byte
	Timeout  int
	User     string
	Password string
	Headers  map[string]string
	Request  *http.Request
	Client   *http.Client
}

func newPushReq() *PushReq {
	return new(PushReq)
}

func (r *PushReq) doPushRequest() (body []byte, statusCode int, header http.Header, err error) {
	if r.Method == "" {
		r.Method = "POST"
	}
	if len(r.Url) == 0 {
		err = MissingUrlErr
		return
	}
	//bodyStr, _ := url.QueryUnescape(string(r.Body))
	fmt.Printf("url.QueryUnescapeBody:%s\n\n", string(r.Body))
	if strings.ToUpper(r.Method) == "POST" {
		r.Request, err = http.NewRequest(r.Method, r.Url, bytes.NewBuffer(r.Body))
	} else if strings.ToUpper(r.Method) == "GET" {
		r.Request, err = http.NewRequest(r.Method, r.Url, nil)
	}
	if _, ok := r.Headers["Content-Type"]; !ok {
		r.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if len(r.Headers) > 0 {
		for k, v := range r.Headers {
			r.Request.Header.Set(k, v)
		}
	}
	if r.Timeout == 0 {
		r.Timeout = 30
	}
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    time.Duration(r.Timeout) * time.Second,
		DisableCompression: true,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	r.Client = &http.Client{Transport: tr}

	resp, err := r.Client.Do(r.Request)
	fmt.Printf("r.Client.resp:%v\n", resp)
	fmt.Printf("r.Client.DoErr:%v\n", err)
	if err != nil {
		return
	}
	fmt.Printf("r.Client.resp.StatusCode:%v\n\n", resp.StatusCode)
	if resp.StatusCode != 200 {
		err = HttpServerErr
		return
	}
	statusCode = resp.StatusCode
	header = resp.Header
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	fmt.Printf("utils----r.Client.resp.Body:%s\n\n", string(body))
	return
}
