package metrics

import (
	"encoding/json"
	"net/http"
)

// Exporter serves metrics over HTTP as JSON.
type Exporter struct {
	collector *Collector
}

// NewExporter creates an Exporter backed by the given Collector.
func NewExporter(c *Collector) *Exporter {
	return &Exporter{collector: c}
}

// Handler returns an http.HandlerFunc that writes a JSON snapshot of
// all current metrics to the response.
func (e *Exporter) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := e.collector.Snapshot()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(snap); err != nil {
			http.Error(w, "failed to encode metrics", http.StatusInternalServerError)
		}
	}
}

// RegisterRoutes attaches the metrics endpoint to the given ServeMux
// at the provided path (e.g. "/metrics").
func (e *Exporter) RegisterRoutes(mux *http.ServeMux, path string) {
	mux.HandleFunc(path, e.Handler())
}
