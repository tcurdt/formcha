package actions

import (
	"context"
	"fmt"
	"os"

	brevo "github.com/getbrevo/brevo-go/lib"
)

// SendWithBrevo sends emails via Brevo API
type SendWithBrevo struct {
	APIKey      string
	SenderName  string
	SenderEmail string
	ToEmail     string
	ToName      string
}

func NewSendWithBrevo() *SendWithBrevo {
	return &SendWithBrevo{
		APIKey:      os.Getenv("BREVO_API_KEY"),
		SenderName:  os.Getenv("BREVO_SENDER_NAME"),
		SenderEmail: os.Getenv("BREVO_SENDER_EMAIL"),
		ToEmail:     os.Getenv("BREVO_TO_EMAIL"),
		ToName:      os.Getenv("BREVO_TO_NAME"),
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
