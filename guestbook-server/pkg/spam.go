package guestbook_server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

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
		return false, nil
	}

	data := url.Values{}
	data.Set("blog", a.siteURL)
	data.Set("user_ip", comment.UserIP)
	data.Set("user_agent", comment.UserAgent)
	data.Set("referrer", comment.Referrer)
	data.Set("comment_type", comment.CommentType)
	data.Set("comment_author", comment.CommentAuthor)
	data.Set("comment_content", comment.CommentContent)

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("https://%s.rest.akismet.com/1.1/comment-check", a.apiKey),
		strings.NewReader(data.Encode()))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "GuestbookServer/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("akismet API returned status %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	result := strings.TrimSpace(buf.String())

	return result == "true", nil
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
		return true, nil // Skip verification if not configured
	}

	data := url.Values{}
	data.Set("secret", r.secretKey)
	data.Set("response", response)
	data.Set("remoteip", remoteIP)

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://www.google.com/recaptcha/api/siteverify",
		strings.NewReader(data.Encode()))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result RecaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	// Log detailed response for debugging
	log.Printf("reCAPTCHA response: Success=%t, Score=%.2f, Action=%s, Hostname=%s, ErrorCodes=%v",
		result.Success, result.Score, result.Action, result.Hostname, result.ErrorCodes)

	// For reCAPTCHA v3, check both success and score
	isValid := result.Success && result.Score >= r.scoreThreshold

	if !isValid {
		if !result.Success {
			return false, fmt.Errorf("reCAPTCHA verification failed: %v", result.ErrorCodes)
		} else if result.Score < r.scoreThreshold {
			return false, fmt.Errorf("reCAPTCHA score too low: %.2f (minimum: %.2f)", result.Score, r.scoreThreshold)
		}
	}

	return isValid, nil
}
