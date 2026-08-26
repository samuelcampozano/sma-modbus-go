package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HandleStream handles Server-Sent Events (SSE) connections, continuously emitting
// JSON-encoded telemetry packets on every Modbus poll cycle.
// GET /api/stream
func (h *Handler) HandleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	subCh := h.store.Subscribe()
	defer h.store.Unsubscribe(subCh)

	// Send initial state immediately
	snapshot := h.store.GetSnapshot()
	if initialBytes, err := json.Marshal(snapshot); err == nil {
		fmt.Fprintf(w, "data: %s\n\n", initialBytes)
		flusher.Flush()
	}

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case data, ok := <-subCh:
			if !ok {
				return
			}
			bytes, err := json.Marshal(data)
			if err == nil {
				fmt.Fprintf(w, "data: %s\n\n", bytes)
				flusher.Flush()
			}
		}
	}
}
