package main

import (
	"crypto/tls"

	klog "k8s.io/klog/v2"
)

type config struct {
	certFile string
	keyFile  string
}

func configTLS(c config) *tls.Config {
	cert, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		klog.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}
