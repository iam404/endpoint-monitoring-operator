package driver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTrinoDriver(t *testing.T) {
	if _, err := NewTrinoDriver(""); err == nil {
		t.Error("expected error for empty endpoint")
	}

	d, err := NewTrinoDriver("http://example.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.GetType() != "trino" {
		t.Errorf("GetType() = %q, want %q", d.GetType(), "trino")
	}
	if d.GetEndpoint() != "http://example.com" {
		t.Errorf("GetEndpoint() = %q, want trailing slash trimmed", d.GetEndpoint())
	}
}

func TestTrinoDriver_Check(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		body        string
		wantSuccess bool
	}{
		{"coordinator and started is healthy", http.StatusOK, `{"coordinator":true,"starting":false,"uptime":"1h","nodeVersion":{"version":"400"},"environment":"prod"}`, true},
		{"still starting is unhealthy", http.StatusOK, `{"coordinator":true,"starting":true}`, false},
		{"non-coordinator is unhealthy", http.StatusOK, `{"coordinator":false,"starting":false}`, false},
		{"non-200 is unhealthy", http.StatusBadGateway, `{}`, false},
		{"invalid json is unhealthy", http.StatusOK, `not-json`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/info" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer ts.Close()

			d, err := NewTrinoDriver(ts.URL)
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
