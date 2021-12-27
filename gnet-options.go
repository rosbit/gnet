package gnet

import (
	"io"
	"os"
)

type Options struct {
	method         string
	timeout           int  // timeout in seconds to wait while connect/send/recv-ing
	dontReadRespBody bool  // if it is true, it's your resposibility to get body from http.Response.Body
	bodyLogger  io.Writer  // copy body to bodyLogger if not nil
	multiBase  *BaseUrl

	params interface{}
	headers map[string]string
	jsonCall bool

	baUser, baPasswd string
	basicAuth bool

	caCert []byte
	certPEMBlock, keyPEMBlock []byte
}

type Option func(*Options)

func BasicAuth(userName, password string) Option {
	return func(options *Options) {
		options.baUser = userName
		options.baPasswd = password
		options.basicAuth = true
	}
}

func Params(params interface{}) Option {
	return func(options *Options) {
		options.params = params
	}
}

func Headers(headers map[string]string) Option {
	return func(options *Options) {
		options.headers = headers
	}
}

func JSONCall() Option {
	return func(options *Options) {
		options.jsonCall = true
	}
}

func M(method string) Option {
	return func(option *Options) {
		option.method = method
	}
}

func WithTimeout(timeout int) Option {
	return func(options *Options) {
		options.timeout = timeout
	}
}

func DontReadRespBody() Option {
	return func(options *Options) {
		options.dontReadRespBody = true
	}
}

func BodyLogger(writer io.Writer) Option {
	return func(options *Options) {
		options.bodyLogger = writer
	}
}

func MultiBase(multiBase *BaseUrl) Option {
	return func(options *Options) {
		options.multiBase = multiBase
	}
}

func WithTLSCertFiles(certPemFile, keyPemFile string) Option {
	return func(options *Options) {
		if certPEMBlock, err := os.ReadFile(certPemFile); err == nil {
			options.certPEMBlock = certPEMBlock
		}
		if keyPEMBlock, err := os.ReadFile(keyPemFile); err == nil {
			options.keyPEMBlock = keyPEMBlock
		}
	}
}

func WithTLSCerts(certPEMBlock, keyPEMBlock []byte) Option {
	return func(options *Options) {
		options.certPEMBlock, options.keyPEMBlock = certPEMBlock, keyPEMBlock
	}
}

func WithCaCertFile(caCertFile string) Option {
	return func(options *Options) {
		if caCert, err := os.ReadFile(caCertFile); err == nil {
			options.caCert = caCert
		}
	}
}

func WithCaCert(caCert []byte) Option {
	return func(options *Options) {
		options.caCert = caCert
	}
}

const (
	connect_timeout = 5    // default seconds to wait while trying to connect
)

func getOptions(options ...Option) *Options {
	var option Options
	for _, o := range options {
		o(&option)
	}

	if option.timeout <= 0 {
			option.timeout = connect_timeout
	}

	return &option
}

