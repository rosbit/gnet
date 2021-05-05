# gnet (an http client wrapper)

gnet is an http client package to make use of Go built-in net/http

### Usage

This package is fully go-getable. So just type `go get github.com/rosbit/gnet` to install.

```go
package main

import (
	"github.com/rosbit/gnet"
	"fmt"
)

func main() {
	params := map[string]interface{}{
		"a": "b",
		"c": 1,
	}
	headers := map[string]string{
		"X-Param": "x value",
	}

	status, _, resp, err := gnet.Http("http://yourname.com/path/to/url", gnet.DontReadRespBody())
	status, content, resp, err := gnet.Http("http://yourname.com/path/to/url")
	status, content, resp, err := gnet.Http("http://yourname.com/path/to/url", gnet.Params(params))
	status, content, resp, err := gnet.Http("http://yourname.com/path/to/url", gnet.Params(params), gnet.Headers(headers))
	/*
	// POST as request method
	status, _, resp, err := gnet.Http("http://yourname.com/path/to/url", gnet.M("post"), gnet.Params(params), gnet.DontReadRespBody())
	status, content, resp, err := gnet.Http("http://yourname.com/path/to/url", gnet.M("post"), gnet.Params(params))
	status, content, resp, err := gnet.Http("http://yourname.com/path/to/url", gnet.M("post"), gnet.Params(params), gnet.Headers(headers))
	// post body as a JSON 
	status, _, resp, err := gnet.JSON("http://yourname.com/path/to/url", gnet.Params(params), gnet.DontReadRespBody())
	status, content, resp, err := gnet.JSON("http://yourname.com/path/to/url", gnet.Params(params))
	status, content, resp, err := gnet.JSON("http://yourname.com/path/to/url", gnet.Params(params), gnet.Headers(headers))
	// post body as a JSON, even the method is GET
	status, content, resp, err := gnet.JSON("http://yourname.com/path/to/url", gnet.M("GET"), gnet.Params(params), gnet.Headers(headers))
	// request method is GET, request params as a FORM body
	status, content, resp, err := gnet.GetUsingBodyParams("http://yourname.com/path/to/url", gnet.Params(params), gnet.Headers(headers))
	*/
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("status: %d\n", status)
	fmt.Printf("reponse content: %s\n", string(content))
	respHeaders := gnet.GetHeaders(resp)
	for k, v := range respHeaders {
		fmt.Printf("%s: %s\n", k, v)
	}
}
```

### Usage as fs
```go
package main

import (
	"github.com/rosbit/gnet"
	"io"
	"os"
	"fmt"
)

func main() {
	// GET
	fp := gnet.Get("http://httpbin.org/get")
	defer fp.Close()
	io.Copy(os.Stdout, fp)

	// POST JSON
	fp2 := gnet.Post("http://httpbin.org/post", gnet.JSONCall(), gnet.Params(map[string]interface{}{"a": "b", "c": 1}))
	defer fp2.Close()
	io.Copy(os.Stdout, fp2)

	// with helper
	status, body, err := gnet.FsCall("http://httpbin.org/post", "POST", gnet.JSONCall(), gnet.Params(map[string]interface{}{"a": "b", "c": 1}))
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Printf("status: %d\n", status)
	if body != nil {
		defer body.Close()
		io.Copy(os.Stdout, body)
	}
}
```

### Usage with multi-baseurl
```go
    multiBase, err := gnet.NewBaseUrl(gnet.BaseItem("http://192.168.0.241:8088"), gnet.BaseItem("http://httpbin.org"))
    if err != nil {
         // err
    }
    status, body, _, err := multiBase.Http("/post", gorul.M(http.MethodPost), gnet.Params(params), gorul.Headers(headers))
    multiBase.JSON("/post", gnet.Params(params), gnet.Headers(headers))
    gnet.JSON("/post", gnet.MultiBase(multiBase), gnet.Params(params), gnet.Headers(headers))
```

### Status

The package is not fully tested, so be careful.

### Contribution

Pull requests are welcome! Also, if you want to discuss something send a pull request with proposal and changes.

__Convention:__ fork the repository and make changes on your fork in a feature branch.
