package main

import (
	"log"
	"os"
	"strconv"

	server "github.com/bryankaraffa/b10a.co/guestbook-server/pkg"
	"github.com/joho/godotenv"
)

func debugLog(format string, args ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func maskKey(key string) string {
	if key == "" {
		return "<not set>"
	}
	if len(key) <= 8 {
		return "<masked>"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func main() {
	// Check if debug mode is enabled
	if os.Getenv("DEBUG") == "true" {
		log.Printf("[DEBUG] Debug mode enabled")
	}

	// Load environment variables from .env files
	// Try to load .env.local first (for local development)
	if err := godotenv.Load(".env.local"); err != nil {
		debugLog("No .env.local file found: %v", err)
		// Try to load .env as fallback
		if err := godotenv.Load(".env"); err != nil {
			debugLog("No .env file found: %v", err)
		}
	} else {
		debugLog("Loaded .env.local file")
	}

	// Load configuration from environment variables
	config := &server.Config{
		Port:               os.Getenv("PORT"),
		AkismetAPIKey:      os.Getenv("AKISMET_API_KEY"),
		AkismetSiteURL:     os.Getenv("AKISMET_SITE_URL"),
		RecaptchaSecretKey: os.Getenv("RECAPTCHA_SECRET_KEY"),
		GitHubToken:        os.Getenv("GITHUB_TOKEN"),
		GitHubOwner:        os.Getenv("GITHUB_OWNER"),
		GitHubRepo:         os.Getenv("GITHUB_REPO"),
		AllowedOrigins:     []string{"https://b10a.co", "http://localhost:1313"},
		RedirectURL:        os.Getenv("REDIRECT_URL"),
		RateLimitRequests:  10, // 10 requests per minute
		RateLimitWindow:    60, // 60 seconds
	}

	// Parse RecaptchaScoreThreshold from environment variable
	if scoreThresholdStr := os.Getenv("RECAPTCHA_SCORE_THRESHOLD"); scoreThresholdStr != "" {
		if threshold, err := strconv.ParseFloat(scoreThresholdStr, 64); err == nil {
			config.RecaptchaScoreThreshold = threshold
			debugLog("Using reCAPTCHA score threshold: %.2f", threshold)
		} else {
			log.Printf("Invalid RECAPTCHA_SCORE_THRESHOLD value: %s, using default 0.5", scoreThresholdStr)
			config.RecaptchaScoreThreshold = 0.5
		}
	} else {
		config.RecaptchaScoreThreshold = 0.5 // Default threshold
		debugLog("Using default reCAPTCHA score threshold: 0.5")
	}

	// Set defaults
	if config.Port == "" {
		config.Port = "8080"
	}
	if config.AkismetSiteURL == "" {
		config.AkismetSiteURL = "https://b10a.co"
	}
	if config.GitHubOwner == "" {
		config.GitHubOwner = "bryankaraffa"
	}
	if config.GitHubRepo == "" {
		config.GitHubRepo = "b10a.co"
	}
	if config.RedirectURL == "" {
		config.RedirectURL = "https://b10a.co/guestbook-success?success=true"
	}

	// Debug configuration
	debugLog("Configuration loaded:")
	debugLog("  Port: %s", config.Port)
	debugLog("  AkismetAPIKey: %s", maskKey(config.AkismetAPIKey))
	debugLog("  AkismetSiteURL: %s", config.AkismetSiteURL)
	debugLog("  RecaptchaSecretKey: %s", maskKey(config.RecaptchaSecretKey))
	debugLog("  RecaptchaScoreThreshold: %.2f", config.RecaptchaScoreThreshold)
	debugLog("  GitHubToken: %s", maskKey(config.GitHubToken))
	debugLog("  GitHubOwner: %s", config.GitHubOwner)
	debugLog("  GitHubRepo: %s", config.GitHubRepo)
	debugLog("  RedirectURL: %s", config.RedirectURL)
	debugLog("  RateLimitRequests: %d", config.RateLimitRequests)
	debugLog("  RateLimitWindow: %d", config.RateLimitWindow)

	// Create and start server
	srv := server.New(config)
	log.Printf("Starting guestbook server on port %s", config.Port)
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
