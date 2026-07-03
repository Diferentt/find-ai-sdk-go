package findai

import (
	"errors"
	"testing"
)

func TestNewClient_RequiresAPIKey(t *testing.T) {
	if _, err := NewClient(""); err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestNewClient_RejectsBadPrefix(t *testing.T) {
	_, err := NewClient("sk_not_a_findai_key", WithBaseURL("https://example.com"))
	if err == nil {
		t.Fatal("expected error for bad API key prefix")
	}
	if !errors.Is(err, ErrInvalidAPIKey) {
		t.Fatalf("expected ErrInvalidAPIKey, got %v", err)
	}
}

func TestNewClient_RequiresBaseURL(t *testing.T) {
	if _, err := NewClient("fai_abc123"); err == nil {
		t.Fatal("expected error when no base URL is configured")
	}
}

func TestNewClient_Succeeds(t *testing.T) {
	c, err := NewClient("fai_abc123", WithBaseURL("https://example.com/"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if c.t.BaseURL != "https://example.com" {
		t.Fatalf("BaseURL = %q, want trailing slash trimmed", c.t.BaseURL)
	}
}
