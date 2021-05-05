// remove go1.16 dependency, build go1.16

package gnet

import (
	"encoding/json"
	"io"
)

func FsCall(url string, method string, options ...Option) (status int, body io.ReadCloser, err error) {
	option := getOptions(options...)
	return fsCall(url, method, option)
}

func fsCall(url string, method string, option *Options) (status int, body io.ReadCloser, err error) {
	fp := gnet_fs_i(url, method, option)
	fi, e := fp.Stat()
	if e != nil {
		err = e
		return
	}
	result := fi.Sys().(*Result)
	status, body = result.Status, result.Resp.Body
	return
}

func FsCallAndParseJSON(url string, method string, res interface{}, options ...Option) (status int, err error) {
	option := getOptions(options...)

	var body io.ReadCloser
	status, body, err = fsCall(url, method, option)
	if err != nil || body == nil {
		return
	}
	defer body.Close()

	respBody, deferFunc := bodyLogger(body, option.bodyLogger)
	defer deferFunc()

	err = json.NewDecoder(respBody).Decode(res)
	return
}
