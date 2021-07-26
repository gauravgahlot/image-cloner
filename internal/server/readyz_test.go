package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadyZ(t *testing.T) {
	s := testServer(t, nil)

	req, err := http.NewRequest("GET", "/readyz", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(s.readyz).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead.", http.StatusOK, status)
	}

	assert.Equal(t, "OK", rr.Body.String(), "response body differs")
}
