package main

import (
	"fmt"
	"os"
	"strings"

	"go.yaml.in/yaml/v2"

	"vafer.org/formcha/actions"
)

type Config struct {
	Server   ServerConfig          `yaml:"server"`
	Altcha   AltchaConfig          `yaml:"altcha"`
	SMTP     actions.SMTPConfig     `yaml:"smtp"`
	Webhook  actions.WebhookConfig  `yaml:"webhook"`
	Brevo    actions.BrevoConfig    `yaml:"brevo"`
	Pushover actions.PushoverConfig `yaml:"pushover"`
	Ntfy     actions.NtfyConfig     `yaml:"ntfy"`
}

type ServerConfig struct {
	Port        string `yaml:"port"`
	IdleTimeout string `yaml:"idle_timeout"`
}

type AltchaConfig struct {
	HMACKey string `yaml:"hmac_key"`
}

func loadConfig(path string) (*Config, error) {
	cfg := &Config{}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	overrideStr(&cfg.Server.Port, "PORT")
	overrideStr(&cfg.Server.IdleTimeout, "FORMCHA_IDLE_TIMEOUT")
	overrideStr(&cfg.Altcha.HMACKey, "ALTCHA_HMAC_KEY")
	overrideStr(&cfg.SMTP.Host, "SMTP_HOST")
	overrideStr(&cfg.SMTP.Port, "SMTP_PORT")
	overrideStr(&cfg.SMTP.Username, "SMTP_USERNAME")
	overrideStr(&cfg.SMTP.Password, "SMTP_PASSWORD")
	overrideStr(&cfg.SMTP.From, "SMTP_FROM")
	overrideStr(&cfg.SMTP.To, "SMTP_TO")
	overrideStr(&cfg.Webhook.URL, "WEBHOOK_URL")
	overrideStr(&cfg.Brevo.APIKey, "BREVO_API_KEY")
	overrideStr(&cfg.Brevo.SenderName, "BREVO_SENDER_NAME")
	overrideStr(&cfg.Brevo.SenderEmail, "BREVO_SENDER_EMAIL")
	overrideStr(&cfg.Brevo.ToEmail, "BREVO_TO_EMAIL")
	overrideStr(&cfg.Brevo.ToName, "BREVO_TO_NAME")
	overrideStr(&cfg.Pushover.Token, "PUSHOVER_TOKEN")
	overrideStr(&cfg.Pushover.UserKey, "PUSHOVER_USER_KEY")
	overrideStr(&cfg.Ntfy.URL, "NTFY_URL")
	overrideStr(&cfg.Ntfy.Token, "NTFY_TOKEN")

	return cfg, nil
}

func overrideStr(field *string, envKey string) {
	if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
		*field = v
	}
}
