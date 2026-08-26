package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/samuelcampozano/sma-modbus-go/internal/modbus"
)

// Handler contains HTTP handler methods backed by the telemetry Store.
type Handler struct {
	store *modbus.Store
}

// NewHandler creates a new API Handler instance.
func NewHandler(store *modbus.Store) *Handler {
	return &Handler{store: store}
}

// HandleData returns the complete plant snapshot in JSON.
// GET /api/data
func (h *Handler) HandleData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	data := h.store.GetSnapshot()
	_ = json.NewEncoder(w).Encode(data)
}

// HandlePlant returns a high-level summary of the plant KPIs.
// GET /api/plant
func (h *Handler) HandlePlant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	summary := h.store.GetPlantSummary()
	_ = json.NewEncoder(w).Encode(summary)
}

// HandleInverters returns either all inverters or a single inverter if an ID/Unit is specified in the URL path.
// GET /api/inverters
// GET /api/inverters/{id} (e.g. /api/inverters/inv1 or /api/inverters/10)
func (h *Handler) HandleInverters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	path := strings.TrimPrefix(r.URL.Path, "/api/inverters")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		inverters := h.store.GetInverters()
		_ = json.NewEncoder(w).Encode(inverters)
		return
	}

	inv, found := h.store.GetInverterByID(path)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "inverter not found"})
		return
	}
	_ = json.NewEncoder(w).Encode(inv)
}

// HandleHealth returns application health and uptime diagnostics.
// GET /api/health
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	health := h.store.GetHealth()
	_ = json.NewEncoder(w).Encode(health)
}
