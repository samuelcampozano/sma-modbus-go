# ☀️ sma-modbus-go

[![Versión Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Licencia: MIT](https://img.shields.io/badge/Licencia-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Protocolo](https://img.shields.io/badge/Protocolo-Modbus_TCP_(Puerto_502)-orange)](https://files.sma.de)
[![Pruebas](https://img.shields.io/badge/Pruebas-Pasando-brightgreen)]()

> Librería y servicio de telemetría de alto rendimiento en **Golang (sin dependencias externas)** para inversores solares **SMA Sunny Highpower** y concentradores **SMA Data Manager M (ennexOS)** a través de Modbus TCP (Puerto 502).

*Read this in other languages: [📖 English](README.md)*

---

## 🌟 Características Principales

* **Cero Dependencias Externas**: Desarrollado exclusivamente con la biblioteca estándar de Go (`net`, `encoding/binary`, `net/http`, `sync`). Compilación instantánea y binario ultra ligero (~8 MB).
* **Arquitectura Dual**:
  * **Como Librería Go (`pkg/smamodbus`)**: Impórtala directamente en tus propios microservicios, programas o colectores SCADA.
  * **Como Servicio Autónomo (`solar_api.exe`)**: Ejecuta un poller periódico, caché en memoria thread-safe, dashboard web HTML5 integrado y API REST.
* **Emisión en Tiempo Real**: Endpoint nativo **Server-Sent Events (SSE)** en `/api/stream` que emite actualizaciones continuas en cada ciclo de lectura (`curl -N`).
* **Decodificador Nativo SMA**:
  * Desempaquetado Big-Endian para enteros de 32 bits (`S32`/`U32`) y 64 bits (`U64`).
  * Filtrado de centinelas de error de SMA (`-0x80000000`, `0xFFFFFFFF`) para evitar números basura durante la noche o fallas de sensor.
  * Traducción automática de códigos de estado de SMA (`307` = OK, `455` = Advertencia, `35` = Falla).
* **Ruteo Automático por Gateway**: Compatible con arquitecturas SMA Data Manager M / ennexOS consultando múltiples inversores por sus Unit IDs (Slave IDs) en una sola conexión TCP.

---

## 📐 Arquitectura

```
 ┌─────────────────────────────────────────────────────────────┐
 │           SMA DATA MANAGER M / ennexOS (GATEWAY)            │
 │                     IP: 192.168.0.100:502                   │
 │                                                             │
 │   [ Inversor #1: Unit ID 10 ]   [ Inversor #2: Unit ID 11 ]  │
 └──────────────────────────────┬──────────────────────────────┘
                                │ 1 sola conexión TCP persistente
                                ▼
 ┌─────────────────────────────────────────────────────────────┐
 │                sma-modbus-go (SERVICIO / DAEMON)            │
 │                                                             │
 │  • Poller en segundo plano con goroutines (cada 2s)         │
 │  • Almacén de telemetría seguro en memoria (sync.RWMutex)   │
 │  • Emisor de eventos Server-Sent Events (Broadcaster)       │
 │  • Enrutador HTTP REST y Dashboard Web integrado            │
 └──────────────────────────────┬──────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
 [ Endpoints REST ]     [ Stream en Vivo (SSE) ]  [ Dashboard Web ]
  /api/data               /api/stream              http://localhost:8050
  /api/plant              (emisión continua)
  /api/inverters
```

---

## 📦 Uso como Librería en Go (`pkg/smamodbus`)

Instalar el paquete:
```bash
go get github.com/samuelcampozano/sma-modbus-go
```

### Ejemplo: Leer Telemetría de los Inversores SMA

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/samuelcampozano/sma-modbus-go/pkg/smamodbus"
)

func main() {
	// Conectar a la IP del Data Manager M o del Inversor
	client := smamodbus.NewClient("192.168.0.100:502", 3*time.Second)
	defer client.Close()

	// 1. Consulta de alto nivel de telemetría (Inversor #1 - Unit ID 10)
	telemetry, err := client.ReadTelemetry(10)
	if err != nil {
		log.Fatalf("Error consultando inversor: %v", err)
	}

	fmt.Printf("Estado Inversor:  %s\n", telemetry.OperatingStatus)
	fmt.Printf("Potencia Activa:  %d W (%.2f kW)\n", telemetry.ActivePowerW, telemetry.ActivePowerKW)
	fmt.Printf("Fases L1/L2/L3:   %.2f / %.2f / %.2f kW\n", telemetry.PL1KW, telemetry.PL2KW, telemetry.PL3KW)
	fmt.Printf("Producción Hoy:   %.2f kWh\n", telemetry.DailyYieldKWh)
	fmt.Printf("Energía Total:    %.2f MWh\n", telemetry.TotalEnergyMWh)

	// 2. Consulta de métricas individuales
	watts, kw, err := client.ReadActivePower(11) // Inversor #2 - Unit ID 11
	if err == nil {
		fmt.Printf("Potencia Inversor 11: %d W (%.2f kW)\n", watts, kw)
	}
}
```

---

## 🚀 Uso como Servicio / CLI Autónomo

### 1. Compilación y Ejecución
```powershell
# Ejecutar pruebas unitarias
go test -v ./...

# Compilar binario de producción
go build -ldflags="-s -w" -o solar_api.exe ./cmd/solar-api

# Iniciar servicio
.\solar_api.exe
```

### 2. Parámetros de Configuración y Variables de Entorno

| Parámetro (Flag) | Variable de Entorno | Valor por Defecto | Descripción |
| :--- | :--- | :--- | :--- |
| `-port` | `PORT` | `8050` | Puerto HTTP del servicio y dashboard |
| `-modbus-host` | `MODBUS_HOST` | `192.168.0.100:502` | Dirección host:puerto del servidor Modbus TCP |
| `-interval` | `POLL_INTERVAL` | `2s` | Intervalo de sondeo Modbus |
| `-web-dir` | `WEB_DIR` | `./web` | Directorio de la interfaz web estática |

---

## 📡 Referencia de la API y Ejemplos con `curl`

### 1. Instantánea Completa: `GET /api/data`
Retorna las métricas globales agregadas y el detalle de cada inversor en JSON:
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

### 2. Emisión Continua en Tiempo Real (SSE Stream): `GET /api/stream`
Mantiene la conexión abierta y transmite eventos línea por línea a medida que los datos son leídos del hardware:
```bash
curl -N http://localhost:8050/api/stream
```
```
data: {"timestamp":"2026-08-26 14:47:36", "total_power_kw":47.74, ...}

data: {"timestamp":"2026-08-26 14:47:38", "total_power_kw":47.81, ...}
```

### 3. Endpoints Especializados

| Endpoint | Método | Descripción | Ejemplo de Curl |
| :--- | :---: | :--- | :--- |
| `/api/plant` | `GET` | Resumen de totales de la planta | `curl -s http://localhost:8050/api/plant` |
| `/api/inverters` | `GET` | Lista de todos los inversores | `curl -s http://localhost:8050/api/inverters` |
| `/api/inverters/10` | `GET` | Datos de un inversor por su Unit ID | `curl -s http://localhost:8050/api/inverters/10` |
| `/api/health` | `GET` | Diagnóstico de salud, uptime y Modbus | `curl -s http://localhost:8050/api/health` |
| `/` | `GET` | Dashboard Web Responsive en HTML5 | Abrir en navegador: `http://localhost:8050` |

---

## 🗃️ Tabla de Registros Modbus SMA

| Registro | Nombre | Tipo | Palabras | Unidad | Descripción |
| :---: | :--- | :---: | :---: | :---: | :--- |
| **`30051`** | `DeviceClass` | U32 | 2 | - | `8001` = Inversor fotovoltaico, `8128` = Data Manager |
| **`30201`** | `OperatingStatus` | U32 | 2 | - | `307` = OK, `455` = Advertencia, `35` = Falla, `303` = Apagado |
| **`30233`** | `ConnectedPower` | U32 | 2 | W | Capacidad nominal conectada (ej. 125,000 W) |
| **`30513`** | `TotalEnergyFedIn` | U64 | 4 | Wh | Energía total inyectada a la red acumulada histórica |
| **`30517`** | `DailyYield` | U64 | 4 | Wh | Producción de energía del día actual |
| **`30775`** | `ActivePowerTotal` | S32 | 2 | W | Potencia activa instantánea total |
| **`30803`** | `GridFrequency` | U32 | 2 | Hz * 100 | Frecuencia de red eléctrica de corriente alterna |
| **`30805`** | `ReactivePower` | S32 | 2 | VAr | Potencia reactiva total |

---

## 🧪 Pruebas Automatizadas

Ejecuta todas las pruebas unitarias:
```powershell
go test -v ./...
```

---

## 👤 Autor y Mantenedor

**Samuel Campozano Lopez**  
*Software Engineer & Systems Architect*

* 💼 **LinkedIn:** [linkedin.com/in/samuel-campozano-lopez](https://www.linkedin.com/in/samuel-campozano-lopez/?locale=es)
* 🐙 **GitHub:** [@samuelcampozano](https://github.com/samuelcampozano)

---

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Consulta el archivo [LICENSE](LICENSE) para más detalles.
