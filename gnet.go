// http client implementation

package gnet

import (
	"net/http"
	"strings"
	"io/ioutil"
	"time"
	"fmt"
	"os"
	"io"
)

type HttpFunc func(url string, options ...Option)(int,[]byte,*http.Response,error)

func Http(url string, options ...Option) (status int, content []byte, resp *http.Response, err error) {
	option := getOptions(options...)
	return http_i(url, option)
}

func JSON(url string, options ...Option) (status int, content []byte, resp *http.Response, err error) {
	option := getOptions(options...)
	option.jsonCall = true
	return json_i(url, option)
}

func GetUsingBodyParams(url string, options ...Option) (status int, content []byte, resp *http.Response, err error) {
	option := getOptions(options...)
	if !isHttpUrl(url) && option.multiBase != nil {
		return option.multiBase.getWithBody(url, option)
	}
	return newRequest(url, option).GetUsingBodyParams(url, option.params, option.headers)
}

func GetStatus(resp *http.Response) (int, string) {
	return resp.StatusCode, resp.Status
}

func GetHeaders(resp *http.Response) map[string]string {
	res := make(map[string]string, len(resp.Header))
	for k, v := range resp.Header {
		if v == nil || len(v) == 0 {
			res[k] = ""
		} else {
			res[k] = v[0]
		}
	}
	return res
}

func GetLastModified(resp *http.Response) (time.Time, error) {
	if resp == nil {
		return time.Time{}, fmt.Errorf("no response given")
	}
	if lastModified, ok := resp.Header["Last-Modified"]; ok {
		return time.Parse(time.RFC1123, lastModified[0])
	}
	return time.Time{}, fmt.Errorf("no response header Last-Modified")
}

func ModTime(rawurl string) (time.Time, error) {
	if isHttpUrl(rawurl) {
		option := getOptions()
		_, _, resp, err := newRequest(rawurl, option).Http(rawurl, http.MethodHead, nil, nil)
		if err != nil {
			return time.Time{}, err
		}
		return GetLastModified(resp)
	} else {
		st, e := os.Stat(rawurl)
		if e != nil {
			return time.Time{}, e
		}
		return st.ModTime(), nil
	}
}

func (gu *Request) Http(url, method string, params interface{}, header map[string]string) (status int, content []byte, resp *http.Response, err error) {
	var paramsReader io.ReadSeeker
	if url, method, paramsReader, header, err = adjustHttpArgs(url, method, params, header); err != nil {
		return
	}
	return gu.run(url, method, paramsReader, header)
}

func (gu *Request) JSON(url, method string, params interface{}, header map[string]string) (status int, content []byte, resp *http.Response, err error) {
	var paramsReader io.ReadSeeker
	if method, paramsReader, header, err = adjustJsonArgs(method, params, header); err != nil {
		return
	}
	return gu.run(url, method, paramsReader, header)
}

func (gu *Request) GetUsingBodyParams(url string, params interface{}, header map[string]string) (status int, content []byte, resp *http.Response, err error) {
	var paramsReader io.ReadSeeker
	// using http.MethodPost to make a trick
	if _, _, paramsReader, header, err = adjustHttpArgs(url, http.MethodPost, params, header); err != nil {
		return
	}
	return gu.run(url, http.MethodGet, paramsReader, header)
}

func isHttpUrl(rawurl string) bool {
	return (strings.Index(rawurl, "http://") == 0) || (strings.Index(rawurl, "https://") == 0)
}

func http_i(url string, option *Options) (status int, content []byte, resp *http.Response, err error) {
	if option.jsonCall {
		return json_i(url, option)
	}

	if len(option.method) == 0 {
		option.method = http.MethodGet
	}

	if !isHttpUrl(url) && option.multiBase != nil {
		return option.multiBase.httpCall(url, option)
	}
	return newRequest(url, option).Http(url, option.method, option.params, option.headers)
}

func json_i(url string, option *Options) (status int, content []byte, resp *http.Response, err error) {
	if len(option.method) == 0 {
		option.method = http.MethodPost
	}
	if !isHttpUrl(url) && option.multiBase != nil {
		return option.multiBase.jsonCall(url, option)
	}
	return newRequest(url, option).JSON(url, option.method, option.params, option.headers)
}

func (gu *Request) run(url, method string, params io.Reader, header map[string]string) (int, []byte, *http.Response, error) {
	var req *http.Request
	var err error
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		if req, err = http.NewRequest(method, url, params); err != nil {
			return http.StatusBadRequest, nil, nil, err
		}
	default:
		return http.StatusMethodNotAllowed, nil, nil, fmt.Errorf("method %s not supported", method)
	}

	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	resp, err := gu.client.Do(req)
	if err != nil {
		return http.StatusInternalServerError, nil, nil, err
	}

	if gu.options.dontReadRespBody {
		return resp.StatusCode, nil, resp, nil
	}

	defer resp.Body.Close()

	respBody, deferFunc := bodyLogger(resp.Body, gu.options.bodyLogger)
	defer deferFunc()

	if body, err := ioutil.ReadAll(respBody); err != nil {
		return resp.StatusCode, nil, nil, err
	} else {
		return resp.StatusCode, body, resp, nil
	}
}
