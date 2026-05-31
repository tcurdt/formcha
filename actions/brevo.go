package actions

import (
	"context"
	"fmt"

	brevo "github.com/getbrevo/brevo-go/lib"
)

type BrevoConfig struct {
	APIKey      string `yaml:"api_key"`
	SenderName  string `yaml:"sender_name"`
	SenderEmail string `yaml:"sender_email"`
	ToEmail     string `yaml:"to_email"`
	ToName      string `yaml:"to_name"`
}

// sends emails via Brevo API
type SendWithBrevo struct {
	APIKey      string
	SenderName  string
	SenderEmail string
	ToEmail     string
	ToName      string
}

func NewSendWithBrevo(cfg BrevoConfig) *SendWithBrevo {
	return &SendWithBrevo{
		APIKey:      cfg.APIKey,
		SenderName:  cfg.SenderName,
		SenderEmail: cfg.SenderEmail,
		ToEmail:     cfg.ToEmail,
		ToName:      cfg.ToName,
	}
}

func (b *SendWithBrevo) Name() string {
	return "brevo"
}

func (b *SendWithBrevo) IsConfigured() bool {
	return b.APIKey != "" && b.SenderEmail != "" && b.ToEmail != ""
}

func (b *SendWithBrevo) Execute(ctx context.Context, data FormData) error {
	if !b.IsConfigured() {
		return nil
	}

	cfg := brevo.NewConfiguration()
	cfg.AddDefaultHeader("api-key", b.APIKey)
	client := brevo.NewAPIClient(cfg)

	body := formatFormData(data)

	_, _, err := client.TransactionalEmailsApi.SendTransacEmail(ctx, brevo.SendSmtpEmail{
		Sender: &brevo.SendSmtpEmailSender{
			Name:  b.SenderName,
			Email: b.SenderEmail,
		},
		To: []brevo.SendSmtpEmailTo{
			{
				Email: b.ToEmail,
				Name:  b.ToName,
			},
		},
		Subject:     "Form Submission",
		TextContent: body,
	})

	if err != nil {
		return fmt.Errorf("brevo send failed: %w", err)
	}

	return nil
}
