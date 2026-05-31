package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_NoFile(t *testing.T) {
	cfg, err := loadConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != "" {
		t.Errorf("expected empty port, got %q", cfg.Server.Port)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := loadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	f := writeTempConfig(t, "not: valid: yaml: {{{")
	_, err := loadConfig(f)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadConfig_FileValues(t *testing.T) {
	yaml := `
server:
  port: "8080"
  idle_timeout: "5m"
altcha:
  hmac_key: "secret"
smtp:
  host: "mail.example.com"
  port: "587"
  username: "user"
  password: "pass"
  from: "from@example.com"
  to: "to@example.com"
webhook:
  url: "https://example.com/hook"
brevo:
  api_key: "brevo-key"
  sender_name: "Sender"
  sender_email: "sender@example.com"
  to_email: "to@example.com"
  to_name: "Recipient"
pushover:
  token: "ptoken"
  user_key: "puser"
ntfy:
  url: "https://ntfy.sh/topic"
  token: "ntoken"
`
	f := writeTempConfig(t, yaml)
	cfg, err := loadConfig(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("server.port: got %q", cfg.Server.Port)
	}
	if cfg.Server.IdleTimeout != "5m" {
		t.Errorf("server.idle_timeout: got %q", cfg.Server.IdleTimeout)
	}
	if cfg.Altcha.HMACKey != "secret" {
		t.Errorf("altcha.hmac_key: got %q", cfg.Altcha.HMACKey)
	}
	if cfg.SMTP.Host != "mail.example.com" {
		t.Errorf("smtp.host: got %q", cfg.SMTP.Host)
	}
	if cfg.SMTP.Port != "587" {
		t.Errorf("smtp.port: got %q", cfg.SMTP.Port)
	}
	if cfg.SMTP.Username != "user" {
		t.Errorf("smtp.username: got %q", cfg.SMTP.Username)
	}
	if cfg.SMTP.Password != "pass" {
		t.Errorf("smtp.password: got %q", cfg.SMTP.Password)
	}
	if cfg.SMTP.From != "from@example.com" {
		t.Errorf("smtp.from: got %q", cfg.SMTP.From)
	}
	if cfg.SMTP.To != "to@example.com" {
		t.Errorf("smtp.to: got %q", cfg.SMTP.To)
	}
	if cfg.Webhook.URL != "https://example.com/hook" {
		t.Errorf("webhook.url: got %q", cfg.Webhook.URL)
	}
	if cfg.Brevo.APIKey != "brevo-key" {
		t.Errorf("brevo.api_key: got %q", cfg.Brevo.APIKey)
	}
	if cfg.Brevo.SenderName != "Sender" {
		t.Errorf("brevo.sender_name: got %q", cfg.Brevo.SenderName)
	}
	if cfg.Brevo.SenderEmail != "sender@example.com" {
		t.Errorf("brevo.sender_email: got %q", cfg.Brevo.SenderEmail)
	}
	if cfg.Brevo.ToEmail != "to@example.com" {
		t.Errorf("brevo.to_email: got %q", cfg.Brevo.ToEmail)
	}
	if cfg.Brevo.ToName != "Recipient" {
		t.Errorf("brevo.to_name: got %q", cfg.Brevo.ToName)
	}
	if cfg.Pushover.Token != "ptoken" {
		t.Errorf("pushover.token: got %q", cfg.Pushover.Token)
	}
	if cfg.Pushover.UserKey != "puser" {
		t.Errorf("pushover.user_key: got %q", cfg.Pushover.UserKey)
	}
	if cfg.Ntfy.URL != "https://ntfy.sh/topic" {
		t.Errorf("ntfy.url: got %q", cfg.Ntfy.URL)
	}
	if cfg.Ntfy.Token != "ntoken" {
		t.Errorf("ntfy.token: got %q", cfg.Ntfy.Token)
	}
}

func TestLoadConfig_EnvOverridesFile(t *testing.T) {
	f := writeTempConfig(t, "smtp:\n  host: file-host\n  port: \"25\"\n")

	t.Setenv("SMTP_HOST", "env-host")
	t.Setenv("SMTP_PORT", "587")

	cfg, err := loadConfig(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.SMTP.Host != "env-host" {
		t.Errorf("expected env to override file for smtp.host, got %q", cfg.SMTP.Host)
	}
	if cfg.SMTP.Port != "587" {
		t.Errorf("expected env to override file for smtp.port, got %q", cfg.SMTP.Port)
	}
}

func TestLoadConfig_EnvOnlyNoFile(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("ALTCHA_HMAC_KEY", "env-secret")
	t.Setenv("WEBHOOK_URL", "https://example.com/hook")

	cfg, err := loadConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("expected 9090, got %q", cfg.Server.Port)
	}
	if cfg.Altcha.HMACKey != "env-secret" {
		t.Errorf("expected env-secret, got %q", cfg.Altcha.HMACKey)
	}
	if cfg.Webhook.URL != "https://example.com/hook" {
		t.Errorf("expected hook url, got %q", cfg.Webhook.URL)
	}
}

func TestLoadConfig_EnvDoesNotOverrideWithEmpty(t *testing.T) {
	f := writeTempConfig(t, "smtp:\n  host: file-host\n")

	t.Setenv("SMTP_HOST", "")

	cfg, err := loadConfig(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.SMTP.Host != "file-host" {
		t.Errorf("empty env var should not override file value, got %q", cfg.SMTP.Host)
	}
}

func TestLoadConfig_EnvTrimsWhitespace(t *testing.T) {
	t.Setenv("SMTP_HOST", "  spaced-host  ")

	cfg, err := loadConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.SMTP.Host != "spaced-host" {
		t.Errorf("expected trimmed value, got %q", cfg.SMTP.Host)
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(f, []byte(content), 0600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return f
}
