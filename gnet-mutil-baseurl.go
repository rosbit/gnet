package gnet

import (
	wr "github.com/mroth/weightedrand"
	// "path"
	"fmt"
	"io"
	"time"
	"net/http"
	"math/rand"
)

type BaseItemT struct {
	baseUrl string
	weight  uint
	lastAccessTime int64
}

func BaseItem(baseUrl string, weight ...uint) BaseItemT {
	getWeight := func() uint {
		if len(weight)>0 {
			return weight[0]
		}
		return 0
	}

	return BaseItemT {
		baseUrl: baseUrl,
		weight: getWeight(),
		lastAccessTime: time.Now().Unix(),
	}
}

type BaseUrl struct {
	baseItems []BaseItemT
	chooser *wr.Chooser
	rd *rand.Rand
	lastOKIndex int
}

func NewBaseUrl(baseItem ...BaseItemT) (b *BaseUrl, err error) {
	if len(baseItem) == 0 {
		err = fmt.Errorf("no items")
		return
	}

	b = &BaseUrl{
		baseItems: baseItem,
	}

	if err = b.caclWeights(); err != nil {
		return
	}

	b.createRandChooser()
	b.lastOKIndex = -1
	return
}

func NewBaseUrl2(baseUrl ...string) (b *BaseUrl, err error) {
	if len(baseUrl) == 0 {
		err = fmt.Errorf("no baseUrl")
		return
	}

	baseItems := make([]BaseItemT, len(baseUrl))
	for i, bu := range baseUrl {
		baseItems[i] = BaseItem(bu)
	}
	return NewBaseUrl(baseItems...)
}

func (b *BaseUrl) Http(uri string, options ...Option) (status int, content []byte, resp *http.Response, err error) {
	option := getOptions(options...)
	return b.httpCall(uri, option)
}

func (b *BaseUrl) JSON(uri string, options ...Option) (status int, content []byte, resp *http.Response, err error) {
	option := getOptions(options...)
	option.jsonCall = true
	return b.jsonCall(uri, option)
}

func (b *BaseUrl) GetWithBody(uri string, options ...Option) (status int, content []byte, resp *http.Response, err error) {
	option := getOptions(options...)
	return b.getWithBody(uri, option)
}

func (b *BaseUrl) httpCall(uri string, option *Options) (status int, content []byte, resp *http.Response, err error) {
	if len(option.method) == 0 {
		option.method = http.MethodGet
	}

	if isHttpUrl(uri) {
		var req *Request
		if req, err = newRequest(uri, option); err != nil {
			return
		}
		return req.Http(uri, option.method, option.params, option.headers)
	}

	var paramsReader io.ReadSeeker
	var header map[string]string
	if uri, option.method, paramsReader, header, err = adjustHttpArgs(uri, option.method, option.params, option.headers, option.bodyLogger); err != nil {
		return
	}

	return b.run(uri, paramsReader, header, option)
}

func (b *BaseUrl) jsonCall(uri string, option *Options) (status int, content []byte, resp *http.Response, err error) {
	if len(option.method) == 0 {
		option.method = http.MethodPost
	}

	if isHttpUrl(uri) {
		var req *Request
		if req, err = newRequest(uri, option); err != nil {
			return
		}
		return req.JSON(uri, option.method, option.params, option.headers)
	}

	var paramsReader io.ReadSeeker
	var header map[string]string
	if option.method, paramsReader, header, err = adjustJsonArgs(option.method, option.params, option.headers, option.bodyLogger); err != nil {
		return
	}

	return b.run(uri, paramsReader, header, option)
}

func (b *BaseUrl) getWithBody(uri string, option *Options) (status int, content []byte, resp *http.Response, err error) {
	if isHttpUrl(uri) {
		var req *Request
		if req, err = newRequest(uri, option); err != nil {
			return
		}
		return req.GetUsingBodyParams(uri, option.params, option.headers)
	}

	var paramsReader io.ReadSeeker
	var header map[string]string

	// using http.MethodPost to make a trick
	if _, _, paramsReader, header, err = adjustHttpArgs(uri, http.MethodPost, option.params, option.headers, option.bodyLogger); err != nil {
		return
	}

	option.method = http.MethodGet
	return b.run(uri, paramsReader, header, option)
}

func (b *BaseUrl) run(uri string, paramsReader io.ReadSeeker, header map[string]string, option *Options) (status int, content []byte, resp *http.Response, err error) {
	var req *Request
	startIdx := b.pick()
	for i:=startIdx; i<len(b.baseItems); i++ {
		url := fmt.Sprintf("%s%s", b.baseItems[i].baseUrl, uri)
		if paramsReader != nil {
			paramsReader.Seek(0, io.SeekStart)
		}
		if req, err = newRequest(url, option); err != nil {
			return
		}
		status, content, resp, err = req.run(url, option.method, paramsReader, header)
		if err == nil {
			return
		}
	}
	for i:=0; i<startIdx; i++ {
		url := fmt.Sprintf("%s%s", b.baseItems[i].baseUrl, uri)
		if paramsReader != nil {
			paramsReader.Seek(0, io.SeekStart)
		}
		if req, err = newRequest(url, option); err != nil {
			return
		}
		status, content, resp, err = req.run(url, option.method, paramsReader, header)
		if err == nil {
			return
		}
	}

	return
}

func (b *BaseUrl) pick() int {
	return b.chooser.PickSource(b.rd).(int)
}

func (b *BaseUrl) caclWeights() error {
	if !isHttpUrl(b.baseItems[0].baseUrl) {
		return fmt.Errorf("prefix of base URL %s is not http or https", b.baseItems[0].baseUrl)
	}
	allNoWeight := (b.baseItems[0].weight == 0)
	c := len(b.baseItems)

	for i:=1; i<c; i++ {
		bi := b.baseItems[i]
		if !isHttpUrl(bi.baseUrl) {
			return fmt.Errorf("prefix of base URL %s is not http or https", bi.baseUrl)
		}
		if bi.weight > 0 {
			if allNoWeight {
				return fmt.Errorf("weights before item #%d expected", i)
			}
		} else {
			if !allNoWeight {
				return fmt.Errorf("weight for item #%d(%s) expected", i, bi.baseUrl)
			}
		}
	}

	if allNoWeight {
		for i, _ := range b.baseItems {
			bi := &b.baseItems[i]
			bi.weight = 20 // any number greater than 0 is ok
		}
	}
	return nil
}

func (b *BaseUrl) createRandChooser() {
	choices := make([]wr.Choice, len(b.baseItems))
	for i, bi := range b.baseItems {
		choices[i].Item = i
		choices[i].Weight = bi.weight
	}

	b.rd = rand.New(rand.NewSource(time.Now().UnixNano()))
	b.chooser, _ = wr.NewChooser(choices...)
	fmt.Printf("b: %#v\n", b)
}
