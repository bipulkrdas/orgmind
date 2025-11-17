package extraction

import (
	"fmt"
	"time"
)

// ExtractionLogger handles logging for extraction operations
// Requirements: 5.2, 7.5 - Log extraction attempts, failures, and performance metrics
type ExtractionLogger struct {
	enabled bool
}

// NewExtractionLogger creates a new extraction logger
func NewExtractionLogger(enabled bool) *ExtractionLogger {
	return &ExtractionLogger{
		enabled: enabled,
	}
}

// LogExtractionStart logs the start of an extraction operation
func (l *ExtractionLogger) LogExtractionStart(contentType string, fileSize int64) {
	if !l.enabled {
		return
	}

	fmt.Printf("[EXTRACTION] Starting extraction - Format: %s, Size: %d bytes (%.2f MB)\n",
		contentType, fileSize, float64(fileSize)/(1024*1024))
}

// LogExtractionSuccess logs a successful extraction
func (l *ExtractionLogger) LogExtractionSuccess(contentType string, fileSize int64, duration time.Duration, textLength int) {
	if !l.enabled {
		return
	}

	fmt.Printf("[EXTRACTION] Success - Format: %s, Size: %d bytes, Duration: %v, Extracted: %d chars, Speed: %.2f MB/s\n",
		contentType, fileSize, duration, textLength,
		float64(fileSize)/(1024*1024)/duration.Seconds())
}

// LogExtractionFailure logs a failed extraction
func (l *ExtractionLogger) LogExtractionFailure(contentType string, fileSize int64, duration time.Duration, err error) {
	if !l.enabled {
		return
	}

	fmt.Printf("[EXTRACTION] Failed - Format: %s, Size: %d bytes, Duration: %v, Error: %v\n",
		contentType, fileSize, duration, err)
}

// LogExtractionTimeout logs an extraction timeout
func (l *ExtractionLogger) LogExtractionTimeout(contentType string, fileSize int64, timeout time.Duration) {
	if !l.enabled {
		return
	}

	fmt.Printf("[EXTRACTION] Timeout - Format: %s, Size: %d bytes, Timeout: %v\n",
		contentType, fileSize, timeout)
}

// LogExtractionMetrics logs current extraction metrics
func (l *ExtractionLogger) LogExtractionMetrics(metrics ExtractionMetrics) {
	if !l.enabled {
		return
	}

	fmt.Printf("[EXTRACTION] Metrics - Total: %d, Active: %d, Queued: %d, Failed: %d, Success Rate: %.2f%%\n",
		metrics.TotalExtractions, metrics.ActiveExtractions, metrics.QueuedExtractions,
		metrics.FailedExtractions, l.calculateSuccessRate(metrics))
}

// LogMemoryWarning logs a memory usage warning
func (l *ExtractionLogger) LogMemoryWarning(used int64, limit int64) {
	if !l.enabled {
		return
	}

	fmt.Printf("[EXTRACTION] Memory Warning - Used: %d bytes (%.2f MB), Limit: %d bytes (%.2f MB), Usage: %.2f%%\n",
		used, float64(used)/(1024*1024), limit, float64(limit)/(1024*1024),
		float64(used)/float64(limit)*100)
}

// LogQueueStatus logs the current queue status
func (l *ExtractionLogger) LogQueueStatus(active int, queued int, maxConcurrent int) {
	if !l.enabled {
		return
	}

	fmt.Printf("[EXTRACTION] Queue Status - Active: %d/%d, Queued: %d\n",
		active, maxConcurrent, queued)
}

// calculateSuccessRate calculates the success rate from metrics
func (l *ExtractionLogger) calculateSuccessRate(metrics ExtractionMetrics) float64 {
	if metrics.TotalExtractions == 0 {
		return 0.0
	}

	successful := metrics.TotalExtractions - metrics.FailedExtractions
	return float64(successful) / float64(metrics.TotalExtractions) * 100
}

// ExtractionEvent represents an extraction event for structured logging
type ExtractionEvent struct {
	Timestamp   time.Time
	ContentType string
	FileSize    int64
	Duration    time.Duration
	Success     bool
	Error       string
	TextLength  int
}

// ExtractionStats tracks extraction statistics
type ExtractionStats struct {
	TotalExtractions      int64
	SuccessfulExtractions int64
	FailedExtractions     int64
	TotalDuration         time.Duration
	TotalBytesProcessed   int64
	ByFormat              map[string]*FormatStats
}

// FormatStats tracks statistics for a specific format
type FormatStats struct {
	Count           int64
	SuccessCount    int64
	FailedCount     int64
	TotalDuration   time.Duration
	TotalBytes      int64
	AverageDuration time.Duration
	AverageSpeed    float64 // MB/s
}

// NewExtractionStats creates a new extraction stats tracker
func NewExtractionStats() *ExtractionStats {
	return &ExtractionStats{
		ByFormat: make(map[string]*FormatStats),
	}
}

// RecordExtraction records an extraction event
func (s *ExtractionStats) RecordExtraction(event ExtractionEvent) {
	s.TotalExtractions++
	s.TotalDuration += event.Duration
	s.TotalBytesProcessed += event.FileSize

	if event.Success {
		s.SuccessfulExtractions++
	} else {
		s.FailedExtractions++
	}

	// Update format-specific stats
	formatStats, exists := s.ByFormat[event.ContentType]
	if !exists {
		formatStats = &FormatStats{}
		s.ByFormat[event.ContentType] = formatStats
	}

	formatStats.Count++
	formatStats.TotalDuration += event.Duration
	formatStats.TotalBytes += event.FileSize

	if event.Success {
		formatStats.SuccessCount++
	} else {
		formatStats.FailedCount++
	}

	// Calculate averages
	if formatStats.Count > 0 {
		formatStats.AverageDuration = formatStats.TotalDuration / time.Duration(formatStats.Count)
		if formatStats.TotalDuration.Seconds() > 0 {
			formatStats.AverageSpeed = float64(formatStats.TotalBytes) / (1024 * 1024) / formatStats.TotalDuration.Seconds()
		}
	}
}

// GetStats returns a copy of the current stats
func (s *ExtractionStats) GetStats() ExtractionStats {
	return *s
}
