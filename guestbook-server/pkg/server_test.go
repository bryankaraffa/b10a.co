package guestbook_server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	config := &Config{
		Port:                    "8080",
		AkismetAPIKey:           "test-key",
		AkismetSiteURL:          "https://example.com",
		RecaptchaSecretKey:      "test-secret",
		RecaptchaScoreThreshold: 0.5,
		GitHubToken:             "test-token",
		GitHubOwner:             "testowner",
		GitHubRepo:              "testrepo",
		AllowedOrigins:          []string{"https://example.com"},
		RedirectURL:             "https://example.com/success",
		RateLimitRequests:       10,
		RateLimitWindow:         60,
	}

	server := New(config)
	assert.NotNil(t, server)
	assert.Equal(t, config, server.config)
	assert.NotNil(t, server.router)
	assert.NotNil(t, server.limiter)
}

func TestHealthEndpoint(t *testing.T) {
	config := &Config{
		Port:              "8080",
		AllowedOrigins:    []string{"*"},
		RateLimitRequests: 100,
		RateLimitWindow:   60,
	}

	server := New(config)

	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestGuestbookSubmission_MissingName(t *testing.T) {
	config := &Config{
		Port:              "8080",
		AllowedOrigins:    []string{"*"},
		RateLimitRequests: 100,
		RateLimitWindow:   60,
	}

	server := New(config)

	// Test JSON request with missing name but valid JSON
	payload := map[string]string{
		"message": "Test message",
	}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/guestbook", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	// The response should contain an error about missing name
	responseBody := rr.Body.String()
	assert.True(t, strings.Contains(responseBody, "Name is required") || strings.Contains(responseBody, "Invalid"))
}

func TestGuestbookSubmission_HoneypotDetection(t *testing.T) {
	config := &Config{
		Port:              "8080",
		AllowedOrigins:    []string{"*"},
		RateLimitRequests: 100,
		RateLimitWindow:   60,
	}

	server := New(config)

	// Test form submission with honeypot field filled
	form := url.Values{}
	form.Add("name", "Test User")
	form.Add("message", "Test message")
	form.Add("website", "http://spam.com") // honeypot field

	req, err := http.NewRequest("POST", "/guestbook", strings.NewReader(form.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Thank you for your submission")
}

func TestIsLikelySpam(t *testing.T) {
	server := &Server{}

	tests := []struct {
		name     string
		request  GuestbookRequest
		expected bool
	}{
		{
			name: "Clean submission",
			request: GuestbookRequest{
				Name:    "John Doe",
				Message: "Hello, this is a nice website!",
			},
			expected: false,
		},
		{
			name: "Contains URL",
			request: GuestbookRequest{
				Name:    "Spammer",
				Message: "Check out https://spam.com for great deals!",
			},
			expected: true,
		},
		{
			name: "Contains spam keywords",
			request: GuestbookRequest{
				Name:    "Bot",
				Message: "Free viagra! Click here now!",
			},
			expected: true,
		},
		{
			name: "Excessive length",
			request: GuestbookRequest{
				Name:    "Long Message User",
				Message: strings.Repeat("This is a very long message. ", 100),
			},
			expected: true,
		},
		{
			name: "Repetitive content",
			request: GuestbookRequest{
				Name:    "Repetitive User",
				Message: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.isLikelySpam(tt.request)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCORSHeaders(t *testing.T) {
	config := &Config{
		Port:              "8080",
		AllowedOrigins:    []string{"https://b10a.co", "http://localhost:1313"},
		RateLimitRequests: 100,
		RateLimitWindow:   60,
	}

	server := New(config)

	// Test OPTIONS request
	req, err := http.NewRequest("OPTIONS", "/guestbook", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "https://b10a.co")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "https://b10a.co", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Methods"), "POST")
}

func TestRateLimiting(t *testing.T) {
	config := &Config{
		Port:              "8080",
		AllowedOrigins:    []string{"*"},
		RateLimitRequests: 2, // Very low limit for testing
		RateLimitWindow:   60,
	}

	server := New(config)

	// Make requests up to the limit
	for i := 0; i < 2; i++ {
		req, err := http.NewRequest("GET", "/health", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// Next request should be rate limited
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
}

// MockGitHubClient implements GitHubClientInterface for testing
type MockGitHubClient struct {
	shouldFail bool
}

func (m *MockGitHubClient) CreateGuestbookEntry(ctx context.Context, req GuestbookRequest) error {
	if m.shouldFail {
		return assert.AnError
	}
	return nil
}

func TestGuestbookSubmission_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &Config{
		Port:              "8080",
		AllowedOrigins:    []string{"*"},
		RateLimitRequests: 100,
		RateLimitWindow:   60,
		RedirectURL:       "https://example.com/success",
	}

	server := New(config)
	// Replace with mock for testing
	server.github = &MockGitHubClient{shouldFail: false}

	payload := map[string]string{
		"name":     "Test User",
		"message":  "This is a test message",
		"redirect": "https://example.com/success",
	}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/guestbook", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	// Should redirect on success
	assert.Equal(t, http.StatusFound, rr.Code)
}
