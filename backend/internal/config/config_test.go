package config

import (
	"strings"
	"testing"
)

func TestLoadRequiresStrongInternalServiceToken(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", strings.Repeat("j", 32))
	t.Setenv("INTERNAL_SERVICE_TOKEN", "short")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "INTERNAL_SERVICE_TOKEN") {
		t.Fatalf("Load() error = %v, want internal token validation", err)
	}
}
