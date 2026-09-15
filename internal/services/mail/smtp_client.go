package mail

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPClient sends HTML emails via net/smtp. net/smtp itself has no context
// support (it predates context.Context) - ctx is only checked up front so a
// caller whose request was already cancelled doesn't still trigger a send.
type SMTPClient struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func NewSMTPClient(host, port, username, password, from string) *SMTPClient {
	return &SMTPClient{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func (c *SMTPClient) Send(ctx context.Context, to, subject, htmlBody string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%s", c.Host, c.Port)
	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", c.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	// smtp.SendMail upgrades to STARTTLS itself when the server advertises
	// it (which smtp.gmail.com:587 does) - no manual TLS dance needed here.
	return smtp.SendMail(addr, auth, c.From, []string{to}, []byte(msg.String()))
}
