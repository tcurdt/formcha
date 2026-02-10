package actions

import (
	"context"
	"log"
)

// LogToStdout logs form content to stdout
type LogToStdout struct{}

func NewLogToStdout() *LogToStdout {
	return &LogToStdout{}
}

func (s *LogToStdout) Name() string {
	return "stdout"
}

func (s *LogToStdout) Execute(ctx context.Context, data FormData) error {
	log.Println("form submission received:")
	for key, values := range data {
		for _, value := range values {
			log.Printf("  %s: %s\n", key, value)
		}
	}
	return nil
}
