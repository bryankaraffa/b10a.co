package guestbook_server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Debug logging function for server
func serverDebugLog(format string, args ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// GitHubClientInterface defines the interface for GitHub operations
type GitHubClientInterface interface {
	CreateGuestbookEntry(ctx context.Context, req GuestbookRequest) error
}

type Config struct {
	Port                    string
	AkismetAPIKey           string
	AkismetSiteURL          string
	RecaptchaSecretKey      string
	RecaptchaScoreThreshold float64
	GitHubToken             string
	GitHubOwner             string
	GitHubRepo              string
	GitHubBranch            string
	AllowedOrigins          []string
	AllowedRedirectDomains  []string
	RedirectURL             string
	RateLimitRequests       int
	RateLimitWindow         int
}

type RecaptchaVerifier interface {
	Verify(ctx context.Context, response, remoteIP string) (bool, error)
}

type Server struct {
	config         *Config
	router         *gin.Engine
	ipRateLimiters map[string]*rate.Limiter
	mu             *sync.Mutex
	akismet        *AkismetClient
	recaptcha      RecaptchaVerifier
	github         GitHubClientInterface
}

func New(config *Config) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Initialize clients
	akismet := NewAkismetClient(config.AkismetAPIKey, config.AkismetSiteURL)
	recaptcha := NewRecaptchaClient(config.RecaptchaSecretKey, config.RecaptchaScoreThreshold)
	github := NewGitHubClient(config.GitHubToken, config.GitHubOwner, config.GitHubRepo, config.GitHubBranch)

	server := &Server{
		config:         config,
		router:         router,
		ipRateLimiters: make(map[string]*rate.Limiter),
		mu:             &sync.Mutex{},
		akismet:        akismet,
		recaptcha:      recaptcha,
		github:         github,
	}

	server.setupRoutes()

	// Start a background goroutine to clean up old rate limiters
	go server.cleanupRateLimiters()

	return server
}

func (s *Server) cleanupRateLimiters() {
	for {
		// Wait for a specified interval before cleaning up
		time.Sleep(10 * time.Minute)

		s.mu.Lock()
		// Create a new map for active limiters
		cleanedLimiters := make(map[string]*rate.Limiter)
		for ip, limiter := range s.ipRateLimiters {
			// A simple heuristic: if the limiter has not been used recently,
			// it might be considered for removal. Here we check if the limit
			// has been reached, which is a proxy for recent activity.
			// A more robust solution would track last access time.
			if limiter.Burst() < s.config.RateLimitRequests {
				cleanedLimiters[ip] = limiter
			}
		}
		s.ipRateLimiters = cleanedLimiters
		s.mu.Unlock()
	}
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
	s.router.Use(s.rateLimitMiddleware)

	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Guestbook submission endpoint
	s.router.POST("/guestbook", s.handleGuestbookSubmission)
}

func (s *Server) rateLimitMiddleware(c *gin.Context) {
	ip := c.ClientIP()
	limiter := s.getRateLimiter(ip)

	if !limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
		c.Abort()
		return
	}

	c.Next()
}

func (s *Server) getRateLimiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	limiter, exists := s.ipRateLimiters[ip]
	if !exists {
		// Create a new limiter for this IP
		limiter = rate.NewLimiter(
			rate.Every(time.Duration(s.config.RateLimitWindow)*time.Second/time.Duration(s.config.RateLimitRequests)),
			s.config.RateLimitRequests,
		)
		s.ipRateLimiters[ip] = limiter
	}

	return limiter
}

func (s *Server) Start() error {
	return s.router.Run(":" + s.config.Port)
}

