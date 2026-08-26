# ☀️ sma-modbus-go

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Protocol](https://img.shields.io/badge/Protocol-Modbus_TCP_(Port_502)-orange)](https://files.sma.de)
[![Coverage](https://img.shields.io/badge/Tests-Passing-brightgreen)]()

> High-performance, zero-dependency **Golang client library and real-time telemetry service** for **SMA Sunny Highpower** solar inverters and **SMA Data Manager M (ennexOS)** systems via Modbus TCP (Port 502).

*Read this in other languages: [📖 Español](README_es.md)*

---

## 🌟 Key Features

* **Zero External Dependencies**: Built entirely on Go's standard library (`net`, `encoding/binary`, `net/http`, `sync`). Ultra-fast compilation and minimal binary footprint (~8 MB).
* **Dual-Use Architecture**:
  * **As a Go Library (`pkg/smamodbus`)**: Import directly into your own Go applications, microservices, or SCADA collectors.
  * **As a Standalone Daemon (`solar_api.exe`)**: Runs a persistent poller, in-memory caching store, embedded HTML5 web dashboard, and REST API.
* **Real-time Event Streaming**: Native **Server-Sent Events (SSE)** endpoint (`/api/stream`) that continuously pushes telemetry updates every poll cycle (`curl -N`).
* **Hardware-Specific SMA Decoder**:
  * Big-Endian word unpacking for 32-bit (`S32`/`U32`) and 64-bit (`U64`) integers.
  * Filters SMA NaN/error sentinels (`-0x80000000`, `0xFFFFFFFF`) to prevent overflow during nighttime or sensor disconnects.
  * Automatically translates SMA status codes (e.g. `307` = OK, `455` = Warning, `35` = Fault).
* **Automatic Gateway Routing**: Full support for SMA Data Manager M / ennexOS gateway architectures addressing individual inverters via Unit IDs (Slave IDs).

---

## 📐 Architecture

```
 ┌─────────────────────────────────────────────────────────────┐
 │           SMA DATA MANAGER M / ennexOS (GATEWAY)            │
 │                     IP: 192.168.0.100:502                   │
 │                                                             │
 │   [ Inverter #1: Unit ID 10 ]   [ Inverter #2: Unit ID 11 ]  │
 └──────────────────────────────┬──────────────────────────────┘
                                │ Single persistent TCP connection
                                ▼
 ┌─────────────────────────────────────────────────────────────┐
 │                sma-modbus-go (DAEMON / SERVICE)             │
 │                                                             │
 │  • Background Poller goroutine (every 2s)                   │
 │  • Thread-safe In-Memory Store (sync.RWMutex)               │
 │  • Server-Sent Events Broadcaster                           │
 │  • HTTP REST API Router & Embedded Web Dashboard            │
 └──────────────────────────────┬──────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
 [ REST Endpoints ]    [ Realtime SSE Stream ]   [ Web Dashboard ]
  /api/data               /api/stream             http://localhost:8050
  /api/plant              (continuous emitter)
  /api/inverters
```

---

## 📦 Using as a Go Library (`pkg/smamodbus`)

Install package:
```bash
go get github.com/samuelcampozano/sma-modbus-go
```

### Example: Read Telemetry from SMA Inverters

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/samuelcampozano/sma-modbus-go/pkg/smamodbus"
)

func main() {
	// Connect to SMA Data Manager or Inverter IP
	client := smamodbus.NewClient("192.168.0.100:502", 3*time.Second)
	defer client.Close()

	// 1. High-level telemetry query (Unit ID 10)
	telemetry, err := client.ReadTelemetry(10)
	if err != nil {
		log.Fatalf("Error querying inverter: %v", err)
	}

	fmt.Printf("Inverter Status:  %s\n", telemetry.OperatingStatus)
	fmt.Printf("Active Power:     %d W (%.2f kW)\n", telemetry.ActivePowerW, telemetry.ActivePowerKW)
	fmt.Printf("Phase L1 / L2 / L3: %.2f / %.2f / %.2f kW\n", telemetry.PL1KW, telemetry.PL2KW, telemetry.PL3KW)
	fmt.Printf("Daily Production: %.2f kWh\n", telemetry.DailyYieldKWh)
	fmt.Printf("Lifetime Total:   %.2f MWh\n", telemetry.TotalEnergyMWh)

	// 2. Query individual metrics
	watts, kw, err := client.ReadActivePower(11) // Inverter #2
	if err == nil {
		fmt.Printf("Inverter 11 Power: %d W (%.2f kW)\n", watts, kw)
	}
}
```

---

## 🚀 Running as a Standalone Daemon / CLI

### 1. Build and Run
```powershell
# Run tests
go test -v ./...

# Build binary
go build -ldflags="-s -w" -o solar_api.exe ./cmd/solar-api

# Start daemon
.\solar_api.exe
```

### 2. Configuration Flags and Environment Variables

| Flag | Env Variable | Default | Description |
| :--- | :--- | :--- | :--- |
| `-port` | `PORT` | `8050` | HTTP API and dashboard port |
| `-modbus-host` | `MODBUS_HOST` | `192.168.0.100:502` | SMA Modbus TCP target host and port |
| `-interval` | `POLL_INTERVAL` | `2s` | Modbus polling interval |
| `-web-dir` | `WEB_DIR` | `./web` | Static dashboard directory |

---

## 📡 API Reference & `curl` Examples

### 1. Full Snapshot: `GET /api/data`
Returns aggregated plant metrics and the complete state of every inverter:
```bash
curl -s http://localhost:8050/api/data
```
```json
{
  "timestamp": "2026-08-26 14:47:36",
  "simulation_mode": false,
  "total_power_kw": 47.74,
  "total_daily_kwh": 712.29,
  "total_accum_mwh": 1354.67,
  "grid_frequency_hz": 60,
  "inverters": [
    {
      "id": "inv1",
      "name": "SMA Sunny Highpower #1",
      "ip": "192.168.0.100:502",
      "unit": 10,
      "connected": true,
      "error": null,
      "operating_status": "OK",
      "power_w": 25203,
      "power_kw": 25.2,
      "p_l1_kw": 8.4,
      "p_l2_kw": 8.4,
      "p_l3_kw": 8.4,
      "frequency_hz": 60,
      "daily_kwh": 363,
      "total_mwh": 678.84,
      "temperature_c": 46.2,
      "serial": "SHP-U10",
      "last_seen": "14:47:36"
    }
  ]
}
```

### 2. Live Continuous Stream (SSE): `GET /api/stream`
Keeps connection open and streams real-time updates line-by-line as data arrives from hardware:
```bash
curl -N http://localhost:8050/api/stream
```
```
data: {"timestamp":"2026-08-26 14:47:36", "total_power_kw":47.74, ...}

data: {"timestamp":"2026-08-26 14:47:38", "total_power_kw":47.81, ...}
```

### 3. Specialized Endpoints

| Endpoint | Method | Description | Example Curl |
| :--- | :---: | :--- | :--- |
| `/api/plant` | `GET` | High-level summary of plant totals | `curl -s http://localhost:8050/api/plant` |
| `/api/inverters` | `GET` | List of all inverter telemetry objects | `curl -s http://localhost:8050/api/inverters` |
| `/api/inverters/10` | `GET` | Specific inverter telemetry by Unit ID | `curl -s http://localhost:8050/api/inverters/10` |
| `/api/health` | `GET` | System health, uptime & Modbus status | `curl -s http://localhost:8050/api/health` |
| `/` | `GET` | HTML5 Responsive Dashboard | Open in browser: `http://localhost:8050` |

---

## 🗃️ SMA Modbus Register Map

| Register | Name | Data Type | Words | Unit | Description |
| :---: | :--- | :---: | :---: | :---: | :--- |
| **`30051`** | `DeviceClass` | U32 | 2 | - | `8001` = Solar Inverter, `8128` = Data Manager |
| **`30201`** | `OperatingStatus` | U32 | 2 | - | `307` = OK, `455` = Warning, `35` = Fault, `303` = Off |
| **`30233`** | `ConnectedPower` | U32 | 2 | W | Nominal connected power capacity (e.g. 125,000 W) |
| **`30513`** | `TotalEnergyFedIn` | U64 | 4 | Wh | Lifetime total feed-in electrical energy |
| **`30517`** | `DailyYield` | U64 | 4 | Wh | Current day energy production |
| **`30775`** | `ActivePowerTotal` | S32 | 2 | W | Instantaneous active power across all phases |
| **`30803`** | `GridFrequency` | U32 | 2 | Hz * 100 | AC grid frequency |
| **`30805`** | `ReactivePower` | S32 | 2 | VAr | Reactive power |

---

## 🧪 Testing

Run all unit tests across packages:
```powershell
go test -v ./...
```

Output:
```
=== RUN   TestDecodeS32
--- PASS: TestDecodeS32 (0.00s)
=== RUN   TestDecodeU32
--- PASS: TestDecodeU32 (0.00s)
=== RUN   TestDecodeU64
--- PASS: TestDecodeU64 (0.00s)
=== RUN   TestDecodeOperatingStatus
--- PASS: TestDecodeOperatingStatus (0.00s)
ok      solar-api/pkg/smamodbus 0.4s
ok      solar-api/internal/api  0.3s
```

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
