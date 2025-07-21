package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier"
)

type DiscordNotifier struct {
	cfg *v1alpha1.DiscordConfig
}

func New(config *v1alpha1.DiscordConfig) (notifier.Notifier, error) {
	if config == nil || !config.Enabled || config.WebhookURL == "" {
		return nil, fmt.Errorf("invalid Discord config")
	}
	return &DiscordNotifier{cfg: config}, nil
}

func (d *DiscordNotifier) SendAlert(status string, msg string) error {
	if !d.shouldAlert(status) {
		return nil // silently skip
	}

	payload := map[string]string{"content": msg}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal discord payload: %w", err)
	}

	resp, err := http.Post(d.cfg.WebhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send discord alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("non-2xx response from discord: %s", resp.Status)
	}

	return nil
}

func (d *DiscordNotifier) shouldAlert(status string) bool {
	if len(d.cfg.AlertOn) == 0 {
		// Default to alert only on failure
		return status == "failure"
	}

	for _, allowed := range d.cfg.AlertOn {
		if allowed == status {
			return true
		}
	}
	return false
} 