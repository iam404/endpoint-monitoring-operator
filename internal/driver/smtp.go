package driver

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	v1 "github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/common/smtpAuth"
	"github.com/LiciousTech/endpoint-monitoring-operator/internal/notifier"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SMTPDriver struct {
	endpoint string
	auth     smtp.Auth
	check    *v1.SmtpCheck
}

func NewSMTPDriver(endpoint string, check *v1.SmtpCheck, namespace string, client client.Client) (Driver, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	var auth smtp.Auth = nil
	if check.EmailSecretRef.Name != "" {
		secret, err := notifier.GetSecret(check.EmailSecretRef.Name, namespace, client)
		if err != nil {
			return nil, fmt.Errorf("unable to read secret: %s", err)
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
		hostPart, _, _ := net.SplitHostPort(endpoint)
		if username != "" && password != "" {
			auth = smtpAuth.CommonAuth(identity, username, password, hostPart)
		}
	}
	return &SMTPDriver{
		endpoint: endpoint,
		auth:     auth,
		check:    check,
	}, nil
}

func (t *SMTPDriver) Check() (*CheckResult, error) {
	start := time.Now()

	var conn net.Conn
	var sconn *smtp.Client
	var err error
	hostPart, _, _ := net.SplitHostPort(t.endpoint)
	if t.check.Tls {
		tlsConfig := &tls.Config{ServerName: hostPart}
		dialer := &net.Dialer{Timeout: 10 * time.Second}
		conn, err = tls.DialWithDialer(dialer, "tcp", t.endpoint, tlsConfig)
	} else {
		conn, err = net.DialTimeout("tcp", t.endpoint, 10*time.Second)
	}

	if err == nil {
		sconn, err = smtp.NewClient(conn, t.endpoint)
	}
	if err == nil && t.check.Helo != "" {
		err = sconn.Hello(t.check.Helo)
	}
	if err == nil && t.check.StartTls && !t.check.Tls {
		tlsConfig := &tls.Config{ServerName: hostPart}
		err = sconn.StartTLS(tlsConfig)
	}
	if err == nil && t.auth != nil {
		err = sconn.Auth(t.auth)
	}
	if err == nil && t.check.VerifyAssertion != "" {
		err = sconn.Verify(t.check.VerifyAssertion)
	}
	if err == nil && t.check.FromAssertion != "" {
		err = sconn.Mail(t.check.FromAssertion)
	}
	if err == nil && t.check.ToAssertion != "" {
		err = sconn.Rcpt(t.check.ToAssertion)
	}
	if sconn != nil {
		defer sconn.Quit()
		defer sconn.Close()
	}
	if conn != nil {
		defer conn.Close()
	}

	duration := time.Since(start)

	result := &CheckResult{
		ResponseTime: duration,
	}

	if err != nil {
		result.Success = false
		result.Error = err
		result.Message = fmt.Sprintf("SMTP check failed: %v", err)
		return result, nil
	}

	result.Success = true
	result.Message = fmt.Sprintf("SMTP check successful (response time: %v)", duration)

	return result, nil
}

func (t *SMTPDriver) GetEndpoint() string {
	return t.endpoint
}

func (t *SMTPDriver) GetType() string {
	return "smtp"
}
