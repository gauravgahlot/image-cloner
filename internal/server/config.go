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
