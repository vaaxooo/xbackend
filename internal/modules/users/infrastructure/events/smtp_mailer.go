package events

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Mailer delivers plain text and/or HTML messages using SMTP.
type Mailer interface {
	Send(ctx context.Context, to string, subject string, textBody string, htmlBody string) error
}

type SMTPMailer struct {
	host     string
	port     int
	username string
	password string
	from     string
	useTLS   bool
	timeout  time.Duration
}

func NewSMTPMailer(host string, port int, username, password, from string, useTLS bool, timeout time.Duration) *SMTPMailer {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &SMTPMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		useTLS:   useTLS,
		timeout:  timeout,
	}
}

func (m *SMTPMailer) Send(ctx context.Context, to string, subject string, textBody string, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	dialer := &net.Dialer{Timeout: m.timeout}

	var conn net.Conn
	var err error
	if m.useTLS {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: m.host})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return err
	}
	defer client.Quit()

	if !m.useTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: m.host}); err != nil {
				return err
			}
		}
	}

	if m.username != "" {
		auth := smtp.PlainAuth("", m.username, m.password, m.host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(m.from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	msg := buildMessage(m.from, to, subject, textBody, htmlBody)
	if _, err := writer.Write([]byte(msg)); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}

func buildMessage(from, to, subject, textBody, htmlBody string) string {
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
	}

	if htmlBody == "" {
		headers = append(headers, "Content-Type: text/plain; charset=utf-8")
		return strings.Join(headers, "\r\n") + "\r\n\r\n" + textBody
	}

	boundary := fmt.Sprintf("=_xbackend_%d", time.Now().UnixNano())
	headers = append(headers, fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"", boundary))

	var builder strings.Builder
	builder.WriteString(strings.Join(headers, "\r\n"))
	builder.WriteString("\r\n\r\n")
	// Plain text part
	builder.WriteString("--" + boundary + "\r\n")
	builder.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	builder.WriteString(textBody)
	builder.WriteString("\r\n")
	// HTML part
	builder.WriteString("--" + boundary + "\r\n")
	builder.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
	builder.WriteString(htmlBody)
	builder.WriteString("\r\n")
	builder.WriteString("--" + boundary + "--")
	return builder.String()
}

var _ Mailer = (*SMTPMailer)(nil)
