package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadyHandler(t *testing.T) {
	ok := func(context.Context) error { return nil }
	fail := func(context.Context) error { return errors.New("unavailable") }
	tests := []struct {
		name      string
		dbPing    func(context.Context) error
		kafkaPing func(context.Context) error
		want      int
	}{
		{name: "ready", dbPing: ok, kafkaPing: ok, want: http.StatusOK},
		{name: "postgres unavailable", dbPing: fail, kafkaPing: ok, want: http.StatusServiceUnavailable},
		{name: "kafka unavailable", dbPing: ok, kafkaPing: fail, want: http.StatusServiceUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()

			readyHandler(test.dbPing, test.kafkaPing).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

			if response.Code != test.want {
				t.Fatalf("status = %d, want %d", response.Code, test.want)
			}
		})
	}
}
