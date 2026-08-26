package modbus

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/samuelcampozano/sma-modbus-go/internal/config"
	"github.com/samuelcampozano/sma-modbus-go/internal/models"
)

// Store provides thread-safe access to live plant telemetry.
type Store struct {
	mu           sync.RWMutex
	current      models.PlantData
	subscribers  map[chan models.PlantData]bool
	subsMu       sync.Mutex
	startTime    time.Time
	cfg          *config.Config
	client       *Client
}

// NewStore initializes the telemetry store and Modbus client.
func NewStore(cfg *config.Config) *Store {
	return &Store{
		startTime:   time.Now(),
		cfg:         cfg,
		client:      NewClient(cfg.ModbusHost, 2500*time.Millisecond),
		subscribers: make(map[chan models.PlantData]bool),
		current: models.PlantData{
			Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
			SimulationMode: false,
			Inverters:      make([]models.InverterData, 0),
		},
	}
}

// GetSnapshot returns a copy of the current plant data.
func (s *Store) GetSnapshot() models.PlantData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// GetPlantSummary returns high-level summary KPIs.
func (s *Store) GetPlantSummary() models.PlantSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	onlineCount := 0
	for _, inv := range s.current.Inverters {
		if inv.Connected {
			onlineCount++
		}
	}

	status := "OFFLINE"
	if onlineCount == len(s.current.Inverters) && len(s.current.Inverters) > 0 {
		status = "ONLINE"
	} else if onlineCount > 0 {
		status = "DEGRADED"
	}

	return models.PlantSummary{
		Timestamp:      s.current.Timestamp,
		Status:         status,
		SimulationMode: s.current.SimulationMode,
		TotalPowerKW:   s.current.TotalPowerKW,
		TotalDailyKWh:  s.current.TotalDailyKWh,
		TotalAccumMWh:  s.current.TotalAccumMWh,
		GridFrequency:  s.current.GridFrequency,
		InvertersCount: len(s.current.Inverters),
		OnlineCount:    onlineCount,
	}
}

// GetInverters returns the slice of inverters.
func (s *Store) GetInverters() []models.InverterData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current.Inverters
}

// GetInverterByID looks up an inverter by internal ID or Unit ID.
func (s *Store) GetInverterByID(idOrUnit string) (models.InverterData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, inv := range s.current.Inverters {
		if inv.ID == idOrUnit || fmt.Sprintf("%d", inv.Unit) == idOrUnit {
			return inv, true
		}
	}
	return models.InverterData{}, false
}

// GetHealth returns system diagnostics.
func (s *Store) GetHealth() models.HealthResponse {
	s.mu.RLock()
	onlineCount := 0
	for _, inv := range s.current.Inverters {
		if inv.Connected {
			onlineCount++
		}
	}
	s.mu.RUnlock()

	status := "healthy"
	if onlineCount == 0 {
		status = "degraded"
	}

	return models.HealthResponse{
		Status:        status,
		UptimeSeconds: RoundToDecimals(time.Since(s.startTime).Seconds(), 1),
		Engine:        "golang",
		ModbusHost:    s.cfg.ModbusHost,
		OnlineCount:   onlineCount,
		Timestamp:     time.Now(),
	}
}

// Subscribe returns a channel that receives live updates on every poll.
func (s *Store) Subscribe() chan models.PlantData {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	ch := make(chan models.PlantData, 10)
	s.subscribers[ch] = true
	return ch
}

// Unsubscribe removes an SSE client channel.
func (s *Store) Unsubscribe(ch chan models.PlantData) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	if _, ok := s.subscribers[ch]; ok {
		delete(s.subscribers, ch)
		close(ch)
	}
}

func (s *Store) broadcast(data models.PlantData) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	for ch := range s.subscribers {
		select {
		case ch <- data:
		default:
		}
	}
}

