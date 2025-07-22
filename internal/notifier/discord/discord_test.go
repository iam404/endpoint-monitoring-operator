package discord

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
)

type alertPayload struct {
	Content string `json:"content"`
}

func TestDiscordNotifier_SendAlert(t *testing.T) {
	var receivedPayload alertPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Errorf("failed to decode payload: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	n := &DiscordNotifier{cfg: &v1alpha1.DiscordConfig{
		Enabled:    true,
		WebhookURL: ts.URL,
		AlertOn:    []string{"failure"},
	}}

	t.Run("should send alert on failure", func(t *testing.T) {
		err := n.SendAlert("failure", "test message")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		// Update: Expect the styled message format
		expected := n.formatDiscordMessage("failure", "test message")
		if receivedPayload.Content != expected {
			t.Errorf("expected message '%s', got '%s'", expected, receivedPayload.Content)
		}
	})

	t.Run("should not send alert on success if not configured", func(t *testing.T) {
		err := n.SendAlert("success", "should not send")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if receivedPayload.Content == "should not send" {
			t.Errorf("should not have sent alert for success")
		}
	})
}

func TestDiscordNotifier_shouldAlert(t *testing.T) {
	n := &DiscordNotifier{cfg: &v1alpha1.DiscordConfig{AlertOn: []string{"failure"}}}
	if !n.shouldAlert("failure") {
		t.Error("should alert on failure")
	}
	if n.shouldAlert("success") {
		t.Error("should not alert on success")
	}

	n.cfg.AlertOn = nil
	if !n.shouldAlert("failure") {
		t.Error("should default to alert on failure")
	}
	if n.shouldAlert("success") {
		t.Error("should not alert on success by default")
	}
}

func TestDiscordNotifier_formatDiscordMessage(t *testing.T) {
	n := &DiscordNotifier{cfg: &v1alpha1.DiscordConfig{}}

	t.Run("failure status", func(t *testing.T) {
		msg := n.formatDiscordMessage("failure", "something went wrong")
		if want := ":x: **Endpoint Monitor Alert** :x:\n\n**Status:** FAILURE\n\n**Details:**\n```\nsomething went wrong\n```"; msg != want {
			t.Errorf("unexpected format for failure: got %q, want %q", msg, want)
		}
	})

	t.Run("success status", func(t *testing.T) {
		msg := n.formatDiscordMessage("success", "all good")
		if want := ":white_check_mark: **Endpoint Monitor Alert** :white_check_mark:\n\n**Status:** SUCCESS\n\n**Details:**\n```\nall good\n```"; msg != want {
			t.Errorf("unexpected format for success: got %q, want %q", msg, want)
		}
	})

	t.Run("other status", func(t *testing.T) {
		msg := n.formatDiscordMessage("info", "misc info")
		if want := ":information_source: **Endpoint Monitor Alert** :information_source:\n\n**Status:** INFO\n\n**Details:**\n```\nmisc info\n```"; msg != want {
			t.Errorf("unexpected format for info: got %q, want %q", msg, want)
		}
	})
}
