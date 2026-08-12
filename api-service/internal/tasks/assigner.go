package tasks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/httpx"
	appmetrics "github.com/avito-hackathon-team-8/avito-hackathon-8/api-service/internal/metrics"
	"github.com/google/uuid"
)

type DailyTasksAssigner struct {
	baseURL string
	token   string
	client  *http.Client
	metrics *appmetrics.Metrics
}

func NewDailyTasksAssigner(baseURL, token string, observed ...*appmetrics.Metrics) *DailyTasksAssigner {
	var metrics *appmetrics.Metrics
	if len(observed) > 0 {
		metrics = observed[0]
	}
	return &DailyTasksAssigner{
		baseURL: strings.TrimRight(baseURL, "/"), token: token,
		client: &http.Client{Timeout: 2 * time.Second}, metrics: metrics,
	}
}

func (assigner *DailyTasksAssigner) EnsureDailyTasks(ctx context.Context, userID uuid.UUID) error {
	endpoint := fmt.Sprintf("%s/internal/v1/users/%s/daily-tasks/ensure", assigner.baseURL, userID)

	var lastErr error

	for attempt := 0; attempt < 2; attempt++ {
		requestCtx, cancel := httpx.WithTimeout(ctx, 2*time.Second)
		request, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, nil)

		if err != nil {
			cancel()
			return err
		}

		request.Header.Set("X-Service-Token", assigner.token)

		started := time.Now()
		response, err := assigner.client.Do(request)
		cancel()
		result := "success"
		if err != nil {
			result = "error"
			if errors.Is(err, context.DeadlineExceeded) {
				result = "timeout"
			}
		}
		if assigner.metrics != nil {
			assigner.metrics.ExternalCalls.WithLabelValues("daily-tasks-service", http.MethodPost, result).Inc()
			assigner.metrics.ExternalTime.WithLabelValues("daily-tasks-service", http.MethodPost).Observe(time.Since(started).Seconds())
		}

		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			lastErr = err

			continue
		}

		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()

		if response.StatusCode == http.StatusNoContent {
			return nil
		}

		lastErr = fmt.Errorf("daily-tasks-service returned %s", response.Status)

		if response.StatusCode < http.StatusInternalServerError {
			return lastErr
		}
	}

	return lastErr
}
