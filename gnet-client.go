package gnet

import (
	"net/http"
	"time"
	"os"
	"fmt"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"sync"
)

const (
	maxIdleConnsPerHost = 2
	idleConnTimeout = 60 * time.Second
)

func httpClientCreator() func(timeout time.Duration) *http.Client {
	clientPool := &sync.Map{}

	return func(timeout time.Duration) *http.Client {
		if c, ok := clientPool.Load(timeout); ok {
			return c.(*http.Client)
		}

		transport := &http.Transport{
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			IdleConnTimeout:     idleConnTimeout,
		}
		c := &http.Client{Transport: transport, Timeout: timeout}
		clientPool.Store(timeout, c)
		return c
	}
}

func httpsClientCreator() func(timeout time.Duration) *http.Client {
	clientPool := &sync.Map{}

	return func(timeout time.Duration) *http.Client {
		if c, ok := clientPool.Load(timeout); ok {
			return c.(*http.Client)
		}

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			IdleConnTimeout:     idleConnTimeout,
		}

		c := &http.Client{Transport: transport, Timeout: timeout}
		clientPool.Store(timeout, c)
		return c
	}
}

func httpsClientWithCertFilesCreator() func(certPemFile, keyPemFile string, timeout time.Duration) (*http.Client, error) {
	clientPool := &sync.Map{}

	return func(certPemFile, keyPemFile string, timeout time.Duration) (*http.Client, error) {
		h := md5.New()
		fmt.Fprintf(h, "%s", certPemFile)
		fmt.Fprintf(h, "%s", keyPemFile)
		fmt.Fprintf(h, "%d", timeout)
		signature := fmt.Sprintf("%x", h.Sum(nil))

		if c, ok := clientPool.Load(signature); ok {
			return c.(*http.Client), nil
		}

		cert, err := tls.LoadX509KeyPair(certPemFile, keyPemFile)
		if err != nil {
			return nil, err
		}
		certBytes, err := os.ReadFile(certPemFile)
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
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			IdleConnTimeout:     idleConnTimeout,
		}

		c := &http.Client{Transport: transport, Timeout: timeout}
		clientPool.Store(signature, c)
		return c, nil
	}
}

func httpsClientWithCertBlocksCreator() func(caCert, certPEMBlock, keyPEMBlock []byte, timeout time.Duration) (*http.Client, error) {
	clientPool := &sync.Map{}

	return func(caCert, certPEMBlock, keyPEMBlock []byte, timeout time.Duration) (*http.Client, error) {
		h := md5.New()
		h.Write(caCert)
		h.Write(certPEMBlock)
		h.Write(keyPEMBlock)
		fmt.Fprintf(h, "%d", timeout)
		signature := fmt.Sprintf("%x", h.Sum(nil))

		if c, ok := clientPool.Load(signature); ok {
			return c.(*http.Client), nil
		}

		cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
		if err != nil {
			return nil, err
		}
		clientCertPool := x509.NewCertPool()
		if len(caCert) > 0 {
			if !clientCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("Failed to AppendCertsFromPEM")
			}
		} else {
			if !clientCertPool.AppendCertsFromPEM(certPEMBlock) {
				return nil, fmt.Errorf("Failed to AppendCertsFromPEM")
			}
		}
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            clientCertPool,
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true,
			},
			MaxIdleConnsPerHost: maxIdleConnsPerHost,
			IdleConnTimeout:     idleConnTimeout,
		}

		c := &http.Client{Transport: transport, Timeout: timeout}
		clientPool.Store(signature, c)
		return c, nil
	}
}
