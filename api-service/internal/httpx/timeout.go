package httpx

import (
	"context"
	"time"
)

func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if deadline, ok := parent.Deadline(); ok {
		remaining := time.Until(deadline)

		if remaining < timeout {
			return context.WithCancel(parent)
		}
	}

	return context.WithTimeout(parent, timeout)
}
