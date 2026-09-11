package driver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "github.com/iam404/endpoint-monitoring-operator/api/v1alpha1"
)

func TestNewHTTPJSONDriver(t *testing.T) {
	check := &v1.HttpJsonCheck{JsonAssertions: map[string]string{"status": "ok"}}

	if _, err := NewHTTPJSONDriver("", check); err == nil {
		t.Error("expected error for empty endpoint")
	}
	if _, err := NewHTTPJSONDriver("http://example.com", nil); err == nil {
		t.Error("expected error for nil check config")
	}

	d, err := NewHTTPJSONDriver("http://example.com", check)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.GetType() != "http-json" {
		t.Errorf("GetType() = %q, want %q", d.GetType(), "http-json")
	}
}

func TestHTTPJSONDriver_Check(t *testing.T) {
	const body = `{"status":"ok","cluster":{"state":"green"},"count":5}`

	tests := []struct {
		name        string
		assertions  map[string]string
		body        string
		wantSuccess bool
	}{
		{
			name:        "all assertions pass",
			assertions:  map[string]string{"status": "ok", "cluster.state": "green", "count": "5"},
			body:        body,
			wantSuccess: true,
		},
		{
			name:        "nested assertion mismatch fails",
			assertions:  map[string]string{"cluster.state": "red"},
			body:        body,
			wantSuccess: false,
		},
		{
			name:        "missing key fails",
			assertions:  map[string]string{"does.not.exist": "x"},
			body:        body,
			wantSuccess: false,
		},
		{
			name:        "invalid json fails",
			assertions:  map[string]string{"status": "ok"},
			body:        `not-json`,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			}))
			defer ts.Close()

			d, err := NewHTTPJSONDriver(ts.URL, &v1.HttpJsonCheck{JsonAssertions: tt.assertions})
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

func TestResolveJSONPath(t *testing.T) {
	data := map[string]interface{}{
		"a": map[string]interface{}{
			"b": "value",
		},
		"top": "level",
	}

	tests := []struct {
		path    string
		want    interface{}
		wantOK  bool
	}{
		{"top", "level", true},
		{"a.b", "value", true},
		{"a.missing", nil, false},
		{"a.b.c", nil, false}, // traversing into a non-map
		{"missing", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got, ok := resolveJSONPath(data, tt.path)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}
