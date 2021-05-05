package gnet

import (
	"net/http"
	"io/ioutil"
	"strings"
	"time"
	"fmt"
	"crypto/tls"
	"crypto/x509"
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
	cert, err := tls.LoadX509KeyPair(certPemFile, keyPemFile)
	if err != nil {
		return nil, err
	}
	certBytes, err := ioutil.ReadFile(certPemFile)
	if err != nil {
		return nil, err
	}
	clientCertPool := x509.NewCertPool()
	if !clientCertPool.AppendCertsFromPEM(certBytes) {
		return nil, fmt.Errorf("Failed to AppendCertsFromPEM")
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:            clientCertPool,
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		},
	}
	return &Request{client: &http.Client{Transport: transport, Timeout: time.Duration(option.timeout)*time.Second}, options: option}, nil
}

func newRequest(url string, option *Options) *Request {
	if strings.Index(url, "https://") == 0 {
		return newHttpsRequest(option)
	} else {
		return newHttpRequest(option)
	}
}

func newHttpRequest(option *Options) *Request {
	return &Request{client: &http.Client{Timeout: time.Duration(option.timeout)*time.Second}, options: option}
}

func newHttpsRequest(option *Options) *Request {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &Request{client: &http.Client{Transport: transport, Timeout: time.Duration(option.timeout)*time.Second}, options: option}
}

