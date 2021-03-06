package request

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	url2 "net/url"
	"os"
	"path"
	"strings"
	"time"
)

//var client = http.Client{
//	Transport: &http.Transport{
//		Dial: func(netw, addr string) (net.Conn, error) {
//			c, err := net.DialTimeout(netw, addr, time.Second*2) //设置建立连接超时
//			if err != nil {
//				return nil, err
//			}
//			return c, nil
//		},
//		DisableKeepAlives: false,
//	},
//}

type Response struct {
	data []byte
	code int
	err  error
}

func (resp *Response) Text() string {
	if resp.data == nil {
		return ""
	}
	return string(resp.data)
}

func (resp *Response) Json(respData interface{}) error {
	return json.Unmarshal(resp.data, respData)
}

func (resp *Response) Code() int {
	return resp.code
}

func (resp *Response) Err() error {
	return resp.err
}

func getClient(timeout time.Duration) http.Client {
	return http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, timeout) //设置建立连接超时
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			DisableKeepAlives: false,
		},
	}
}

// header 默认添加{"Content-Type": "application/json" }
func Request(method, url string, reqData interface{}, headers, params map[string]string, timeout time.Duration, username, password string) *Response {
	var req *http.Request
	var err error
	method = strings.ToUpper(method)
	methods := map[string]bool{"GET": true, "POST": true, "DELETE": true, "PUT": true, "PATCH": true, "HEAD": true}
	if _, ok := methods[method]; ok == false {
		return &Response{[]byte{}, 0, errors.New(fmt.Sprintf("http method 错误，不支持%s", method))}
	}
	if params != nil {
		if len(params) > 0 {
			var args []string
			for k, v := range params {
				args = append(args, k+"="+v)
			}
			url = url + "?" + url2.QueryEscape(strings.Join(args, "&"))
		}
	}
	if reqData != nil {
		reqData1, _ := json.Marshal(reqData)
		req, err = http.NewRequest(method, url, bytes.NewReader(reqData1))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return &Response{[]byte{}, 0, err}
	}
	client := getClient(timeout)
	if headers != nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}
	resp, err := client.Do(req)
	if err != nil {
		return &Response{[]byte{}, 0, err}
	}
	defer func() { _ = resp.Body.Close() }()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &Response{[]byte{}, 0, err}
	}
	//return &bodyData, resp.StatusCode, nil
	return &Response{bodyData, resp.StatusCode, nil}
}

func Head(url string, reqData interface{}, headers, params map[string]string, timeout time.Duration) *Response {
	return Request("head", url, reqData, headers, params, timeout, "", "")
}

func Get(url string, reqData interface{}, headers, params map[string]string, timeout time.Duration) *Response {
	return Request("get", url, reqData, headers, params, timeout, "", "")
}

func GetBaseAuth(url string, reqData interface{}, headers, params map[string]string, timeout time.Duration, username, password string) *Response {
	return Request("get", url, reqData, headers, params, timeout, username, password)
}

func Post(url string, reqData interface{}, headers, params map[string]string, timeout time.Duration) *Response {
	return Request("post", url, reqData, headers, params, timeout, "", "")
}

func PostBaseAuth(url string, reqData interface{}, headers, params map[string]string, timeout time.Duration, username, password string) *Response {
	return Request("post", url, reqData, headers, params, timeout, username, password)
}

func Put(url string, reqData interface{}, headers, params map[string]string, timeout time.Duration) *Response {
	return Request("put", url, reqData, headers, params, timeout, "", "")
}

func PutBaseAuth(url string, reqData interface{}, headers, params map[string]string, timeout time.Duration, username, password string) *Response {
	return Request("put", url, reqData, headers, params, timeout, username, password)
}

func Delete(url string, reqData interface{}, headers, params map[string]string, timeout time.Duration) *Response {
	return Request("delete", url, reqData, headers, params, timeout, "", "")
}

func DeleteBaseAuth(url string, reqData interface{}, headers, params map[string]string, timeout time.Duration, username, password string) *Response {
	return Request("delete", url, reqData, headers, params, timeout, username, password)
}

func DownloadFile(url, toDir string, timeout time.Duration) error {
	var err error
	var winSize = 1024 * 10
	var f = path.Base(url)
	var p = path.Join(toDir, f)

	client := getClient(timeout)
	req, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = req.Body.Close() }()
	reader := bufio.NewReaderSize(req.Body, winSize)
	file, err := os.Create(p)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	buff := make([]byte, winSize)
	if req.StatusCode != 200 {
		return errors.New(fmt.Sprintf("下载的文件%s不存在，下载失败", url))
	}
	for {
		nr, err := reader.Read(buff)
		if nr > 0 {
			nw, err := writer.Write(buff[0:nr])
			if err != nil {
				break
			}
			if nr != nw {
				break
			}
		}
		if err != nil {
			break
		}
	}
	return err
}
