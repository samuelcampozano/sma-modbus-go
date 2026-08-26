package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

// InverterConfig holds configuration for an individual inverter.
type InverterConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Unit int    `json:"unit"`
}

// Config contains the entire application configuration.
type Config struct {
	ServerPort   int              `json:"server_port"`
	ModbusHost   string           `json:"modbus_host"`
	PollInterval time.Duration    `json:"poll_interval"`
	WebDir       string           `json:"web_dir"`
	Inverters    []InverterConfig `json:"inverters"`
}

// Load loads configuration from command-line flags and environment variables with sensible defaults.
func Load() *Config {
	defaultPort := 8050
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			defaultPort = p
		}
	}

	defaultModbus := "192.168.0.100:502"
	if envHost := os.Getenv("MODBUS_HOST"); envHost != "" {
		defaultModbus = envHost
	}

	defaultInterval := 2 * time.Second
	if envInterval := os.Getenv("POLL_INTERVAL"); envInterval != "" {
		if d, err := time.ParseDuration(envInterval); err == nil {
			defaultInterval = d
		}
	}

	defaultWebDir := "./web"
	if envWeb := os.Getenv("WEB_DIR"); envWeb != "" {
		defaultWebDir = envWeb
	}

	port := flag.Int("port", defaultPort, "HTTP server listening port")
	modbusHost := flag.String("modbus-host", defaultModbus, "Modbus TCP server address (host:port)")
	pollInterval := flag.Duration("interval", defaultInterval, "Telemetry polling interval (e.g. 2s, 5s)")
	webDir := flag.String("web-dir", defaultWebDir, "Path to static web assets directory")
	flag.Parse()

	return &Config{
		ServerPort:   *port,
		ModbusHost:   *modbusHost,
		PollInterval: *pollInterval,
		WebDir:       *webDir,
		Inverters: []InverterConfig{
			{ID: "inv1", Name: "SMA Sunny Highpower #1", Unit: 10},
			{ID: "inv2", Name: "SMA Sunny Highpower #2", Unit: 11},
		},
	}
}
