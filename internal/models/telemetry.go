package models

import "time"

// InverterData represents the live state and telemetry of an SMA inverter.
type InverterData struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	IP              string   `json:"ip"`
	Unit            int      `json:"unit"`
	Connected       bool     `json:"connected"`
	Error           *string  `json:"error"`
	OperatingStatus string   `json:"operating_status"` // e.g. "OK", "Warning", "Fault", "Offline"
	PowerW          int32    `json:"power_w"`
	PowerKW         float64  `json:"power_kw"`
	PL1KW           float64  `json:"p_l1_kw"`
	PL2KW           float64  `json:"p_l2_kw"`
	PL3KW           float64  `json:"p_l3_kw"`
	FrequencyHz     float64  `json:"frequency_hz"`
	DailyKWh        float64  `json:"daily_kwh"`
	TotalMWh        float64  `json:"total_mwh"`
	TemperatureC    float64  `json:"temperature_c"`
	Serial          string   `json:"serial"`
	LastSeen        string   `json:"last_seen"`
}

// PlantSummary represents high-level metrics for the entire PV facility.
type PlantSummary struct {
	Timestamp      string  `json:"timestamp"`
	Status         string  `json:"status"` // "ONLINE", "DEGRADED", "OFFLINE"
	SimulationMode bool    `json:"simulation_mode"`
	TotalPowerKW   float64 `json:"total_power_kw"`
	TotalDailyKWh  float64 `json:"total_daily_kwh"`
	TotalAccumMWh  float64 `json:"total_accum_mwh"`
	GridFrequency  float64 `json:"grid_frequency_hz"`
	InvertersCount int     `json:"inverters_count"`
	OnlineCount    int     `json:"online_count"`
}

// PlantData encapsulates the complete snapshot served by the API.
type PlantData struct {
	Timestamp      string         `json:"timestamp"`
	SimulationMode bool           `json:"simulation_mode"`
	TotalPowerKW   float64        `json:"total_power_kw"`
	TotalDailyKWh  float64        `json:"total_daily_kwh"`
	TotalAccumMWh  float64        `json:"total_accum_mwh"`
	GridFrequency  float64        `json:"grid_frequency_hz"`
	Inverters      []InverterData `json:"inverters"`
}

// HealthResponse provides service health diagnostics.
type HealthResponse struct {
	Status        string    `json:"status"`
	UptimeSeconds float64   `json:"uptime_seconds"`
	Engine        string    `json:"engine"`
	ModbusHost    string    `json:"modbus_host"`
	OnlineCount   int       `json:"online_inverters"`
	Timestamp     time.Time `json:"timestamp"`
}
