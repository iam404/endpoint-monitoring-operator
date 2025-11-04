package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/common/smtpAuth"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EmailNotifier struct {
	cfg *v1alpha1.EmailConfig
}

func New(cfg *v1alpha1.EmailConfig) (notifier.Notifier, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, fmt.Errorf("email config is nil or disabled")
	}

	provider := strings.ToUpper(cfg.EmailProvider)
	if cfg.EmailSecretRef.Name == "" && (provider == "" || provider == "SMTP") {
		if provider == "" && cfg.Host == "" {
			return nil, fmt.Errorf("email config is nil or missing Host or emailProvider")
		}
		if provider == "SMTP" && cfg.Host == "" {
			return nil, fmt.Errorf("emailProvider is smtp and missing Host")
		}
		if cfg.From == "" {
			return nil, fmt.Errorf("invalid email configuration: from field is required")
		}
	}

	if len(cfg.To) == 0 {
		return nil, fmt.Errorf("invalid email configuration: to array is required")
	}

	return &EmailNotifier{cfg: cfg}, nil
}

func (e *EmailNotifier) SendAlert(status string, values *notifier.NoticeValues, client client.Client) error {
	if !e.shouldAlert(status) {
		return nil // skip silently
	}

	var msg, subject string
	if e.cfg.Subject == "" {
		subject = values.AlertMessage
	} else {
		tmpl, err := template.New("subject").Parse(e.cfg.Subject)
		if err != nil {
			return fmt.Errorf("EMAIL: Failed to parse Subject Template: %s\n", err)
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, values)
		if err != nil {
			return fmt.Errorf("EMAIL: Failed to solve Subject Template: %s\n", err)
		}
		subject = buf.String()
	}
	if e.cfg.Body == "" {
		msg = values.AlertMessage
	} else {
		tmpl, err := template.New("body").Parse(e.cfg.Body)
		if err != nil {
			return fmt.Errorf("EMAIL: Failed to parse Body Template: %s\n", err)
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, values)
		if err != nil {
			return fmt.Errorf("EMAIL: Failed to solve Body Template: %s\n", err)
		}
		msg = buf.String()
	}
	provider := strings.ToUpper(e.cfg.EmailProvider)
	if provider == "" || provider == "SMTP" {
		if err := e.SendSMTP(subject, msg, values, client); err != nil {
			return fmt.Errorf("EMAIL SMTP: Send failed, Returned %s\n", err)
		}
	}
	if provider == "SES" {
		if err := e.SendSES(subject, msg, values); err != nil {
			return fmt.Errorf("EMAIL SES: Send failed, Returned %s\n", err)
		}
	}
	//fmt.Printf("EMAIL ALERT: Status=%s, To=%v, From=%s, Message=%s\n",
	//	status, e.cfg.To, e.cfg.From, msg)

	return nil
}

func (e *EmailNotifier) shouldAlert(status string) bool {
	return notifier.ShouldAlert(e.cfg.AlertOn, status)
}

func (e *EmailNotifier) SendSMTP(subject string, msg string, values *notifier.NoticeValues, client client.Client) error {
	host := e.cfg.Host
	from := e.cfg.From

	var auth smtp.Auth = nil
	if e.cfg.EmailSecretRef.Name != "" {
		secret, err := notifier.GetSecret(e.cfg.EmailSecretRef.Name, values.Namespace, client)
		if err != nil {
			return fmt.Errorf("unable to read secret: %s\n", err)
		}

		var identity, username, password string
		identity = ""
		if identitySecret, ok := secret.Data["identity"]; ok {
			identity = string(identitySecret)
		}
		if usernameSecret, ok := secret.Data["username"]; ok {
			username = string(usernameSecret)
		}
		if passwordSecret, ok := secret.Data["password"]; ok {
			password = string(passwordSecret)
		}
		if hostSecret, ok := secret.Data["host"]; ok {
			host = string(hostSecret)
		}
		if fromSecret, ok := secret.Data["from"]; ok {
			from = string(fromSecret)
		}

		hostPart, _, _ := net.SplitHostPort(host)
		if username != "" && password != "" {
			auth = smtpAuth.CommonAuth(identity, username, password, hostPart)
		}
	}

	if host == "" {
		return fmt.Errorf("emailProvider is smtp and missing host")
	}
	if from == "" {
		return fmt.Errorf("invalid email configuration: from field is required")
	}

	body := []byte(
		"From: " + from + "\r\n" +
			"To: " + from + "\r\n" +
			"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
			"Subject: " + escapeUTF8(subject) + "\r\n\r\n" +
			strings.ReplaceAll(msg, "\n", "\r\n") + "\r\n")

	return smtp.SendMail(host, auth, from, e.cfg.To, body)
}

func (e *EmailNotifier) SendSES(subject string, msg string, values *notifier.NoticeValues) error {
	ctx := context.Background()
	sdkcfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	sesClient := sesv2.NewFromConfig(sdkcfg)

	body := strings.ReplaceAll(msg, "\n", "\r\n") + "\r\n"

	charset := "UTF-8"
	params := sesv2.SendEmailInput{
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    &subject,
					Charset: &charset,
				},
				Body: &types.Body{
					Text: &types.Content{
						Data:    &body,
						Charset: &charset,
					},
				},
			},
		},
		Destination: &types.Destination{
			ToAddresses: e.cfg.To,
		},
		FromEmailAddress: &e.cfg.From,
	}

	_, err = sesClient.SendEmail(ctx, &params)
	if err != nil {
		return err
	}

	return nil
}

func escapeUTF8(s string) string {
	for _, r := range s {
		if r > 127 || r < 32 {
			return "=?utf-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
		}
	}
	return s
}
