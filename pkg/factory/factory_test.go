package factory

import (
	"errors"
	"testing"

	"github.com/iam404/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/iam404/endpoint-monitoring-operator/internal/notifier"
)

func newMonitor(driver string) *v1alpha1.EndpointMonitor {
	return &v1alpha1.EndpointMonitor{
		Spec: v1alpha1.EndpointMonitorSpec{
			Driver:        driver,
			Endpoint:      "http://example.com",
			HttpJsonCheck: &v1alpha1.HttpJsonCheck{JsonAssertions: map[string]string{"status": "ok"}},
		},
	}
}

func TestNewDriver_Selection(t *testing.T) {
	tests := []struct {
		driverType string
		wantType   string
		wantErr    bool
	}{
		{"http", "http", false},
		{"http-json", "http-json", false},
		{"tcp", "tcp", false},
		{"dns", "dns", false},
		{"ping", "ping", false},
		{"trino", "trino", false},
		{"opensearch", "opensearch", false},
		{"unknown", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.driverType, func(t *testing.T) {
			d, err := NewDriver(tt.driverType, "http://example.com", newMonitor(tt.driverType))
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewDriver(%q) error = %v, wantErr %v", tt.driverType, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if d.GetType() != tt.wantType {
				t.Errorf("driver type = %q, want %q", d.GetType(), tt.wantType)
			}
		})
	}
}

func TestNewNotifier(t *testing.T) {
	t.Run("nil config errors", func(t *testing.T) {
		if _, err := NewNotifier(nil); err == nil {
			t.Error("expected error for nil config")
		}
	})

	t.Run("no notifiers enabled errors", func(t *testing.T) {
		cfg := &v1alpha1.NotifyConfig{
			Slack: &v1alpha1.SlackConfig{Enabled: false},
		}
		if _, err := NewNotifier(cfg); err == nil {
			t.Error("expected error when no notifiers are enabled")
		}
	})

	t.Run("enabled slack builds notifier", func(t *testing.T) {
		cfg := &v1alpha1.NotifyConfig{
			Slack: &v1alpha1.SlackConfig{Enabled: true, WebhookURL: "http://x"},
		}
		n, err := NewNotifier(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n == nil {
			t.Fatal("expected a notifier")
		}
	})

	t.Run("invalid enabled slack config errors", func(t *testing.T) {
		cfg := &v1alpha1.NotifyConfig{
			Slack: &v1alpha1.SlackConfig{Enabled: true, WebhookURL: ""},
		}
		if _, err := NewNotifier(cfg); err == nil {
			t.Error("expected error for enabled slack with empty webhook")
		}
	})
}

// recordingNotifier captures calls for CompositeNotifier tests.
type recordingNotifier struct {
	calls int
	err   error
}

func (r *recordingNotifier) SendAlert(status, msg string) error {
	r.calls++
	return r.err
}

func TestCompositeNotifier_SendAlert(t *testing.T) {
	t.Run("fans out to all notifiers", func(t *testing.T) {
		a, b := &recordingNotifier{}, &recordingNotifier{}
		c := &CompositeNotifier{notifiers: []notifier.Notifier{a, b}}

		if err := c.SendAlert("failure", "boom"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.calls != 1 || b.calls != 1 {
			t.Errorf("expected each notifier called once, got a=%d b=%d", a.calls, b.calls)
		}
	})

	t.Run("aggregates errors but still calls all", func(t *testing.T) {
		a := &recordingNotifier{err: errors.New("a failed")}
		b := &recordingNotifier{}
		c := &CompositeNotifier{notifiers: []notifier.Notifier{a, b}}

		if err := c.SendAlert("failure", "boom"); err == nil {
			t.Error("expected aggregated error")
		}
		if b.calls != 1 {
			t.Errorf("expected second notifier still called despite first error, got %d", b.calls)
		}
	})
}
