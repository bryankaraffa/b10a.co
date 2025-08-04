package guestbook_server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	secretKey string
	client    *http.Client
}

type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

func NewRecaptchaClient(secretKey string) *RecaptchaClient {
	if secretKey == "" {
		return nil
	}
	return &RecaptchaClient{
		secretKey: secretKey,
		client:    &http.Client{},
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

	// For reCAPTCHA v3, check both success and score
	return result.Success && result.Score >= 0.5, nil
}
