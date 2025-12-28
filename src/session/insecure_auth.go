package session

import (
	"errors"
	netsmtp "net/smtp"
)

// insecurePlainAuth is like smtp.PlainAuth but doesn't enforce TLS for non-localhost servers
type insecurePlainAuth struct {
	identity, username, password, host string
}

func (a *insecurePlainAuth) Start(server *netsmtp.ServerInfo) (string, []byte, error) {
	// SKIP TLS CHECK: if !server.TLS && !isLocalhost(server.Name) { ... }
	resp := []byte(a.identity + "\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *insecurePlainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
}
