package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iam404/endpoint-monitoring-operator/api/v1alpha1"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *v1alpha1.SlackConfig
		wantErr bool
	}{
		{"nil config", nil, true},
		{"disabled", &v1alpha1.SlackConfig{Enabled: false, WebhookURL: "http://x"}, true},
		{"empty webhook", &v1alpha1.SlackConfig{Enabled: true, WebhookURL: ""}, true},
		{"valid", &v1alpha1.SlackConfig{Enabled: true, WebhookURL: "http://x"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSlackNotifier_SendAlert(t *testing.T) {
	var called bool
	var payload map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n, err := New(&v1alpha1.SlackConfig{
		Enabled:    true,
		WebhookURL: ts.URL,
		AlertOn:    []string{"failure"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("sends on configured status", func(t *testing.T) {
		called, payload = false, nil
		if err := n.SendAlert("failure", "boom"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("expected webhook to be called")
		}
		if payload["text"] != "boom" {
			t.Errorf("payload text = %q, want %q", payload["text"], "boom")
		}
	})

	t.Run("skips non-configured status", func(t *testing.T) {
		called = false
		if err := n.SendAlert("success", "all good"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if called {
			t.Error("expected webhook NOT to be called for success")
		}
	})
}

func TestSlackNotifier_SendAlert_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	n, err := New(&v1alpha1.SlackConfig{
		Enabled:    true,
		WebhookURL: ts.URL,
		AlertOn:    []string{"failure"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := n.SendAlert("failure", "boom"); err == nil {
		t.Error("expected error on non-200 response")
	}
}
