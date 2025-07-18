package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier"
)

type SlackNotifier struct {
	cfg *v1alpha1.SlackConfig
}

func New(config *v1alpha1.SlackConfig) (notifier.Notifier, error) {
	if config == nil || !config.Enabled || config.WebhookURL == "" {
		return nil, fmt.Errorf("invalid Slack config")
	}
	return &SlackNotifier{cfg: config}, nil
}

func (s *SlackNotifier) SendAlert(status string, msg string) error {
	if !s.shouldAlert(status) {
		return nil // silently skip
	}

	payload := map[string]string{"text": msg}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	resp, err := http.Post(s.cfg.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send slack alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 response from slack: %s", resp.Status)
	}

	return nil
}

func (s *SlackNotifier) shouldAlert(status string) bool {
	if len(s.cfg.AlertOn) == 0 {
		// Default to alert only on failure
		return status == "failure"
	}

	for _, allowed := range s.cfg.AlertOn {
		if allowed == status {
			return true
		}
	}
	return false
}
