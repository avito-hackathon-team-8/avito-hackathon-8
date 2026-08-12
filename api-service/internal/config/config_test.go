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

func TestLoadParsesDatabasePool(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", strings.Repeat("j", 32))
	t.Setenv("INTERNAL_SERVICE_TOKEN", strings.Repeat("i", 32))
	t.Setenv("DB_MAX_OPEN_CONNS", "40")
	t.Setenv("DB_MAX_IDLE_CONNS", "12")
	t.Setenv("DB_CONN_MAX_LIFETIME", "45m")
	t.Setenv("DB_CONN_MAX_IDLE_TIME", "3m")

	cfg, err := Load()

	if err != nil {
		t.Fatal(err)
	}

	if cfg.DatabasePool.MaxOpenConns != 40 || cfg.DatabasePool.MaxIdleConns != 12 || cfg.DatabasePool.ConnMaxLifetime.String() != "45m0s" || cfg.DatabasePool.ConnMaxIdleTime.String() != "3m0s" {
		t.Fatalf("database pool = %+v", cfg.DatabasePool)
	}
}
