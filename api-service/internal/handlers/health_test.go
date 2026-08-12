package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type readinessStub struct{ err error }

func (stub readinessStub) Ready(context.Context) error { return stub.err }

func TestHealthReady(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})

	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		checker readinessChecker
		want    int
	}{
		{name: "ready", checker: readinessStub{}, want: http.StatusOK},
		{name: "dependency unavailable", checker: readinessStub{err: errors.New("unavailable")}, want: http.StatusServiceUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()

			healthReady(db, test.checker).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/health/ready", nil))

			if response.Code != test.want {
				t.Fatalf("status = %d, want %d", response.Code, test.want)
			}
		})
	}
}
