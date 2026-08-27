package main

import (
	"fmt"
	"log"
	"time"

	"github.com/samuelcampozano/sma-modbus-go/pkg/smamodbus"
)

func main() {
	modbusHost := "192.168.0.100:502"
	fmt.Printf("Connecting to SMA Data Manager at %s...\n\n", modbusHost)

	client := smamodbus.NewClient(modbusHost, 3*time.Second)
	defer client.Close()

	inverterUnits := []struct {
		Name string
		Unit uint8
	}{
		{"SMA Sunny Highpower #1", 10},
		{"SMA Sunny Highpower #2", 11},
	}

	for _, inv := range inverterUnits {
		fmt.Println("=======================================================")
		fmt.Printf(" Querying %s (Unit %d)\n", inv.Name, inv.Unit)
		fmt.Println("=======================================================")

		telemetry, err := client.ReadTelemetry(inv.Unit)
		if err != nil {
			log.Printf("❌ Error reading Unit %d: %v\n\n", inv.Unit, err)
			continue
		}

		fmt.Printf("🔹 Operating Status : %s\n", telemetry.OperatingStatus)
		fmt.Printf("⚡ Active Power     : %d W (%.2f kW)\n", telemetry.ActivePowerW, telemetry.ActivePowerKW)
		fmt.Printf("   L1: %.2f kW  |  L2: %.2f kW  |  L3: %.2f kW\n", telemetry.PL1KW, telemetry.PL2KW, telemetry.PL3KW)
		fmt.Printf("🌐 Grid Frequency   : %.2f Hz\n", telemetry.FrequencyHz)
		fmt.Printf("☀️ Daily Yield       : %.2f kWh\n", telemetry.DailyYieldKWh)
		fmt.Printf("📊 Total Energy     : %.2f MWh\n", telemetry.TotalEnergyMWh)
		fmt.Printf("🌡️ Temp (Estimated) : %.1f °C\n\n", telemetry.TemperatureC)
	}

	fmt.Println("Telemetry poll completed successfully.")
}
