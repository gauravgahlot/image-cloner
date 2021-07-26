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
