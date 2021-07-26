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
	"net/http"
	"os"

	"github.com/gauravgahlot/image-cloner/internal/docker"
)

// Server defines the basic operations for image-cloner server.
type Server interface {
	Serve() error
}

type server struct {
	httpServer http.Server

	client       docker.Client
	registryUser string
	registry     string
}

// Setup initializes and returns a server; error otherwise.
func Setup(cfg Config) (Server, error) {
	client, err := docker.CreateClient()
	if err != nil {
		return nil, err
	}

	s := server{
		client:       client,
		registryUser: docker.RegistryUser(),
		registry:     os.Getenv("REGISTRY"),
		httpServer: http.Server{
			Addr:      cfg.Addr,
			TLSConfig: configTLS(cfg),
		},
	}

	http.HandleFunc("/readyz", s.readyz)
	http.HandleFunc("/clone-image", s.cloneImage)

	return &s, nil
}

func (s *server) Serve() error {
	return s.httpServer.ListenAndServeTLS("", "")
}
