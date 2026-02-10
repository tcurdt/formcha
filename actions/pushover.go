package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/gregdel/pushover"
)

// SendWithPushover sends push notifications via Pushover
type SendWithPushover struct {
	Token   string
	UserKey string
}

func NewSendWithPushover() *SendWithPushover {
	return &SendWithPushover{
		Token:   os.Getenv("PUSHOVER_TOKEN"),
		UserKey: os.Getenv("PUSHOVER_USER_KEY"),
	}
}

func (p *SendWithPushover) Name() string {
	return "pushover"
}

func (p *SendWithPushover) IsConfigured() bool {
	return p.Token != "" && p.UserKey != ""
}

func (p *SendWithPushover) Execute(ctx context.Context, data FormData) error {
	if !p.IsConfigured() {
		return nil
	}

	app := pushover.New(p.Token)
	recipient := pushover.NewRecipient(p.UserKey)

	body := formatFormData(data)
	message := pushover.NewMessage(body)
	message.Title = "Form Submission"

	_, err := app.SendMessage(message, recipient)
	if err != nil {
		return fmt.Errorf("pushover send failed: %w", err)
	}

	return nil
}
