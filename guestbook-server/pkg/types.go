package guestbook_server

import (
	"time"
)

type GuestbookRequest struct {
	Name              string `form:"name" json:"name" binding:"required"`
	Message           string `form:"message" json:"message"`
	RecaptchaResponse string `form:"g-recaptcha-response" json:"g-recaptcha-response"`
	Redirect          string `form:"redirect" json:"redirect"`
	Honeypot          string `form:"website" json:"website"` // Honeypot field
}

type GuestbookEntry struct {
	ID      string `yaml:"_id"`
	Name    string `yaml:"name"`
	Message string `yaml:"message"`
	Date    int64  `yaml:"date"`
}

func (r *GuestbookRequest) ToEntry() *GuestbookEntry {
	return &GuestbookEntry{
		ID:      generateID(),
		Name:    sanitizeString(r.Name),
		Message: sanitizeString(r.Message),
		Date:    time.Now().Unix(),
	}
}

func sanitizeString(s string) string {
	// Escape HTML special characters to prevent XSS
	return template.HTMLEscapeString(s)
}

func generateID() string {
	return time.Now().Format("20060102150405")
}
