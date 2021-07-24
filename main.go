package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	klog "k8s.io/klog/v2"
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

	c := config{
		certFile: certFile,
		keyFile:  keyFile,
	}

	server := http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		TLSConfig: configTLS(c),
	}

	registerHandlers()

	klog.Info("server listening at: ", port)
	err := server.ListenAndServeTLS("", "")
	if err != nil {
		klog.Fatal(err)
	}
}

func registerHandlers() {
	http.HandleFunc("/clone-image", cloneImage)
	http.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func cloneImage(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		klog.Fatal(err)
	}

	ioutil.WriteFile("/tmp/request", body, 0644)
	if err != nil {
		klog.Fatal(err)
	}
}
