// Package metrics provides job execution tracking and HTTP exposition.
package metrics

import (
	"net/http"
	"strconv"
	"time"
)

// ResponseRecorder wraps http.ResponseWriter to capture the status code
// written during request handling.
type ResponseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before delegating to the underlying writer.
func (r *ResponseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// newResponseRecorder returns a ResponseRecorder with a default 200 status.
func newResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
}

// InstrumentHandler wraps an http.Handler and records request duration and
// status code into the provided Collector under the given route label.
//
// Recorded metric key format: "http.<route>.<status>" (e.g. "http./metrics.200").
func InstrumentHandler(route string, next http.Handler, c *Collector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := newResponseRecorder(w)

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		statusStr := strconv.Itoa(rec.statusCode)
		key := "http." + route + "." + statusStr

		c.mu.Lock()
		entry := c.jobs[key]
		entry.Checks++
		// Reuse the DriftSeconds field to accumulate total latency in seconds
		// so callers can derive average latency if desired.
		entry.DriftSeconds += duration.Seconds()
		c.jobs[key] = entry
		c.mu.Unlock()
	})
}
