package gnet

import (
	"net/http"
	"strings"
)

var (
	getHttpClient = httpClientCreator()
	getHttpsClient = httpsClientCreator()
	getHttpsClientWithCertFiles = httpsClientWithCertFilesCreator()
	getHttpsClientWithCertBlocks = httpsClientWithCertBlocksCreator()
)

type Request struct {
	client  *http.Client
	options *Options
}

func NewRequest(options ...Option) *Request {
	option := getOptions(options...)
	return newHttpRequest(option)
}

func NewHttpsRequest(options ...Option) *Request {
	option := getOptions(options...)
	return newHttpsRequest(option)
}

func NewHttpsRequestWithCerts(certPemFile, keyPemFile string, options ...Option) (*Request, error) {
	option := getOptions(options...)
	client, err := getHttpsClientWithCertFiles(certPemFile, keyPemFile, option.timeout)
	if err != nil {
		return nil, err
	}
	return &Request{client: client, options: option}, nil
}

func newRequest(url string, option *Options) (*Request, error) {
	if strings.Index(url, "https://") == 0 {
		if len(option.certPEMBlock) > 0 && len(option.keyPEMBlock) > 0 {
			return newHttpsRequestWithCerts(option)
		}
		return newHttpsRequest(option), nil
	} else {
		return newHttpRequest(option), nil
	}
}

func newHttpRequest(option *Options) *Request {
	client := getHttpClient(option.timeout)
	return &Request{client: client, options: option}
}

func newHttpsRequest(option *Options) *Request {
	client := getHttpsClient(option.timeout)
	return &Request{client: client, options: option}
}

func newHttpsRequestWithCerts(option *Options) (*Request, error) {
	client, err := getHttpsClientWithCertBlocks(option.caCert, option.certPEMBlock, option.keyPEMBlock, option.timeout)
	if err != nil {
		return nil, err
	}
	return &Request{client: client, options: option}, nil
}

func (g *Request) GetClient() *http.Client {
	return g.client
}
