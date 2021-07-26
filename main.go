// Copyright 2021 The image-cloner Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
