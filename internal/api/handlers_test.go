package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/samuelcampozano/sma-modbus-go/internal/config"
	"github.com/samuelcampozano/sma-modbus-go/internal/modbus"
	"github.com/samuelcampozano/sma-modbus-go/internal/models"
)

func createTestStore() *modbus.Store {
	cfg := &config.Config{
		ServerPort:   8050,
		ModbusHost:   "127.0.0.1:50200", // dummy
		PollInterval: 10 * time.Second,
		Inverters: []config.InverterConfig{
			{ID: "inv1", Name: "SMA Inverter 1", Unit: 10},
			{ID: "inv2", Name: "SMA Inverter 2", Unit: 11},
		},
	}
	return modbus.NewStore(cfg)
}

func TestHandleHealth(t *testing.T) {
	store := createTestStore()
	h := NewHandler(store)
	router := SetupRouter(h, "./web")

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var health models.HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &health); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if health.Engine != "golang" {
		t.Errorf("expected engine 'golang', got '%s'", health.Engine)
	}
}

func TestHandlePlant(t *testing.T) {
	store := createTestStore()
	h := NewHandler(store)
	router := SetupRouter(h, "./web")

	req := httptest.NewRequest(http.MethodGet, "/api/plant", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var summary models.PlantSummary
	if err := json.Unmarshal(rr.Body.Bytes(), &summary); err != nil {
		t.Fatalf("failed to decode summary: %v", err)
	}

	if summary.Status == "" {
		t.Errorf("expected non-empty status")
	}
}

func TestHandleInverters(t *testing.T) {
	store := createTestStore()
	h := NewHandler(store)
	router := SetupRouter(h, "./web")

	// 1. Get all inverters
	req := httptest.NewRequest(http.MethodGet, "/api/inverters", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	// 2. Inverter not found test
	reqNotFound := httptest.NewRequest(http.MethodGet, "/api/inverters/unknown-id", nil)
	rrNotFound := httptest.NewRecorder()
	router.ServeHTTP(rrNotFound, reqNotFound)

	if rrNotFound.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rrNotFound.Code)
	}
}
