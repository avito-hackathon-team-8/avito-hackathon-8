package health

import (
	"sync"
	"testing"
	"time"
)

func TestStatusSupportsConcurrentRecordAndSnapshot(t *testing.T) {
	status := NewStatus()

	const workers = 8
	const iterations = 100

	var waitGroup sync.WaitGroup

	waitGroup.Add(workers * 2)

	for worker := 0; worker < workers; worker++ {
		go func(worker int) {
			defer waitGroup.Done()

			for iteration := 0; iteration < iterations; iteration++ {
				status.Record("job-"+string(rune('a'+worker)), time.Now(), nil)
			}
		}(worker)

		go func() {
			defer waitGroup.Done()

			for iteration := 0; iteration < iterations; iteration++ {
				_ = status.Snapshot()
			}
		}()
	}

	waitGroup.Wait()

	if len(status.Snapshot()) != workers {
		t.Fatalf("status entries = %d, want %d", len(status.Snapshot()), workers)
	}
}
