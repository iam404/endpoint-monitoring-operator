package driver

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	v1 "github.com/LiciousTech/endpoint-monitoring-operator/api/v1alpha1"
)

type SMTPDriver struct {
	endpoint string
	auth     smtp.Auth
	check    *v1.SmtpCheck
}

func NewSMTPDriver(endpoint string, check *v1.SmtpCheck) (Driver, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	var auth smtp.Auth = nil
	if check != nil && len(check.Username) > 0 && len(check.Password) > 0 {
		auth = smtp.PlainAuth("", check.Username, check.Password, endpoint)
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
	if t.check.Tls {
		tlsConfig := &tls.Config{ServerName: t.endpoint}
		dialer := &net.Dialer{Timeout: 10 * time.Second}
		conn, err = tls.DialWithDialer(dialer, "tcp", t.endpoint, tlsConfig)
	} else {
		conn, err = net.DialTimeout("tcp", t.endpoint, 10*time.Second)
	}

	if err == nil {
		sconn, err = smtp.NewClient(conn, t.endpoint)
	}
	if err == nil && t.check != nil && len(t.check.Helo) > 0 {
		err = sconn.Hello(t.check.Helo)
	}
	if err == nil && t.check != nil && t.check.StartTls && !t.check.Tls {
		tlsConfig := &tls.Config{ServerName: t.endpoint}
		err = sconn.StartTLS(tlsConfig)
	}
	if err == nil && t.auth != nil {
		err = sconn.Auth(t.auth)
	}
	if err == nil && t.check != nil && len(t.check.VerifyAssertion) > 0 {
		err = sconn.Verify(t.check.VerifyAssertion)
	}
	if err == nil && t.check != nil && len(t.check.FromAssertion) > 0 {
		err = sconn.Mail(t.check.FromAssertion)
	}
	if err == nil && t.check != nil && len(t.check.ToAssertion) > 0 {
		err = sconn.Rcpt(t.check.ToAssertion)
	}
	if err == nil {
		sconn.Quit()
		sconn.Close()
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

	defer conn.Close()

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
