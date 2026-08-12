package petstate

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestPublishDeliversToEverySubscriber(t *testing.T) {
	service := &Service{subscribers: make(map[uuid.UUID]map[chan Update]struct{})}
	userID := uuid.New()

	first, unsubscribeFirst := service.Subscribe(userID)
	defer unsubscribeFirst()

	second, unsubscribeSecond := service.Subscribe(userID)
	defer unsubscribeSecond()

	want := Update{UserID: userID, State: Snapshot{Happiness: 80}}
	service.publish(want)

	if got := <-first; got.State.Happiness != want.State.Happiness {
		t.Fatalf("first update = %+v, want %+v", got, want)
	}

	if got := <-second; got.State.Happiness != want.State.Happiness {
		t.Fatalf("second update = %+v, want %+v", got, want)
	}
}

func TestCareForwardsIdempotencyKey(t *testing.T) {
	var calls int

	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls++

		if got := request.Header.Get("Idempotency-Key"); got != "request-1" {
			t.Errorf("Idempotency-Key = %q", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"happiness":60}`)),
		}, nil
	})}

	service := &Service{baseURL: "http://pet-state", token: "token", httpClient: client}

	for attempt := 0; attempt < 2; attempt++ {
		result, err := service.Care(t.Context(), uuid.New(), CareStroke, "request-1")

		if err != nil || result.Happiness != 60 {
			t.Fatalf("Care() = %+v, err = %v", result, err)
		}
	}

	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestCareMapsIdempotencyConflict(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusConflict,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"code":"IDEMPOTENCY_KEY_CONFLICT"}`)),
		}, nil
	})}

	service := &Service{baseURL: "http://pet-state", httpClient: client}
	_, err := service.Care(t.Context(), uuid.New(), CareFeed, "request-1")

	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("Care() error = %v", err)
	}
}

func TestPublishKeepsLatestUpdateWhenBufferIsFull(t *testing.T) {
	service := &Service{subscribers: make(map[uuid.UUID]map[chan Update]struct{})}
	userID := uuid.New()

	updates, unsubscribe := service.Subscribe(userID)
	defer unsubscribe()

	for value := 1; value <= 12; value++ {
		service.publish(Update{UserID: userID, State: Snapshot{Happiness: float64(value)}})
	}

	latest := 0.0

	for len(updates) > 0 {
		latest = (<-updates).State.Happiness
	}

	if latest != 12 {
		t.Fatalf("latest happiness = %.0f, want 12", latest)
	}
}

func TestUnsubscribeClosesChannelOnce(t *testing.T) {
	service := &Service{subscribers: make(map[uuid.UUID]map[chan Update]struct{})}
	updates, unsubscribe := service.Subscribe(uuid.New())

	unsubscribe()
	unsubscribe()

	if _, open := <-updates; open {
		t.Fatal("updates channel is still open")
	}
}

func TestSanitizeGroupPart(t *testing.T) {
	if got := sanitizeGroupPart(" api-service/one "); got != "api-service-one" {
		t.Fatalf("sanitizeGroupPart() = %q", got)
	}

	if got := sanitizeGroupPart(""); got != "local" {
		t.Fatalf("empty sanitizeGroupPart() = %q", got)
	}
}
