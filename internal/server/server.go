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
