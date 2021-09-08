package gnet

import (
	"fmt"
	"testing"
	"net/http"
	"strings"
	"io"
	"os"
)

var (
	params = map[string]interface{}{
		"a": "b",
		"c": 1,
	}

	headers = map[string]string{
		"X-Param": "x value",
	}
)

func print_result(status int, content []byte, resp *http.Response, err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("status: %d\n", status)
	if content != nil {
		fmt.Printf("response content: %s\n", string(content))
	} else {
		defer resp.Body.Close()
		fmt.Printf("response from body: ")
		io.Copy(os.Stdout, resp.Body)
		fmt.Printf("\n")
	}
	respHeaders := GetHeaders(resp)
	for k, v := range respHeaders {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func http_test(url string, method string) {
	print_result(Http(url, M(method), Params(params), Headers(headers)))
	fmt.Printf("------------ done for Http %s with %s -------------\n", url, method)
}

func json_test(url string, method string) {
	print_result(JSON(url, M(method), Params(params), Headers(headers)))
	fmt.Printf("------------ done for JSON %s with %s -------------\n", url, method)
}

func Test_Http(t *testing.T) {
	http_test("http://httpbin.org/get",  http.MethodGet)
	http_test("http://httpbin.org/post", http.MethodPost)
}

func Test_JSON(t *testing.T) {
	json_test("http://httpbin.org/get",  http.MethodGet)
	json_test("http://httpbin.org/post", http.MethodPost)
}

func Test_MultiBase(t *testing.T) {
	multiBase, err := NewBaseUrl(BaseItem("http://192.168.0.241:8088"), BaseItem("http://httpbin.org"))
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	print_result(multiBase.Http("/get", Params(params), Headers(headers)))
	print_result(multiBase.Http("/post", M(http.MethodPost), Params(params), Headers(headers)))
	fmt.Printf("------------ done for BaseUrl::Http -------------\n")
	print_result(multiBase.JSON("/post", Params(params), Headers(headers)))
	print_result(JSON("/post", MultiBase(multiBase), Params(params), Headers(headers)))
	fmt.Printf("------------ done for BaseUrl::JSON -------------\n")
}

func Test_Reader(t *testing.T) {
	r, w := io.Pipe()
	go func() {
		w.Write([]byte(`{"a":"b","c":"d"}`))
		w.Close()
	}()
	print_result(JSON("http://httpbin.org/post", Params(r), Headers(headers), BodyLogger(os.Stderr)))
	fmt.Printf("------------ done for JSON io.Reader with POST -------------\n")
}

func Test_DontReadBody(t *testing.T) {
	r, w := io.Pipe()
	go func() {
		w.Write([]byte(`{"a":"b","c":"d"}`))
		w.Close()
	}()
	print_result(JSON("http://httpbin.org/post", Params(r), Headers(headers), DontReadRespBody()))
	fmt.Printf("------------ done for JSON io.Reader with POST (don't read response body)  -------------\n")
}

func Test_httpBuildParmas(t *testing.T) {
	s := strings.NewReader(`{"a":"b","c":"d"}`)
	if _, err := buildHttpParams(s, os.Stderr); err != nil {
		fmt.Printf("----failed to buildHttpParams: %v\n", err)
	} else {
		fmt.Printf("----buildHttpParmas ok\n")
	}
}

