package main

import (
	"log"
	"os"

	server "github.com/bryankaraffa/b10a.co/guestbook-server/pkg"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env files
	// Try to load .env.local first (for local development)
	if err := godotenv.Load(".env.local"); err != nil {
		log.Printf("No .env.local file found: %v", err)
		// Try to load .env as fallback
		if err := godotenv.Load(".env"); err != nil {
			log.Printf("No .env file found: %v", err)
		}
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

	// Create and start server
	srv := server.New(config)
	log.Printf("Starting guestbook server on port %s", config.Port)
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
