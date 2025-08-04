package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	server "github.com/bryankaraffa/b10a.co/guestbook-server/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGuestbookServerIntegration runs integration tests against the actual server
func TestGuestbookServerIntegration(t *testing.T) {
	// Skip if we don't have required environment variables
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("GITHUB_TOKEN not set, skipping integration tests")
	}

	// Load configuration like main.go does
	config := &server.Config{
		Port:               "8081", // Use different port for testing
		AkismetAPIKey:      os.Getenv("AKISMET_API_KEY"),
		AkismetSiteURL:     os.Getenv("AKISMET_SITE_URL"),
		RecaptchaSecretKey: os.Getenv("RECAPTCHA_SECRET_KEY"),
		GitHubToken:        os.Getenv("GITHUB_TOKEN"),
		GitHubOwner:        os.Getenv("GITHUB_OWNER"),
		GitHubRepo:         os.Getenv("GITHUB_REPO"),
		AllowedOrigins:     []string{"http://localhost:8081", "*"},
		RedirectURL:        os.Getenv("REDIRECT_URL"),
		RateLimitRequests:  100, // Higher limit for testing
		RateLimitWindow:    60,
	}

	// Set defaults
	if config.AkismetSiteURL == "" {
		config.AkismetSiteURL = "https://b10a.co"
	}
	if config.GitHubOwner == "" {
		config.GitHubOwner = "bryankaraffa"
	}
	if config.GitHubRepo == "" {
		config.GitHubRepo = "b10a.co"
	}

	// Create and start server
	srv := server.New(config)

	// Start server in background
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(time.Second)

	// Test health endpoint
	t.Run("Health Check", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8081/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "ok", result["status"])
	})

	// Test CORS
	t.Run("CORS Headers", func(t *testing.T) {
		req, err := http.NewRequest("OPTIONS", "http://localhost:8081/guestbook", nil)
		require.NoError(t, err)
		req.Header.Set("Origin", "http://localhost:8081")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Methods"))
	})

	// Test guestbook submission validation
	t.Run("Invalid Submission", func(t *testing.T) {
		payload := map[string]string{
			"message": "Test message without name",
		}
		jsonData, _ := json.Marshal(payload)

		resp, err := http.Post("http://localhost:8081/guestbook", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Test honeypot detection
	t.Run("Honeypot Detection", func(t *testing.T) {
		payload := map[string]string{
			"name":    "Test User",
			"message": "Test message",
			"website": "http://spam.com", // honeypot field
		}
		jsonData, _ := json.Marshal(payload)

		resp, err := http.Post("http://localhost:8081/guestbook", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Contains(t, result["message"], "Thank you")
	})
}

// TestRealGitHubIntegration tests actual GitHub integration (commented out by default)
func TestRealGitHubIntegration(t *testing.T) {
	t.Skip("Skipping real GitHub integration test - uncomment to test with real API")

	// Uncomment the following to test real GitHub integration
	/*
		if os.Getenv("GITHUB_TOKEN") == "" || os.Getenv("RUN_GITHUB_TESTS") != "true" {
			t.Skip("GITHUB_TOKEN not set or RUN_GITHUB_TESTS != true, skipping GitHub integration tests")
		}

		githubClient := server.NewGitHubClient(
			os.Getenv("GITHUB_TOKEN"),
			os.Getenv("GITHUB_OWNER"),
			os.Getenv("GITHUB_REPO"),
		)

		req := server.GuestbookRequest{
			Name:    "Integration Test User",
			Message: fmt.Sprintf("Test message from integration test at %s", time.Now().Format(time.RFC3339)),
		}

		err := githubClient.CreateGuestbookEntry(context.Background(), req)
		assert.NoError(t, err)
	*/
}
