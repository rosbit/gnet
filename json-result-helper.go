package gnet

import (
	"encoding/json"
)

type FnCallJ func(url string, res interface{}, options ...Option) (status int, err error)

func HttpCallJ(url string, res interface{}, options ...Option) (int, error) {
	return callWgetJ(url, http_i, res, options...)
}

func JSONCallJ(url string, res interface{}, options ...Option) (int, error) {
	return callWgetJ(url, json_i, res, options...)
}

func callWgetJ(url string, fnCall httpFunc_i, res interface{}, options ...Option) (int, error) {
	option := getOptions(options...)
	option.dontReadRespBody = true
	status, _, resp, err := fnCall(url, option)
	if err != nil || resp.Body == nil {
		return status, err
	}
	defer resp.Body.Close()

	respBody, deferFunc := bodyLogger(resp.Body, option.bodyLogger)
	defer deferFunc()

	return status, json.NewDecoder(respBody).Decode(res)
}
