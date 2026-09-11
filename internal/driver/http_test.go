package driver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHTTPDriver(t *testing.T) {
	if _, err := NewHTTPDriver(""); err == nil {
		t.Fatal("expected error for empty endpoint, got nil")
	}

	d, err := NewHTTPDriver("http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.GetType() != "http" {
		t.Errorf("GetType() = %q, want %q", d.GetType(), "http")
	}
	if d.GetEndpoint() != "http://example.com" {
		t.Errorf("GetEndpoint() = %q, want %q", d.GetEndpoint(), "http://example.com")
	}
}

func TestHTTPDriver_Check(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		wantSuccess bool
	}{
		{"200 OK is healthy", http.StatusOK, true},
		{"204 No Content is healthy", http.StatusNoContent, true},
		{"301 redirect target is unhealthy", http.StatusMovedPermanently, false},
		{"404 is unhealthy", http.StatusNotFound, false},
		{"500 is unhealthy", http.StatusInternalServerError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer ts.Close()

			d, err := NewHTTPDriver(ts.URL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result, err := d.Check()
			if err != nil {
				t.Fatalf("Check() returned error: %v", err)
			}
			if result.Success != tt.wantSuccess {
				t.Errorf("Success = %v, want %v (message: %s)", result.Success, tt.wantSuccess, result.Message)
			}
		})
	}
}

func TestHTTPDriver_Check_ConnectionError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := ts.URL
	ts.Close() // server is now unreachable

	d, err := NewHTTPDriver(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := d.Check()
	if err != nil {
		t.Fatalf("Check() should not return a hard error on connection failure, got: %v", err)
	}
	if result.Success {
		t.Error("expected Success = false on connection error")
	}
	if result.Error == nil {
		t.Error("expected result.Error to be set on connection failure")
	}
}
