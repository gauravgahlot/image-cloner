package server

import (
	"net/http"

	klog "k8s.io/klog/v2"
)

func (s *server) readyz(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("OK"))
	if err != nil {
		klog.Errorf("[error]: %v", err)
	}
}
