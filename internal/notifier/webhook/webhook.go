package webhook

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WebhookNotifier struct {
	cfg *v1alpha1.WebhookConfig
}

func New(config *v1alpha1.WebhookConfig) (notifier.Notifier, error) {
	if config == nil || !config.Enabled || config.WebhookURL == "" {
		return nil, fmt.Errorf("invalid Webhook config")
	}
	return &WebhookNotifier{cfg: config}, nil
}

func (w *WebhookNotifier) SendAlert(status string, values *notifier.NoticeValues, client client.Client) error {
	if !w.shouldAlert(status) {
		return nil // silently skip
	}

	contentType := "application/json"
	if w.cfg.ContentType != "" {
		contentType = w.cfg.ContentType
	}

	var req *http.Request
	if w.cfg.Contents != "" {
		tmpl, err := template.New("contents").Parse(w.cfg.Contents)
		if err != nil {
			fmt.Printf("WEBHOOK: Failed to parse Contents Template: %s\n", err)
			return nil
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, values)
		if err != nil {
			fmt.Printf("WEBHOOK: Failed to solve Contents Template: %s\n", err)
			return nil
		}

		req, err = http.NewRequest(http.MethodPost, w.cfg.WebhookURL, io.Reader(&buf))
		if err != nil {
			return fmt.Errorf("WEBHOOK: failed to create POST webhook: %w", err)
		}
		req.Header.Add("Content-Type", contentType)
	} else {
		var err error
		req, err = http.NewRequest(http.MethodGet, w.cfg.WebhookURL, nil)
		if err != nil {
			return fmt.Errorf("WEBHOOK: failed to GET webhook alert: %w", err)
		}
	}

	if w.cfg.Authorization.Name != "" {
		secret, err := notifier.GetSecret(w.cfg.Authorization.Name, values.Namespace, client)
		if err == nil {
			if auth, ok := secret.Data["Raw"]; ok {
				req.Header.Add("Authorization", string(auth))
			}
			if auth, ok := secret.Data["Basic"]; ok {
				req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString(auth))
			}
			if auth, ok := secret.Data["Bearer"]; ok {
				req.Header.Add("Authorization", "Bearer "+string(auth))
			}
		}
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("WEBHOOK: failed to make http webhook request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("WEBHOOK: non-200 response from webhook: %s", resp.Status)
	}

	return nil
}

func (w *WebhookNotifier) shouldAlert(status string) bool {
	return notifier.ShouldAlert(w.cfg.AlertOn, status)
}