func (s *Server) handleGuestbookSubmission(c *gin.Context) {
	var req GuestbookRequest

	serverDebugLog("Received guestbook submission from IP: %s, User-Agent: %s", c.ClientIP(), c.Request.UserAgent())

	// Try to bind based on content type
	contentType := c.GetHeader("Content-Type")
	serverDebugLog("Request content type: %s", contentType)
	if strings.Contains(contentType, "application/json") {
		if err := c.ShouldBindJSON(&req); err != nil {
			serverDebugLog("Failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}
	} else {
		if err := c.ShouldBind(&req); err != nil {
			serverDebugLog("Failed to bind form data: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}
	}

	serverDebugLog("Parsed request - Name: %s, Message length: %d, RecaptchaResponse present: %t",
		req.Name, len(req.Message), req.RecaptchaResponse != "")

	// Validate required fields
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	// Check honeypot field (if present, it's likely a bot)
	if req.Honeypot != "" {
		serverDebugLog("Honeypot field detected from IP: %s, silently rejecting", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{"message": "Thank you for your submission"})
		return
	}

	// Verify reCAPTCHA (required)
	if req.RecaptchaResponse == "" {
		serverDebugLog("No reCAPTCHA response provided from IP: %s", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "reCAPTCHA verification is required"})
		return
	}

	serverDebugLog("Verifying reCAPTCHA for IP: %s, Response length: %d", c.ClientIP(), len(req.RecaptchaResponse))
	if s.recaptcha == nil {
		serverDebugLog("reCAPTCHA client is nil, rejecting submission")
		c.JSON(http.StatusBadRequest, gin.H{"error": "reCAPTCHA client is nil"})
		return
	}

	valid, err := s.recaptcha.Verify(c.Request.Context(), req.RecaptchaResponse, c.ClientIP())
	if err != nil {
		serverDebugLog("reCAPTCHA verification error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "reCAPTCHA verification failed", "details": err.Error()})
		// Log the error for debugging
		c.Errors = append(c.Errors, &gin.Error{
			Err:  fmt.Errorf("reCAPTCHA verification failed: %v", err),
			Type: gin.ErrorTypePublic,
		})
		return
	}
	if !valid {
		serverDebugLog("reCAPTCHA verification failed: response was valid but score/success check failed for IP: %s", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "reCAPTCHA verification failed", "details": "Invalid reCAPTCHA response"})
		// Log the error for debugging
		c.Errors = append(c.Errors, &gin.Error{
			Err:  fmt.Errorf("reCAPTCHA verification failed: invalid response"),
			Type: gin.ErrorTypePublic,
		})
		return
	}
	serverDebugLog("reCAPTCHA verification successful for IP: %s", c.ClientIP())

	// Check for spam using Akismet
	if s.akismet != nil {
		serverDebugLog("Starting Akismet spam check for submission from %s", req.Name)
		isSpam, err := s.akismet.CheckSpam(c.Request.Context(), AkismetComment{
			UserIP:         c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),
			Referrer:       c.Request.Referer(),
			CommentType:    "guestbook",
			CommentAuthor:  req.Name,
			CommentContent: req.Message,
		})
		if err != nil {
			serverDebugLog("Akismet check failed with error: %v", err)
			fmt.Printf("Akismet error: %v\n", err)
		} else if isSpam {
			serverDebugLog("Akismet detected spam from %s (IP: %s), silently rejecting", req.Name, c.ClientIP())
			c.JSON(http.StatusOK, gin.H{"message": "Thank you for your submission"})
			return
		} else {
			serverDebugLog("Akismet check passed for %s (IP: %s)", req.Name, c.ClientIP())
		}
	} else {
		serverDebugLog("Akismet client not configured, skipping spam check")
	}

	// Additional AI-based spam detection
	serverDebugLog("Running additional spam heuristics for %s", req.Name)
	if s.isLikelySpam(req) {
		serverDebugLog("Custom spam heuristics detected spam from %s (IP: %s), silently rejecting", req.Name, c.ClientIP())
		c.JSON(http.StatusOK, gin.H{"message": "Thank you for your submission"})
		return
	} else {
		serverDebugLog("Custom spam heuristics passed for %s", req.Name)
	}

	// Create pull request with the guestbook entry
	serverDebugLog("All spam checks passed, creating GitHub pull request for %s", req.Name)
	if err := s.github.CreateGuestbookEntry(c.Request.Context(), req); err != nil {
		serverDebugLog("Failed to create GitHub pull request for %s: %v", req.Name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit entry"})
		return
	}

	serverDebugLog("Successfully created GitHub pull request for guestbook entry from %s", req.Name)

	// Redirect or return success
	if req.Redirect != "" {
		if s.isValidRedirect(req.Redirect) {
			serverDebugLog("Redirecting to valid URL: %s", req.Redirect)
			c.Redirect(http.StatusFound, req.Redirect)
		} else {
			serverDebugLog("Invalid redirect URL blocked: %s", req.Redirect)
			// Do not redirect, instead return a generic success message
			c.JSON(http.StatusOK, gin.H{"message": "Thank you for your submission! It will be reviewed before being published."})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Thank you for your submission! It will be reviewed before being published."})
	}
}

func (s *Server) isValidRedirect(redirectURL string) bool {
	// Parse the redirect URL
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		return false // Invalid URL format
	}

	// Check if the hostname is in the allowed list
	for _, domain := range s.config.AllowedRedirectDomains {
		if parsedURL.Hostname() == domain {
			return true
		}
	}

	return false
}

func (s *Server) isLikelySpam(req GuestbookRequest) bool {
	// Simple heuristics for detecting spam
	suspiciousPatterns := []string{
		`[URL=http`, `[url=http`, `[link=http`,
		"click here", "buy now", "free", "offer", "deal",
		"viagra", "casino", "loan", "crypto", "bitcoin",
	}

	content := req.Name + " " + req.Message
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(content), pattern) {
			serverDebugLog("Suspicious pattern detected: '%s' in content from %s", pattern, req.Name)
			return true
		}
	}

	// Check for excessive links
	linkRegex := regexp.MustCompile(`(http|ftp|https)://([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?`)
	if len(linkRegex.FindAllString(req.Message, -1)) > 2 {
		serverDebugLog("Too many links detected in message from %s, flagging as spam", req.Name)
		return true
	}

	// Check for excessive length
	if len(req.Message) > 1000 {
		serverDebugLog("Message too long (%d chars) from %s, flagging as spam", len(req.Message), req.Name)
		return true
	}

	// Check for repetitive content
	if isRepetitive(req.Message) {
		serverDebugLog("Repetitive content detected from %s, flagging as spam", req.Name)
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
