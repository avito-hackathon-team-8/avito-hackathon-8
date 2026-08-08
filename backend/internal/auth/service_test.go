package auth

import (
	"regexp"
	"testing"
)

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	email, err := normalizeEmail("  USER@example.com ")

	if err != nil {
		t.Fatalf("normalizeEmail returned an error: %v", err)
	}

	if email != "user@example.com" {
		t.Fatalf("normalizeEmail = %q, want user@example.com", email)
	}
}

func TestNormalizeEmailRejectsDisplayName(t *testing.T) {
	t.Parallel()

	if _, err := normalizeEmail("User <user@example.com>"); err == nil {
		t.Fatal("normalizeEmail accepted a display name")
	}
}

func TestGenerateCode(t *testing.T) {
	t.Parallel()

	code, err := generateCode(8)

	if err != nil {
		t.Fatalf("generateCode returned an error: %v", err)
	}

	if !regexp.MustCompile(`^[0-9]{8}$`).MatchString(code) {
		t.Fatalf("generateCode = %q, want 8 digits", code)
	}
}
