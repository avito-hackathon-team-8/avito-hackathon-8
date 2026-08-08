package tasks

import (
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestPuppeteerAssignerRetriesOneServerFailure(t *testing.T) {
	var calls atomic.Int32
	token := strings.Repeat("t", 32)
	assigner := NewPuppeteerAssigner("http://puppeteer.test", token)
	assigner.client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("X-Service-Token") != token {
			t.Errorf("service token = %q", request.Header.Get("X-Service-Token"))
		}
		if calls.Add(1) == 1 {
			return &http.Response{StatusCode: http.StatusServiceUnavailable, Status: "503 Service Unavailable", Body: http.NoBody}, nil
		}
		return &http.Response{StatusCode: http.StatusNoContent, Status: "204 No Content", Body: http.NoBody}, nil
	})

	if assigner.client.Timeout != 2*time.Second {
		t.Fatalf("client timeout = %v, want 2s", assigner.client.Timeout)
	}
	if err := assigner.EnsureDailyTasks(t.Context(), uuid.New()); err != nil {
		t.Fatalf("EnsureDailyTasks() error = %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestPuppeteerAssignerDoesNotRetryClientFailure(t *testing.T) {
	var calls atomic.Int32
	assigner := NewPuppeteerAssigner("http://puppeteer.test", strings.Repeat("t", 32))
	assigner.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad Request", Body: http.NoBody}, nil
	})

	if err := assigner.EnsureDailyTasks(t.Context(), uuid.New()); err == nil {
		t.Fatal("EnsureDailyTasks() error = nil")
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d, want 1", calls.Load())
	}
}
