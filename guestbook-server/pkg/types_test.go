package guestbook_server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGuestbookRequest_ToEntry(t *testing.T) {
	req := &GuestbookRequest{
		Name:    "John Doe",
		Message: "Hello, world!",
	}

	entry := req.ToEntry()

	assert.NotEmpty(t, entry.ID)
	assert.Equal(t, "John Doe", entry.Name)
	assert.Equal(t, "Hello, world!", entry.Message)
	assert.True(t, entry.Date > 0)
	assert.True(t, entry.Date <= time.Now().Unix())
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Clean string",
			input:    "Hello, world!",
			expected: "Hello, world!",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "String with special characters",
			input:    "Test & validation",
			expected: "Test & validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	time.Sleep(time.Second) // Ensure different second timestamps
	id2 := generateID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 14) // YYYYMMDDHHMMSS = 14 characters
}
