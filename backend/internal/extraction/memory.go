package extraction

import (
	"context"
	"fmt"
	"runtime"
	"strings"
)

// MemoryTracker tracks memory usage during extraction
type MemoryTracker struct {
	maxMemory int64
	initial   uint64
}

// NewMemoryTracker creates a new memory tracker
func NewMemoryTracker(maxMemory int64) *MemoryTracker {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemoryTracker{
		maxMemory: maxMemory,
		initial:   m.Alloc,
	}
}

// Check verifies memory usage hasn't exceeded limits
func (mt *MemoryTracker) Check() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	used := m.Alloc - mt.initial
	if int64(used) > mt.maxMemory {
		return fmt.Errorf("memory limit exceeded: used %d bytes, limit %d bytes", used, mt.maxMemory)
	}

	return nil
}

// StreamingTextBuilder builds text content with memory limits
// Requirements: 5.4 - Use streaming where possible for large files
type StreamingTextBuilder struct {
	builder   strings.Builder
	maxSize   int64
	chunkSize int
	tracker   *MemoryTracker
}

// NewStreamingTextBuilder creates a new streaming text builder
func NewStreamingTextBuilder(maxSize int64, tracker *MemoryTracker) *StreamingTextBuilder {
	return &StreamingTextBuilder{
		maxSize:   maxSize,
		chunkSize: 4096, // 4KB chunks
		tracker:   tracker,
	}
}

// WriteString writes a string with memory checking
func (stb *StreamingTextBuilder) WriteString(s string) error {
	// Check if adding this string would exceed max size
	if int64(stb.builder.Len()+len(s)) > stb.maxSize {
		return fmt.Errorf("text size limit exceeded: would exceed %d bytes", stb.maxSize)
	}

	// Check memory usage periodically
	if stb.builder.Len()%stb.chunkSize == 0 && stb.tracker != nil {
		if err := stb.tracker.Check(); err != nil {
			return err
		}
	}

	stb.builder.WriteString(s)
	return nil
}

// String returns the built string
func (stb *StreamingTextBuilder) String() string {
	return stb.builder.String()
}

// Len returns the current length
func (stb *StreamingTextBuilder) Len() int {
	return stb.builder.Len()
}

// extractWithMemoryLimit wraps extraction with memory monitoring
func extractWithMemoryLimit(ctx context.Context, maxMemory int64, extractFn func() (string, error)) (string, error) {
	tracker := NewMemoryTracker(maxMemory)

	// Create a channel for the result
	type result struct {
		text string
		err  error
	}
	resultCh := make(chan result, 1)

	// Run extraction in a goroutine with memory monitoring
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultCh <- result{err: fmt.Errorf("extraction panic: %v", r)}
			}
		}()

		text, err := extractFn()
		resultCh <- result{text: text, err: err}
	}()

	// Wait for result or context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-resultCh:
		// Check final memory usage
		if res.err == nil {
			if err := tracker.Check(); err != nil {
				return "", err
			}
		}
		return res.text, res.err
	}
}
