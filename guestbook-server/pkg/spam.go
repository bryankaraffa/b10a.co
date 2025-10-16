package guestbook_server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var debugEnabled = os.Getenv("DEBUG") == "true"

func debugLog(format string, args ...interface{}) {
	if debugEnabled {
		log.Printf("[DEBUG] "+format, args...)
	}
}

type AkismetClient struct {
	apiKey  string
	siteURL string
	client  *http.Client
}

type AkismetComment struct {
	UserIP         string
	UserAgent      string
	Referrer       string
	CommentType    string
	CommentAuthor  string
	CommentContent string
}

func NewAkismetClient(apiKey, siteURL string) *AkismetClient {
	if apiKey == "" {
		return nil
	}
	return &AkismetClient{
		apiKey:  apiKey,
		siteURL: siteURL,
		client:  &http.Client{},
	}
}

func (a *AkismetClient) CheckSpam(ctx context.Context, comment AkismetComment) (bool, error) {
	if a == nil {
		debugLog("Akismet client is nil, skipping spam check")
		return false, nil
	}

	debugLog("Starting Akismet spam check for IP: %s, Author: %s", comment.UserIP, comment.CommentAuthor)

	data := url.Values{}
	data.Set("blog", a.siteURL)
	data.Set("user_ip", comment.UserIP)
	data.Set("user_agent", comment.UserAgent)
	data.Set("referrer", comment.Referrer)
	data.Set("comment_type", comment.CommentType)
	data.Set("comment_author", comment.CommentAuthor)
	data.Set("comment_content", comment.CommentContent)

	debugLog("Akismet request data: blog=%s, user_ip=%s, comment_type=%s, content_length=%d",
		a.siteURL, comment.UserIP, comment.CommentType, len(comment.CommentContent))

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("https://%s.rest.akismet.com/1.1/comment-check", a.apiKey),
		strings.NewReader(data.Encode()))
	if err != nil {
		debugLog("Akismet request creation failed: %v", err)
		return false, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "GuestbookServer/1.0")

	debugLog("Sending request to Akismet API: %s", req.URL.String())

	resp, err := a.client.Do(req)
	if err != nil {
		debugLog("Akismet API request failed: %v", err)
		return false, err
	}
	defer resp.Body.Close()

	debugLog("Akismet API response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("akismet API returned status %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	result := strings.TrimSpace(buf.String())

	isSpam := result == "true"
	debugLog("Akismet result: %s (isSpam: %t) for author: %s", result, isSpam, comment.CommentAuthor)

	// Check for additional Akismet headers that provide debugging info
	if debugEnabled {
		if debugInfo := resp.Header.Get("X-akismet-debug-help"); debugInfo != "" {
			debugLog("Akismet debug info: %s", debugInfo)
		}
		if proTip := resp.Header.Get("X-akismet-pro-tip"); proTip != "" {
			debugLog("Akismet pro tip: %s", proTip)
		}
	}

	return isSpam, nil
}

type RecaptchaClient struct {
	secretKey      string
	scoreThreshold float64
	client         *http.Client
}

type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

func NewRecaptchaClient(secretKey string, scoreThreshold float64) *RecaptchaClient {
	if secretKey == "" {
		return nil
	}
	if scoreThreshold <= 0 {
		scoreThreshold = 0.5 // Default threshold
	}
	return &RecaptchaClient{
		secretKey:      secretKey,
		scoreThreshold: scoreThreshold,
		client:         &http.Client{},
	}
}

func (r *RecaptchaClient) Verify(ctx context.Context, response, remoteIP string) (bool, error) {
	if r == nil {
		debugLog("reCAPTCHA client is nil, skipping verification")
		return true, nil // Skip verification if not configured
	}

	debugLog("Starting reCAPTCHA verification for IP: %s, response length: %d", remoteIP, len(response))

	data := url.Values{}
	data.Set("secret", r.secretKey)
	data.Set("response", response)
	data.Set("remoteip", remoteIP)
	// Don't set action here - let the response tell us what action was used

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://www.google.com/recaptcha/api/siteverify",
		strings.NewReader(data.Encode()))
	if err != nil {
		debugLog("reCAPTCHA request creation failed: %v", err)
		return false, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	debugLog("Sending request to reCAPTCHA API")

	resp, err := r.client.Do(req)
	if err != nil {
		debugLog("reCAPTCHA API request failed: %v", err)
		return false, err
	}
	defer resp.Body.Close()

	var result RecaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		debugLog("Failed to decode reCAPTCHA response: %v", err)
		return false, err
	}

	// Log detailed response for debugging
	if debugEnabled {
		log.Printf("[DEBUG] reCAPTCHA response: Success=%t, Score=%.2f, Action=%s, Hostname=%s, ErrorCodes=%v",
			result.Success, result.Score, result.Action, result.Hostname, result.ErrorCodes)
	} else {
		log.Printf("reCAPTCHA response: Success=%t, Score=%.2f, Action=%s, Hostname=%s, ErrorCodes=%v",
			result.Success, result.Score, result.Action, result.Hostname, result.ErrorCodes)
	}

	// Additional debugging for your specific case
	if result.Score == 0.0 {
		if len(result.ErrorCodes) > 0 {
			debugLog("reCAPTCHA score is 0.0 with errors - this suggests a configuration issue")
		} else {
      return false, fmt.Errorf("reCAPTCHA score is 0.0 with no errors - this might be reCAPTCHA v2 on the frontend or a site key mismatch. Considering Spam to be safe.")
		}
	}

	// Check if basic verification succeeded first
	if !result.Success {
		debugLog("reCAPTCHA verification failed with errors: %v", result.ErrorCodes)
		return false, fmt.Errorf("reCAPTCHA verification failed: %v", result.ErrorCodes)
	}

	// Verify the action name
	if result.Action != "submit" {
		debugLog("reCAPTCHA action mismatch: expected 'submit', got '%s'", result.Action)
		return false, fmt.Errorf("reCAPTCHA action mismatch: expected 'submit', got '%s'", result.Action)
	}

	// For reCAPTCHA v3, check the score
	if result.Score < r.scoreThreshold {
		debugLog("reCAPTCHA score too low: %.2f (minimum: %.2f)", result.Score, r.scoreThreshold)
		return false, fmt.Errorf("reCAPTCHA score too low: %.2f (minimum: %.2f)", result.Score, r.scoreThreshold)
	}

	debugLog("reCAPTCHA verification successful: score=%.2f, threshold=%.2f", result.Score, r.scoreThreshold)
	return true, nil
}
