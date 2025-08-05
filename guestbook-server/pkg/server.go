package guestbook_server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// GitHubClientInterface defines the interface for GitHub operations
type GitHubClientInterface interface {
	CreateGuestbookEntry(ctx context.Context, req GuestbookRequest) error
}

type Config struct {
	Port               string
	AkismetAPIKey      string
	AkismetSiteURL     string
	RecaptchaSecretKey string
	GitHubToken        string
	GitHubOwner        string
	GitHubRepo         string
	AllowedOrigins     []string
	RedirectURL        string
	RateLimitRequests  int
	RateLimitWindow    int
}

type Server struct {
	config    *Config
	router    *gin.Engine
	limiter   *rate.Limiter
	akismet   *AkismetClient
	recaptcha *RecaptchaClient
	github    GitHubClientInterface
}

func New(config *Config) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Create rate limiter: X requests per minute
	limiter := rate.NewLimiter(rate.Every(time.Duration(config.RateLimitWindow)*time.Second/time.Duration(config.RateLimitRequests)), config.RateLimitRequests)

	// Initialize clients
	akismet := NewAkismetClient(config.AkismetAPIKey, config.AkismetSiteURL)
	recaptcha := NewRecaptchaClient(config.RecaptchaSecretKey)
	github := NewGitHubClient(config.GitHubToken, config.GitHubOwner, config.GitHubRepo)

	server := &Server{
		config:    config,
		router:    router,
		limiter:   limiter,
		akismet:   akismet,
		recaptcha: recaptcha,
		github:    github,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// CORS middleware
	s.router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		for _, allowedOrigin := range s.config.AllowedOrigins {
			if origin == allowedOrigin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// Rate limiting middleware
	s.router.Use(func(c *gin.Context) {
		if !s.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	})

	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Guestbook submission endpoint
	s.router.POST("/guestbook", s.handleGuestbookSubmission)
}

func (s *Server) Start() error {
	return s.router.Run(":" + s.config.Port)
}

func (s *Server) handleGuestbookSubmission(c *gin.Context) {
	var req GuestbookRequest

	// Try to bind based on content type
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}
	} else {
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}
	}

	// Validate required fields
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	// Check honeypot field (if present, it's likely a bot)
	if req.Honeypot != "" {
		c.JSON(http.StatusOK, gin.H{"message": "Thank you for your submission"})
		return
	}

	// Verify reCAPTCHA
	if req.RecaptchaResponse != "" {
		if valid, err := s.recaptcha.Verify(c.Request.Context(), req.RecaptchaResponse, c.ClientIP()); err != nil || !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "reCAPTCHA verification failed", "details": err.Error()})
			// Log the error for debugging
			c.Errors = append(c.Errors, &gin.Error{
				Err:  fmt.Errorf("reCAPTCHA verification failed: %v", err),
				Type: gin.ErrorTypePublic,
			})
			return
		}
	}

	// Check for spam using Akismet
	if s.akismet != nil {
		isSpam, err := s.akismet.CheckSpam(c.Request.Context(), AkismetComment{
			UserIP:         c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),
			Referrer:       c.Request.Referer(),
			CommentType:    "guestbook",
			CommentAuthor:  req.Name,
			CommentContent: req.Message,
		})
		if err != nil {
			fmt.Printf("Akismet error: %v\n", err)
		} else if isSpam {
			c.JSON(http.StatusOK, gin.H{"message": "Thank you for your submission"})
			return
		}
	}

	// Additional AI-based spam detection
	if s.isLikelySpam(req) {
		c.JSON(http.StatusOK, gin.H{"message": "Thank you for your submission"})
		return
	}

	// Create pull request with the guestbook entry
	if err := s.github.CreateGuestbookEntry(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit entry"})
		return
	}

	// Redirect or return success
	if req.Redirect != "" {
		c.Redirect(http.StatusFound, req.Redirect)
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Thank you for your submission! It will be reviewed before being published."})
	}
}

func (s *Server) isLikelySpam(req GuestbookRequest) bool {
	// Simple heuristics for detecting spam
	suspiciousPatterns := []string{
		"http://", "https://", "www.", ".com", ".net", ".org",
		"click here", "buy now", "free", "offer", "deal",
		"viagra", "casino", "loan", "crypto", "bitcoin",
	}

	content := req.Name + " " + req.Message
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	// Check for excessive length
	if len(req.Message) > 1000 {
		return true
	}

	// Check for repetitive content
	if isRepetitive(req.Message) {
		return true
	}

	return false
}

func isRepetitive(text string) bool {
	// Simple check for repetitive patterns
	if len(text) < 10 {
		return false
	}

	// Check if the same character appears more than 70% of the time
	charCount := make(map[rune]int)
	for _, char := range text {
		charCount[char]++
	}

	for _, count := range charCount {
		if float64(count)/float64(len(text)) > 0.7 {
			return true
		}
	}

	return false
}
