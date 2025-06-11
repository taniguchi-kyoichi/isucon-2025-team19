package middleware

import (
	"log"
	"net/http"
	"time"
)

// RequestMetrics stores metrics for a single request
type RequestMetrics struct {
	Path       string
	Method     string
	StatusCode int
	Duration   time.Duration
}

var (
	// MetricsChannel is used to collect request metrics
	MetricsChannel = make(chan RequestMetrics, 1000)
)

// MetricsCollector collects and logs metrics
func MetricsCollector() {
	// Aggregate metrics every second
	ticker := time.NewTicker(1 * time.Second)
	metrics := make(map[string][]time.Duration)

	for {
		select {
		case m := <-MetricsChannel:
			key := m.Method + " " + m.Path
			metrics[key] = append(metrics[key], m.Duration)
		case <-ticker.C:
			// Calculate and log statistics
			for path, durations := range metrics {
				var total time.Duration
				max := time.Duration(0)
				for _, d := range durations {
					total += d
					if d > max {
						max = d
					}
				}
				avg := total / time.Duration(len(durations))
				log.Printf("Path: %s, Count: %d, Avg: %v, Max: %v\n",
					path, len(durations), avg, max)
			}
			// Reset metrics
			metrics = make(map[string][]time.Duration)
		}
	}
}

// MetricsMiddleware measures request processing time
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture the status code
		rw := &responseWriter{ResponseWriter: w}

		// Process the request
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start)
		MetricsChannel <- RequestMetrics{
			Path:       r.URL.Path,
			Method:     r.Method,
			StatusCode: rw.statusCode,
			Duration:   duration,
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = 200
	}
	return rw.ResponseWriter.Write(b)
} 