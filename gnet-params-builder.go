package gnet

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"fmt"
	"io"
	"bytes"
	"strings"
	"encoding/json"
)

func buildHttpParams(params interface{}) (io.ReadSeeker, error) {
	if params == nil {
		return nil, nil
	}
	switch v := params.(type) {
	case io.ReadSeeker:
		return v, nil
	default:
		param, err := buildHttpStringParams(params)
		if err != nil {
			return nil, err
		}
		if len(param) == 0 {
			return nil, nil
		}
		return strings.NewReader(param), nil
	}
}

func buildHttpStringParams(params interface{}) (string, error) {
	if params == nil {
		return "", nil
	}
	switch v := params.(type) {
	case *strings.Builder:
		return v.String(), nil
	case *bytes.Buffer:
		return v.String(), nil
	case io.WriterTo:
		b := &bytes.Buffer{}
		if _, err := v.WriteTo(b); err != nil {
			return "", err
		}
		return b.String(), nil
	case io.Reader:
		p, err := ioutil.ReadAll(v)
		return string(p), err
	case []byte:
		return string(v), nil
	case string:
		return v, nil
	case map[string]interface{}:
		u := url.Values{}
		for k, vv := range v {
			u.Set(k, fmt.Sprintf("%v", vv))
		}
		return u.Encode(), nil
	case map[string]string:
		u := url.Values{}
		for k, vv := range v {
			u.Set(k, vv)
		}
		return u.Encode(), nil
	case url.Values:
		return v.Encode(), nil
	case map[string][]string:
		return url.Values(v).Encode(), nil
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool:
		return fmt.Sprintf("%v", v), nil
	default:
		return "", fmt.Errorf("unknown type to build http params")
	}
}

func buildJsonParams(params interface{}) (io.ReadSeeker, error) {
	if params == nil {
		return strings.NewReader("null"), nil
	}

	switch v := params.(type) {
	case io.ReadSeeker:
		return v, nil
	case *strings.Builder:
		return strings.NewReader(v.String()), nil
	case *bytes.Buffer:
		return bytes.NewReader(v.Bytes()), nil
	case io.WriterTo:
		b := &bytes.Buffer{}
		if _, err := v.WriteTo(b); err != nil {
			return nil, err
		}
		return bytes.NewReader(b.Bytes()), nil
	case io.Reader:
		p, err := ioutil.ReadAll(v)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(p), nil
	case []byte:
		return bytes.NewReader(v), nil
	default:
		buf := &bytes.Buffer{}
		jsonEncoder := json.NewEncoder(buf)
		jsonEncoder.SetEscapeHTML(false)
		if err := jsonEncoder.Encode(params); err != nil {
			return nil, err
		}
		return bytes.NewReader(buf.Bytes()), nil
	}
}

func adjustHttpArgs(url, method string, params interface{}, header map[string]string) (string, string, io.ReadSeeker, map[string]string, error) {
	if len(method) == 0 {
		method = http.MethodGet
	} else {
		method = strings.ToUpper(method)
	}

	var paramsReader io.ReadSeeker

	switch method {
	case http.MethodGet, http.MethodHead:
		p, err := buildHttpStringParams(params)
		if err != nil {
			return url, method, paramsReader, header, err
		}
		if len(p) > 0 {
			deli := '?'
			if strings.Contains(url, "?") {
				deli = '&'
			}
			url = fmt.Sprintf("%s%c%s", url, deli, p)
		}
	default:
		p, err := buildHttpParams(params)
		if err != nil {
			return url, method, paramsReader, header, err
		}

		paramsReader = p
		if header == nil {
			header = make(map[string]string, 1)
		}
		header["Content-Type"] = "application/x-www-form-urlencoded"
	}
	return url, method, paramsReader, header, nil
}

func adjustJsonArgs(method string, params interface{}, header map[string]string) (string, io.ReadSeeker, map[string]string, error) {
	j, err := buildJsonParams(params)
	if err != nil {
		return method, nil, header, err
	}

	if len(method) == 0 {
		method = http.MethodPost
	} else {
		method = strings.ToUpper(method)
	}

	if header == nil {
		header = make(map[string]string, 1)
	}
	header["Content-Type"] = "application/json"
	return method, j, header, nil
}

