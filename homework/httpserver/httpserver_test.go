package main

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// test healthy
func TestHealthz(t *testing.T) {
	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthz)

	atomic.StoreInt32(&healthy, UP)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code : got %v want %v",
			status, http.StatusOK)
	}

	expected := "200"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body : got %v want %v",
			rr.Body.String(), expected)
	}

}

// test header
func TestRespHeader(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("From", "garlic@example.com")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(index)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code : got %v want %v",
			status, http.StatusOK)
	}

	expected := "garlic@example.com"
	if rr.Header().Get("From") != expected {
		t.Errorf("Handler returned unexpected body : got %v want %v",
			rr.Header().Get("From"), expected)
	}

}

// test version
func TestVersion(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(index)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code : got %v want %v",
			status, http.StatusOK)
	}

	version := rr.Header().Get("Version")
	if version == "" || version == "Unknown" {
		t.Errorf("Handler returned unexpected header: got  %v  want Version",
			version)
	}
}
