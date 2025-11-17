package extraction

import (
	"context"
	"fmt"
	"sync"
)

// ExtractionQueue manages concurrent extraction operations
// Requirements: 5.5 - Ensure thread-safe extraction operations
type ExtractionQueue struct {
	semaphore chan struct{}
	active    int
	mu        sync.Mutex
	metrics   *internalMetrics
}

// ExtractionMetrics tracks extraction statistics (exported for reading)
type ExtractionMetrics struct {
	TotalExtractions  int64
	ActiveExtractions int
	FailedExtractions int64
	QueuedExtractions int
}

// internalMetrics holds the internal metrics with mutex
type internalMetrics struct {
	mu                sync.RWMutex
	totalExtractions  int64
	activeExtractions int
	failedExtractions int64
	queuedExtractions int
}

// NewExtractionQueue creates a new extraction queue with max concurrent operations
func NewExtractionQueue(maxConcurrent int) *ExtractionQueue {
	return &ExtractionQueue{
		semaphore: make(chan struct{}, maxConcurrent),
		metrics:   &internalMetrics{},
	}
}

// Acquire acquires a slot in the extraction queue
func (eq *ExtractionQueue) Acquire(ctx context.Context) error {
	// Increment queued count
	eq.metrics.mu.Lock()
	eq.metrics.queuedExtractions++
	eq.metrics.mu.Unlock()

	// Try to acquire semaphore
	select {
	case eq.semaphore <- struct{}{}:
		// Successfully acquired
		eq.mu.Lock()
		eq.active++
		eq.mu.Unlock()

		eq.metrics.mu.Lock()
		eq.metrics.queuedExtractions--
		eq.metrics.activeExtractions++
		eq.metrics.totalExtractions++
		eq.metrics.mu.Unlock()

		return nil
	case <-ctx.Done():
		// Context cancelled while waiting
		eq.metrics.mu.Lock()
		eq.metrics.queuedExtractions--
		eq.metrics.mu.Unlock()
		return fmt.Errorf("extraction queue wait cancelled: %w", ctx.Err())
	}
}

// Release releases a slot in the extraction queue
func (eq *ExtractionQueue) Release() {
	<-eq.semaphore

	eq.mu.Lock()
	eq.active--
	eq.mu.Unlock()

	eq.metrics.mu.Lock()
	eq.metrics.activeExtractions--
	eq.metrics.mu.Unlock()
}

// RecordFailure records a failed extraction
func (eq *ExtractionQueue) RecordFailure() {
	eq.metrics.mu.Lock()
	eq.metrics.failedExtractions++
	eq.metrics.mu.Unlock()
}

// GetMetrics returns current extraction metrics
func (eq *ExtractionQueue) GetMetrics() ExtractionMetrics {
	eq.metrics.mu.RLock()
	defer eq.metrics.mu.RUnlock()

	return ExtractionMetrics{
		TotalExtractions:  eq.metrics.totalExtractions,
		ActiveExtractions: eq.metrics.activeExtractions,
		FailedExtractions: eq.metrics.failedExtractions,
		QueuedExtractions: eq.metrics.queuedExtractions,
	}
}

// ActiveCount returns the number of active extractions
func (eq *ExtractionQueue) ActiveCount() int {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	return eq.active
}

// Execute runs an extraction function with queue management
func (eq *ExtractionQueue) Execute(ctx context.Context, fn func() (string, error)) (string, error) {
	// Acquire slot in queue
	if err := eq.Acquire(ctx); err != nil {
		return "", err
	}
	defer eq.Release()

	// Execute extraction
	text, err := fn()
	if err != nil {
		eq.RecordFailure()
	}

	return text, err
}
