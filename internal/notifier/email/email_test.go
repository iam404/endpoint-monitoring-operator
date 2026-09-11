package email

import (
	"testing"

	"github.com/iam404/endpoint-monitoring-operator/api/v1alpha1"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *v1alpha1.EmailConfig
		wantErr bool
	}{
		{"nil config", nil, true},
		{"disabled", &v1alpha1.EmailConfig{Enabled: false}, true},
		{"missing from", &v1alpha1.EmailConfig{Enabled: true, To: []string{"a@b.com"}}, true},
		{"missing to", &v1alpha1.EmailConfig{Enabled: true, From: "a@b.com"}, true},
		{"valid", &v1alpha1.EmailConfig{Enabled: true, From: "a@b.com", To: []string{"c@d.com"}}, false},
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

func TestEmailNotifier_SendAlert(t *testing.T) {
	n, err := New(&v1alpha1.EmailConfig{Enabled: true, From: "a@b.com", To: []string{"c@d.com"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Email is a placeholder implementation; it should never error and should
	// only "alert" on failure by default.
	if err := n.SendAlert("failure", "boom"); err != nil {
		t.Errorf("unexpected error on failure alert: %v", err)
	}
	if err := n.SendAlert("success", "ok"); err != nil {
		t.Errorf("unexpected error on success alert: %v", err)
	}
}
