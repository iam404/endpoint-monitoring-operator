package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func (d *DiscordNotifier) SendAlert(status string, values *notifier.NoticeValues, client client.Client) error {
	if !d.shouldAlert(status) {
		return nil // silently skip
	}

	styledMsg := d.formatDiscordMessage(status, values)
	payload := map[string]string{"content": styledMsg}
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
	return notifier.ShouldAlert(d.cfg.AlertOn, status)
}

func (d *DiscordNotifier) formatDiscordMessage(status string, values *notifier.NoticeValues) string {
	var statusEmoji string
	switch values.Status {
	case "success":
		statusEmoji = ":white_check_mark:"
	case "failure":
		statusEmoji = ":x:"
	default:
		statusEmoji = ":information_source:"
	}

	return fmt.Sprintf(
		"%s **Endpoint Monitor Alert** %s\n\n**Status:** %s\n\n**Details:**\n```\n%s\n```",
		statusEmoji, statusEmoji, strings.ToUpper(values.Status), values.AlertMessage,
	)
}
