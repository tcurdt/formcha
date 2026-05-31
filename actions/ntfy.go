package actions

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

type NtfyConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

// sends push notifications via ntfy
type SendWithNtfy struct {
	URL    string
	Token  string
	client *http.Client
}

func NewSendWithNtfy(cfg NtfyConfig) *SendWithNtfy {
	return &SendWithNtfy{
		URL:    cfg.URL,
		Token:  cfg.Token,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (n *SendWithNtfy) Name() string {
	return "ntfy"
}

func (n *SendWithNtfy) IsConfigured() bool {
	return n.URL != ""
}

func (n *SendWithNtfy) Execute(ctx context.Context, data FormData) error {
	if !n.IsConfigured() {
		return nil
	}

	body := formatFormData(data)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, bytes.NewBufferString(body))
	if err != nil {
		return fmt.Errorf("ntfy: failed to create request: %w", err)
	}
	req.Header.Set("Title", "Form Submission")
	if n.Token != "" {
		req.Header.Set("Authorization", "Bearer "+n.Token)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("ntfy: send failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy: unexpected status %d", resp.StatusCode)
	}

	return nil
}
