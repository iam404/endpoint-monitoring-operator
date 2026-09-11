package driver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOpenSearchDriver(t *testing.T) {
	if _, err := NewOpenSearchDriver(""); err == nil {
		t.Error("expected error for empty endpoint")
	}

	d, err := NewOpenSearchDriver("http://example.com/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.GetType() != "opensearch" {
		t.Errorf("GetType() = %q, want %q", d.GetType(), "opensearch")
	}
	// trailing slash should be trimmed
	if d.GetEndpoint() != "http://example.com" {
		t.Errorf("GetEndpoint() = %q, want trailing slash trimmed", d.GetEndpoint())
	}
}

func TestOpenSearchDriver_Check(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		body        string
		wantSuccess bool
	}{
		{"green is healthy", http.StatusOK, `{"cluster_name":"c","status":"green","number_of_nodes":3,"active_shards":10}`, true},
		{"yellow is unhealthy", http.StatusOK, `{"status":"yellow","unassigned_shards":2}`, false},
		{"red is unhealthy", http.StatusOK, `{"status":"red","unassigned_shards":5}`, false},
		{"unknown status is unhealthy", http.StatusOK, `{"status":"purple"}`, false},
		{"non-200 is unhealthy", http.StatusServiceUnavailable, `{}`, false},
		{"invalid json is unhealthy", http.StatusOK, `not-json`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/_cluster/health" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer ts.Close()

			d, err := NewOpenSearchDriver(ts.URL)
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
