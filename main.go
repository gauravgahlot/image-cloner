package main

import (
	"flag"
	"fmt"

	klog "k8s.io/klog/v2"

	"github.com/gauravgahlot/image-cloner/internal/server"
)

var (
	certFile string
	keyFile  string
	port     int
)

func init() {
	flag.StringVar(&certFile, "tls-cert-file", "",
		"file containing the certificate for HTTPS.")
	flag.StringVar(&keyFile, "tls-private-key-file", "",
		"file containing the private key matching --tls-cert-file.")
	flag.IntVar(&port, "port", 443,
		"secure port that the webhook listens on")
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	c := server.Config{
		CertFile: certFile,
		KeyFile:  keyFile,
		Addr:     fmt.Sprintf(":%d", port),
	}

	server, err := server.Setup(c)
	if err != nil {
		klog.Fatalf("[error]: %v", err)
	}

	klog.Info("[info] server listening at: ", c.Addr)
	err = server.Serve()
	if err != nil {
		klog.Fatalf("[error]: %v", err)
	}
}
