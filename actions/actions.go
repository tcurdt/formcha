package actions

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
)

type FormData map[string][]string

type Action interface {
	Execute(ctx context.Context, data FormData) error
	Name() string
}

type Runner struct {
	actions []Action
}

func NewRunner(actions ...Action) *Runner {
	return &Runner{actions: actions}
}

func (r *Runner) Run(ctx context.Context, data FormData) error {
	var errors []string
	for _, action := range r.actions {
		if err := action.Execute(ctx, data); err != nil {
			log.Printf("action %s failed: %v", action.Name(), err)
			errors = append(errors, fmt.Sprintf("%s: %v", action.Name(), err))
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("some actions failed: %s", strings.Join(errors, "; "))
	}
	return nil
}

func formatFormData(data FormData) string {
	var lines []string
	for key, values := range data {
		for _, value := range values {
			lines = append(lines, fmt.Sprintf("%s: %s", key, value))
		}
	}
	return strings.Join(lines, "\n")
}

func FormDataFromURLValues(v url.Values) FormData {
	return FormData(v)
}
