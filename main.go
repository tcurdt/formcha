package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/altcha-org/altcha-lib-go"
	"github.com/coreos/go-systemd/v22/activation"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"vafer.org/formcha/actions"
)

var altchaHMACKey = os.Getenv("ALTCHA_HMAC_KEY")

var actionRunner *actions.Runner

func init() {
	var enabledActions []actions.Action

	actionRunner = actions.NewRunner(
		actions.NewLogToStdout(),
		actions.NewCallWebhook(),
		actions.NewSendWithSMTP(),
		actions.NewSendWithBrevo(),
		actions.NewSendWithPushover(),
	)

	actionRunner = actions.NewRunner(enabledActions...)
}

func main() {
	mux := http.NewServeMux()

	// GET /ping - health check
	// GET /metrics - prometheus metrics
	// GET /altcha - use this endpoint as challengeurl for the widget
	// POST /submit - use this endpoint as the form action
	// POST /submit_spam_filter - use this endpoint for form submissions with spam filtering

	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/altcha", altchaHandler)
	mux.HandleFunc("/submit", submitHandler)
	mux.HandleFunc("/submit_spam_filter", submitSpamFilterHandler)

	srv := &http.Server{
		Handler: corsMiddleware(mux),
	}

	// determine listener: systemd socket activation takes priority, then PORT.
	var ln net.Listener
	listeners, err := activation.Listeners()
	if err != nil {
		log.Fatalf("socket activation error: %v", err)
	}
	if len(listeners) > 0 {
		ln = listeners[0]
		log.Printf("Server listening on systemd socket")
	} else {
		port := getPort()
		ln, err = net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("listen error: %v", err)
		}
		log.Printf("Server listening on port %s", port)
	}

	// handle SIGTERM / SIGINT for graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-quit
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve error: %v", err)
	}
	log.Println("Server stopped")
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "pong")
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(`ALTCHA server`))
}

func altchaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	challenge, err := altcha.CreateChallenge(altcha.ChallengeOptions{
		HMACKey:   altchaHMACKey,
		MaxNumber: 50000,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create challenge: %s", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, challenge)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	formData := r.FormValue("altcha")
	if formData == "" {
		http.Error(w, "altcha payload missing", http.StatusBadRequest)
		return
	}

	verified, err := altcha.VerifySolution(formData, altchaHMACKey, true)
	if err != nil || !verified {
		http.Error(w, "invalid altcha payload", http.StatusBadRequest)
		return
	}

	// execute actions on form submission
	allFormData, _ := formToMap(r)
	if err := actionRunner.Run(context.Background(), actions.FormData(allFormData)); err != nil {
		log.Printf("some actions failed: %v", err)
	}

	writeJSON(w, map[string]any{
		"success": true,
		"data":    formData,
	})
}

func submitSpamFilterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	formData, err := formToMap(r)
	if err != nil {
		http.Error(w, "cannot read form data", http.StatusBadRequest)
		return
	}

	payload := r.FormValue("altcha")
	if payload == "" {
		http.Error(w, "altcha payload missing", http.StatusBadRequest)
		return
	}

	verified, verificationData, err := altcha.VerifyServerSignature(payload, altchaHMACKey)
	if err != nil || !verified {
		http.Error(w, "invalid Altcha payload", http.StatusBadRequest)
		return
	}

	if verificationData.Verified && verificationData.Expire > time.Now().Unix() {
		if verificationData.Classification == "BAD" {
			http.Error(w, "classified as spam", http.StatusBadRequest)
			return
		}

		if verificationData.FieldsHash != "" {
			verified, err := altcha.VerifyFieldsHash(formData, verificationData.Fields, verificationData.FieldsHash, "SHA-256")
			if err != nil || !verified {
				http.Error(w, "invalid fields hash", http.StatusBadRequest)
				return
			}
		}

		// execute actions on form submission
		if err := actionRunner.Run(context.Background(), actions.FormData(formData)); err != nil {
			log.Printf("some actions failed: %v", err)
		}

		writeJSON(w, map[string]any{
			"success":          true,
			"data":             formData,
			"verificationData": verificationData,
		})
		return
	}

	http.Error(w, "invalid Altcha payload", http.StatusBadRequest)
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "3000"
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
	}
}

func formToMap(r *http.Request) (map[string][]string, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	return r.Form, nil
}
