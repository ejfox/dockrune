package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ejfox/dockrune/internal/config"
	"github.com/gin-gonic/gin"
)

func TestWebhookSignatureValidation(t *testing.T) {
	secret := "test-webhook-secret"

	tests := []struct {
		name      string
		payload   []byte
		signature string
		wantValid bool
	}{
		{
			name:      "valid signature",
			payload:   []byte(`{"test": "data"}`),
			signature: computeSignature([]byte(`{"test": "data"}`), secret),
			wantValid: true,
		},
		{
			name:      "invalid signature",
			payload:   []byte(`{"test": "data"}`),
			signature: "sha256=invalid",
			wantValid: false,
		},
		{
			name:      "empty signature",
			payload:   []byte(`{"test": "data"}`),
			signature: "",
			wantValid: false,
		},
		{
			name:      "wrong payload",
			payload:   []byte(`{"test": "data"}`),
			signature: computeSignature([]byte(`{"test": "wrong"}`), secret),
			wantValid: false,
		},
	}

	cfg := &config.Config{
		WebhookSecret: secret,
	}
	server := &Server{config: cfg}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := server.verifySignature(tt.payload, tt.signature)
			if valid != tt.wantValid {
				t.Errorf("verifySignature() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		WebhookPort:   8000,
		WebhookSecret: "test",
	}
	server := NewServer(cfg, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	server.healthCheck(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	expected := `{"status":"healthy"}`
	if w.Body.String() != expected {
		t.Errorf("Expected body %s, got %s", expected, w.Body.String())
	}
}

func TestGitHubWebhookHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret"
	cfg := &config.Config{
		WebhookPort:   8000,
		WebhookSecret: secret,
	}
	server := NewServer(cfg, nil, nil)

	tests := []struct {
		name       string
		payload    string
		signature  string
		eventType  string
		wantStatus int
	}{
		{
			name:       "ping event",
			payload:    `{"zen": "Design for failure."}`,
			signature:  computeSignature([]byte(`{"zen": "Design for failure."}`), secret),
			eventType:  "ping",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid signature",
			payload:    `{"test": "data"}`,
			signature:  "sha256=invalid",
			eventType:  "push",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing event type",
			payload:    `{"test": "data"}`,
			signature:  computeSignature([]byte(`{"test": "data"}`), secret),
			eventType:  "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request
			req, _ := http.NewRequest("POST", "/webhook/github", bytes.NewBufferString(tt.payload))
			req.Header.Set("X-Hub-Signature-256", tt.signature)
			if tt.eventType != "" {
				req.Header.Set("X-GitHub-Event", tt.eventType)
			}
			c.Request = req

			server.handleGitHubWebhook(c)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func computeSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
