package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/altcha-org/altcha-lib-go"
)

const testHMACKey = "test-secret-key"

func init() {
	altchaHMACKey = testHMACKey
}

func TestRootHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	rootHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "ALTCHA server") {
		t.Error("expected response to contain server info")
	}
}

func TestAltchaHandler_Get(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/altcha", nil)
	w := httptest.NewRecorder()

	altchaHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var challenge altcha.Challenge
	if err := json.Unmarshal(w.Body.Bytes(), &challenge); err != nil {
		t.Fatalf("failed to parse challenge JSON: %v", err)
	}

	if challenge.Challenge == "" {
		t.Error("expected challenge to have a non-empty Challenge field")
	}
	if challenge.Salt == "" {
		t.Error("expected challenge to have a non-empty Salt field")
	}
	if challenge.Signature == "" {
		t.Error("expected challenge to have a non-empty Signature field")
	}
}

func TestAltchaHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/altcha", nil)
	w := httptest.NewRecorder()

	altchaHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestSubmitHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/submit", nil)
	w := httptest.NewRecorder()

	submitHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestSubmitHandler_MissingPayload(t *testing.T) {
	form := url.Values{}
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	submitHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestSubmitHandler_InvalidPayload(t *testing.T) {
	form := url.Values{}
	form.Set("altcha", "invalid-payload")
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	submitHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestSubmitHandler_ValidPayload(t *testing.T) {
	// first, create a challenge
	challenge, err := altcha.CreateChallenge(altcha.ChallengeOptions{
		HMACKey:   testHMACKey,
		MaxNumber: 50000,
	})
	if err != nil {
		t.Fatalf("failed to create challenge: %v", err)
	}

	// solve the challenge
	solution, err := altcha.SolveChallenge(challenge.Challenge, challenge.Salt, altcha.Algorithm(challenge.Algorithm), 50000, 0, nil)
	if err != nil {
		t.Fatalf("failed to solve challenge: %v", err)
	}

	// create payload
	payload := struct {
		Algorithm string `json:"algorithm"`
		Challenge string `json:"challenge"`
		Number    int    `json:"number"`
		Salt      string `json:"salt"`
		Signature string `json:"signature"`
	}{
		Algorithm: challenge.Algorithm,
		Challenge: challenge.Challenge,
		Number:    solution.Number,
		Salt:      challenge.Salt,
		Signature: challenge.Signature,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	payloadBase64 := base64.StdEncoding.EncodeToString(payloadBytes)

	form := url.Values{}
	form.Set("altcha", payloadBase64)
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	submitHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	if response["success"] != true {
		t.Error("expected success to be true")
	}
}

func TestCorsMiddleware(t *testing.T) {
	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// test regular request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected Access-Control-Allow-Origin header to be *")
	}

	// test OPTIONS preflight request
	req = httptest.NewRequest(http.MethodOptions, "/", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for OPTIONS, got %d", w.Code)
	}
}

func TestGetPort(t *testing.T) {
	// test default port
	port := getPort()
	if port != "3000" {
		t.Errorf("expected default port 3000, got %s", port)
	}
}

func TestGetIdleTimeout_Default(t *testing.T) {
	t.Setenv("FORMCHA_IDLE_TIMEOUT", "")

	idleTimeout, err := getIdleTimeout()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if idleTimeout != 0 {
		t.Fatalf("expected %v, got %v", 0, idleTimeout)
	}
}

func TestGetIdleTimeout_ExplicitZero(t *testing.T) {
	t.Setenv("FORMCHA_IDLE_TIMEOUT", "0")

	idleTimeout, err := getIdleTimeout()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if idleTimeout != 0 {
		t.Fatalf("expected 0, got %v", idleTimeout)
	}
}

func TestGetIdleTimeout_ValidDuration(t *testing.T) {
	t.Setenv("FORMCHA_IDLE_TIMEOUT", "30s")

	idleTimeout, err := getIdleTimeout()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if idleTimeout != 30*time.Second {
		t.Fatalf("expected 30s, got %v", idleTimeout)
	}
}

func TestGetIdleTimeout_InvalidDuration(t *testing.T) {
	t.Setenv("FORMCHA_IDLE_TIMEOUT", "not-a-duration")

	_, err := getIdleTimeout()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestGetIdleTimeout_NegativeDuration(t *testing.T) {
	t.Setenv("FORMCHA_IDLE_TIMEOUT", "-1s")

	_, err := getIdleTimeout()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

func TestSpamFilterHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/submit_spam_filter", nil)
	w := httptest.NewRecorder()

	submitSpamFilterHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestSpamFilterHandler_MissingPayload(t *testing.T) {
	form := url.Values{}
	req := httptest.NewRequest(http.MethodPost, "/submit_spam_filter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	submitSpamFilterHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
