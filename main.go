package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/altcha-org/altcha-lib-go"
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

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/altcha", altchaHandler)
	mux.HandleFunc("/submit", submitHandler)
	mux.HandleFunc("/submit_spam_filter", submitSpamFilterHandler)

	port := getPort()
	fmt.Printf("Server is running on port %s\n", port)
	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(
		`ALTCHA server demo endpoints:

GET /altcha - use this endpoint as challengeurl for the widget
POST /submit - use this endpoint as the form action
POST /submit_spam_filter - use this endpoint for form submissions with spam filtering`))
}

func altchaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	challenge, err := altcha.CreateChallenge(altcha.ChallengeOptions{
		HMACKey:   altchaHMACKey,
		MaxNumber: 50000,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create challenge: %s", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, challenge)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	formData := r.FormValue("altcha")
	if formData == "" {
		http.Error(w, "Altcha payload missing", http.StatusBadRequest)
		return
	}

	verified, err := altcha.VerifySolution(formData, altchaHMACKey, true)
	if err != nil || !verified {
		http.Error(w, "Invalid Altcha payload", http.StatusBadRequest)
		return
	}

	// execute actions on form submission
	allFormData, _ := formToMap(r)
	if err := actionRunner.Run(context.Background(), actions.FormData(allFormData)); err != nil {
		log.Printf("some actions failed: %v", err)
	}

	writeJSON(w, map[string]interface{}{
		"success": true,
		"data":    formData,
	})
}

func submitSpamFilterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	formData, err := formToMap(r)
	if err != nil {
		http.Error(w, "Cannot read form data", http.StatusBadRequest)
		return
	}

	payload := r.FormValue("altcha")
	if payload == "" {
		http.Error(w, "Altcha payload missing", http.StatusBadRequest)
		return
	}

	verified, verificationData, err := altcha.VerifyServerSignature(payload, altchaHMACKey)
	if err != nil || !verified {
		http.Error(w, "Invalid Altcha payload", http.StatusBadRequest)
		return
	}

	if verificationData.Verified && verificationData.Expire > time.Now().Unix() {
		if verificationData.Classification == "BAD" {
			http.Error(w, "Classified as spam", http.StatusBadRequest)
			return
		}

		if verificationData.FieldsHash != "" {
			verified, err := altcha.VerifyFieldsHash(formData, verificationData.Fields, verificationData.FieldsHash, "SHA-256")
			if err != nil || !verified {
				http.Error(w, "Invalid fields hash", http.StatusBadRequest)
				return
			}
		}

		// execute actions on form submission
		if err := actionRunner.Run(context.Background(), actions.FormData(formData)); err != nil {
			log.Printf("some actions failed: %v", err)
		}

		writeJSON(w, map[string]interface{}{
			"success":          true,
			"data":             formData,
			"verificationData": verificationData,
		})
		return
	}

	http.Error(w, "Invalid Altcha payload", http.StatusBadRequest)
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

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func formToMap(r *http.Request) (map[string][]string, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	return r.Form, nil
}
