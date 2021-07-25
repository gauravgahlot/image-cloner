package server

import (
	"crypto/tls"

	klog "k8s.io/klog/v2"
)

// Config defines the HTTP server
type Config struct {
	CertFile string
	KeyFile  string
	Addr     string
}

func configTLS(c Config) *tls.Config {
	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		klog.Fatalf("[error]: %v", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}
