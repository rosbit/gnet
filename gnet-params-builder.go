package gnet

import (
	"net/http"
	"net/url"
	"fmt"
	"io"
	"bytes"
	"strings"
	"encoding/json"
)

const (
	headerContentType = "Content-Type"
	mimeURLEncoded = "application/x-www-form-urlencoded"
	mimeJSON = "application/json; charset=UTF-8"
)

func buildHttpParams(params interface{}, bodyLogger io.Writer) (io.ReadSeeker, error) {
	if params == nil {
		return nil, nil
	}
	switch v := params.(type) {
	case io.ReadSeeker:
		return v, nil
	default:
		param, err := buildHttpStringParams(params, bodyLogger)
		if err != nil {
			return nil, err
		}
		if len(param) == 0 {
			return nil, nil
		}
		return strings.NewReader(param), nil
	}
}

func buildHttpStringParams(params interface{}, bodyLogger io.Writer) (string, error) {
	var r string
	defer func() {
		if bodyLogger != nil {
			fmt.Fprintf(bodyLogger, "HTTP params: %s\n", r)
		}
	}()

	if params == nil {
		return r, nil
	}
	switch v := params.(type) {
	case *strings.Builder:
		r = v.String()
		return r, nil
	case *bytes.Buffer:
		r = v.String()
		return r, nil
	case io.WriterTo:
		b := &bytes.Buffer{}
		if _, err := v.WriteTo(b); err != nil {
			return r, err
		}
		r = b.String()
		return r, nil
	case io.Reader:
		p, err := io.ReadAll(v)
		r = string(p)
		return r, err
	case []byte:
		r = string(v)
		return r, nil
	case string:
		r = v
		return r, nil
	case map[string]interface{}:
		u := url.Values{}
		for k, vv := range v {
			u.Set(k, fmt.Sprintf("%v", vv))
		}
		r = u.Encode()
		return r, nil
	case map[string]string:
		u := url.Values{}
		for k, vv := range v {
			u.Set(k, vv)
		}
		r = u.Encode()
		return r, nil
	case url.Values:
		r = v.Encode()
		return r, nil
	case map[string][]string:
		r = url.Values(v).Encode()
		return r, nil
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool:
		r = fmt.Sprintf("%v", v)
		return r, nil
	default:
		return r, fmt.Errorf("unknown type to build http params")
	}
}

func buildJsonParams(params interface{}, bodyLogger io.Writer) (io.ReadSeeker, error) {
	var j interface{}

	defer func() {
		if bodyLogger != nil {
			fmt.Fprintf(bodyLogger, "JSON params: %s\n", j)
		}
	}()
	if params == nil {
		j = "null"
		return strings.NewReader("null"), nil
	}

	switch v := params.(type) {
	case io.ReadSeeker:
		j = "[io.ReadSeeker]"
		return v, nil
	case *strings.Builder:
		s := v.String()
		j = s
		return strings.NewReader(s), nil
	case *bytes.Buffer:
		b := v.Bytes()
		j = b
		return bytes.NewReader(b), nil
	case io.WriterTo:
		b := &bytes.Buffer{}
		if _, err := v.WriteTo(b); err != nil {
			return nil, err
		}
		bb := b.Bytes()
		j = bb
		return bytes.NewReader(bb), nil
	case io.Reader:
		p, err := io.ReadAll(v)
		if err != nil {
			return nil, err
		}
		j = p
		return bytes.NewReader(p), nil
	case []byte:
		j = v
		return bytes.NewReader(v), nil
	default:
		buf := &bytes.Buffer{}
		jsonEncoder := json.NewEncoder(buf)
		jsonEncoder.SetEscapeHTML(false)
		if err := jsonEncoder.Encode(params); err != nil {
			return nil, err
		}
		b := buf.Bytes()
		j = b
		return bytes.NewReader(b), nil
	}
}

func adjustHttpArgs(url, method string, params interface{}, header map[string]string, bodyLogger io.Writer) (string, string, io.ReadSeeker, map[string]string, error) {
	if len(method) == 0 {
		method = http.MethodGet
	} else {
		method = strings.ToUpper(method)
	}

	var paramsReader io.ReadSeeker

	switch method {
	case http.MethodGet, http.MethodHead:
		p, err := buildHttpStringParams(params, bodyLogger)
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
		p, err := buildHttpParams(params, bodyLogger)
		if err != nil {
			return url, method, paramsReader, header, err
		}

		paramsReader = p
		if header == nil {
			header = map[string]string{headerContentType: mimeURLEncoded}
		} else {
			ct := http.CanonicalHeaderKey(headerContentType)
			found := false
			for k, _ := range header {
				if http.CanonicalHeaderKey(k) == ct {
					found = true
					break
				}
			}
			if !found {
				header[ct] = mimeURLEncoded
			}
		}
	}
	return url, method, paramsReader, header, nil
}

func adjustJsonArgs(method string, params interface{}, header map[string]string, bodyLogger io.Writer) (string, io.ReadSeeker, map[string]string, error) {
	j, err := buildJsonParams(params, bodyLogger)
	if err != nil {
		return method, nil, header, err
	}

	if len(method) == 0 {
		method = http.MethodPost
	} else {
		method = strings.ToUpper(method)
	}

	if header == nil {
		header = map[string]string{headerContentType: mimeJSON}
	} else {
		ct := http.CanonicalHeaderKey(headerContentType)
		found := false
		for k, _ := range header {
			if http.CanonicalHeaderKey(k) == ct {
				found = true
				break
			}
		}
		if !found {
			header[ct] = mimeJSON
		}
	}
	return method, j, header, nil
}

