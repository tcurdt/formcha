package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// CallWebhook calls a webhook URL with form data
type CallWebhook struct {
	URL        string
	MaxRetries int
	client     *http.Client
}

func NewCallWebhook() *CallWebhook {
	return &CallWebhook{
		URL:        os.Getenv("WEBHOOK_URL"),
		MaxRetries: 3,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *CallWebhook) Name() string {
	return "webhook"
}

func (w *CallWebhook) IsConfigured() bool {
	return w.URL != ""
}

func (w *CallWebhook) Execute(ctx context.Context, data FormData) error {
	if !w.IsConfigured() {
		return nil
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= w.MaxRetries; attempt++ {
		if attempt > 0 {
			// exponential backoff
			delay := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := w.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		body, _ := io.ReadAll(resp.Body)
		lastErr = fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("webhook failed after %d retries: %w", w.MaxRetries, lastErr)
}
