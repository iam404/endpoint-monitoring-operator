package email

import (
	"fmt"

	"github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier"
)

type EmailNotifier struct {
	config *v1alpha1.EmailConfig
}

func New(config *v1alpha1.EmailConfig) (notifier.Notifier, error) {
	if config == nil || !config.Enabled {
		return nil, fmt.Errorf("email config is nil or disabled")
	}

	// Basic validation
	if config.From == "" || len(config.To) == 0 {
		return nil, fmt.Errorf("invalid email configuration: from and to fields are required")
	}

	return &EmailNotifier{config: config}, nil
}

func (e *EmailNotifier) SendAlert(status string, msg string) error {
	if !e.shouldAlert(status) {
		return nil // skip silently
	}

	// TODO: Actual SES or SMTP integration can go here
	// For now, log output
	fmt.Printf("EMAIL ALERT: Status=%s, To=%v, From=%s, Message=%s\n",
		status, e.config.To, e.config.From, msg)

	return nil
}

func (e *EmailNotifier) shouldAlert(status string) bool {
	// Placeholder: extend EmailConfig with AlertOn []string if needed
	// For now, default to failure only
	return status == "failure"
}
