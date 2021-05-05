package gnet

import (
	"net/http"
	"io"
)

type FnCall func(url string, options ...Option) (status int, body io.ReadCloser, err error)

func HttpCall(url string, options ...Option) (int, io.ReadCloser, error) {
	return gnetCall(url, http_i, options...)
}

func JsonCall(url string, options ...Option) (int, io.ReadCloser, error) {
	return gnetCall(url, json_i, options...)
}

type httpFunc_i func(string,*Options)(int,[]byte,*http.Response,error)

func gnetCall(url string, fnCall httpFunc_i, options ...Option) (int, io.ReadCloser, error) {
	option := getOptions(options...)
	option.dontReadRespBody = true

	status, _, resp, err := fnCall(url, option)
	if err != nil {
		return status, nil, err
	}
	return status, resp.Body, nil
}