// StartPoller initiates continuous background polling of all configured inverters.
func (s *Store) StartPoller(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()
	defer s.client.Close()

	// Initial poll immediately
	s.pollAll()

	for {
		select {
		case <-ctx.Done():
			log.Println("[Poller] Stopping background poller...")
			return
		case <-ticker.C:
			s.pollAll()
		}
	}
}

func (s *Store) pollAll() {
	var inverters []models.InverterData
	var totPowerKW float64
	var totDailyKWh float64
	var totAccumMWh float64
	gridFreq := 60.00

	for _, invCfg := range s.cfg.Inverters {
		inv := s.pollSingle(invCfg)
		inverters = append(inverters, inv)
		if inv.Connected {
			totPowerKW += inv.PowerKW
			totDailyKWh += inv.DailyKWh
			totAccumMWh += inv.TotalMWh
			if inv.FrequencyHz > 0 {
				gridFreq = inv.FrequencyHz
			}
		}
	}

	snapshot := models.PlantData{
		Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
		SimulationMode: false,
		TotalPowerKW:   RoundToDecimals(totPowerKW, 2),
		TotalDailyKWh:  RoundToDecimals(totDailyKWh, 2),
		TotalAccumMWh:  RoundToDecimals(totAccumMWh, 2),
		GridFrequency:  gridFreq,
		Inverters:      inverters,
	}

	s.mu.Lock()
	s.current = snapshot
	s.mu.Unlock()

	s.broadcast(snapshot)
}

func (s *Store) pollSingle(cfg config.InverterConfig) models.InverterData {
	inv := models.InverterData{
		ID:              cfg.ID,
		Name:            cfg.Name,
		IP:              s.cfg.ModbusHost,
		Unit:            cfg.Unit,
		Connected:       true,
		OperatingStatus: "OK",
		Serial:          fmt.Sprintf("SHP-U%d", cfg.Unit),
		FrequencyHz:     60.00,
		TemperatureC:    46.2,
		LastSeen:        time.Now().Format("15:04:05"),
	}

	// 1. Operating Status (30201)
	if regsStatus, err := s.client.ReadRegisters(uint8(cfg.Unit), RegOperatingStatus, 2); err == nil {
		if statusVal, ok := DecodeU32(regsStatus); ok {
			inv.OperatingStatus = DecodeOperatingStatus(statusVal)
		}
	}

	// 2. Active Power (30775, S32)
	regsPower, err := s.client.ReadRegisters(uint8(cfg.Unit), RegActivePowerTotal, 2)
	if err == nil {
		if pVal, ok := DecodeS32(regsPower); ok {
			inv.PowerW = pVal
			inv.PowerKW = RoundToDecimals(float64(pVal)/1000.0, 2)
			pPhase := RoundToDecimals(inv.PowerKW/3.0, 2)
			inv.PL1KW = pPhase
			inv.PL2KW = pPhase
			inv.PL3KW = pPhase
		}
	} else {
		inv.Connected = false
		inv.OperatingStatus = "Offline"
		errMsg := fmt.Sprintf("Modbus read error (Reg %d): %v", RegActivePowerTotal, err)
		inv.Error = &errMsg
	}

	// 3. Cumulative Energy Fed-In (30513, U64, 4 regs)
	if regsTot, err := s.client.ReadRegisters(uint8(cfg.Unit), RegTotalEnergyFedIn, 4); err == nil {
		if totVal, ok := DecodeU64(regsTot); ok {
			totalKWh := float64(totVal) / 1000.0
			inv.TotalMWh = RoundToDecimals(totalKWh/1000.0, 2)
		}
	}

	// 4. Daily Yield (30517, U64, 4 regs)
	if regsDay, err := s.client.ReadRegisters(uint8(cfg.Unit), RegDailyYield, 4); err == nil {
		if dayVal, ok := DecodeU64(regsDay); ok {
			inv.DailyKWh = RoundToDecimals(float64(dayVal)/1000.0, 2)
		}
	}

	return inv
}
