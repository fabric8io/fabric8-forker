package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testOwner = "fabric8io"
	testRepo  = "fabric8-forker"
)

func TestErrorOnMissingAuthorization(t *testing.T) {
	req, err := http.NewRequest("GET", "/fork", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(fork)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}

func TestErrorOnMissingURLQuery(t *testing.T) {
	req, err := http.NewRequest("GET", "/fork", nil)
	req.Header.Set(headerAuthorization, "Bearer Xxxxxxx")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(fork)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestExtractOwnerRepoFromHTTPSURL(t *testing.T) {
	url := "https://github.com/fabric8io/fabric8-forker.git"
	verifyURLExtraction(t, url)
}

func TestExtractOwnerRepoFromGitURL(t *testing.T) {
	url := "git@github.com:fabric8io/fabric8-forker.git"
	verifyURLExtraction(t, url)
}

func verifyURLExtraction(t *testing.T, url string) {
	o, r, err := ParseOwnerAndRepo(url)
	if err != nil {
		t.Errorf("not expecting error %v", err.Error())
	}
	if o != testOwner {
		t.Errorf("expected %s, actual %s", testOwner, o)
	}
	if r != testRepo {
		t.Errorf("expected %s, actual %s", testRepo, r)
	}
}
