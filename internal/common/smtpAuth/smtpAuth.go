package smtpAuth

import (
	"bytes"
	"errors"
	"fmt"
	"net/smtp"
	"slices"
)

type smtpAuth struct {
	identity string
	username string
	password string
	host     string
	method   string
}

// TODO: add oauth

func CommonAuth(identity string, username string, password string, host string) smtp.Auth {
	return &smtpAuth{identity: identity, username: username, password: password, host: host}
}

func (a *smtpAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// PLAIN and LOGIN require secure communications
	if !server.TLS {
		return "", nil, errors.New("insecure connection")
	}
	if server.Name != a.host {
		return "", nil, errors.New("server and auth host do not match")
	}
	if !slices.Contains(server.Auth, "PLAIN") {
		a.method = "LOGIN"
		return a.method, nil, nil
	} else {
		a.method = "PLAIN"
		resp := []byte(a.identity + "\x00" + a.username + "\x00" + a.password)
		return a.method, resp, nil
	}
}

func (a *smtpAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	if a.method == "PLAIN" {
		return nil, errors.New("unexpected server challenge during plain")
	}

	switch {
	case bytes.Equal(fromServer, []byte("Username:")):
		return []byte(a.username), nil
	case bytes.Equal(fromServer, []byte("Password:")):
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected server challenge during login: %s", fromServer)
	}
}
