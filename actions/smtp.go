package actions

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

// sends emails via SMTP
type SendWithSMTP struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	To       string
}

func NewSendWithSMTP() *SendWithSMTP {
	return &SendWithSMTP{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
		To:       os.Getenv("SMTP_TO"),
	}
}

func (s *SendWithSMTP) Name() string {
	return "smtp"
}

func (s *SendWithSMTP) IsConfigured() bool {
	return s.Host != "" && s.Port != "" && s.From != "" && s.To != ""
}

func (s *SendWithSMTP) Execute(ctx context.Context, data FormData) error {
	if !s.IsConfigured() {
		return nil
	}

	subject := "Form Submission"
	body := formatFormData(data)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.From, s.To, subject, body)

	addr := s.Host + ":" + s.Port

	var auth smtp.Auth
	if s.Username != "" && s.Password != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}

	// support for TLS connections (port 465)
	if s.Port == "465" {
		return s.sendWithTLS(addr, auth, msg)
	}

	return smtp.SendMail(addr, auth, s.From, []string{s.To}, []byte(msg))
}

func (s *SendWithSMTP) sendWithTLS(addr string, auth smtp.Auth, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: s.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}
	}

	if err := client.Mail(s.From); err != nil {
		return fmt.Errorf("mail from failed: %w", err)
	}

	if err := client.Rcpt(s.To); err != nil {
		return fmt.Errorf("rcpt to failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data failed: %w", err)
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	return client.Quit()
}
