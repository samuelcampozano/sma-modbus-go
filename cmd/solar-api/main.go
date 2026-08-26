package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/samuelcampozano/sma-modbus-go/internal/api"
	"github.com/samuelcampozano/sma-modbus-go/internal/config"
	"github.com/samuelcampozano/sma-modbus-go/internal/modbus"
)

func main() {
	cfg := config.Load()

	// Initialize thread-safe telemetry store
	store := modbus.NewStore(cfg)

	// Context for poller & graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background Modbus poller
	go store.StartPoller(ctx)

	// Setup HTTP handler
	handler := api.NewHandler(store)
	router := api.SetupRouter(handler, cfg.WebDir)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // Disabled to support indefinite SSE streaming
		IdleTimeout:  60 * time.Second,
	}

	// Print professional startup banner
	fmt.Println("==================================================================")
	fmt.Println(" ☀️  MANACRIPEX SOLAR MONITORING API & DASHBOARD (GOLANG SERVICE) ")
	fmt.Println("==================================================================")
	fmt.Printf(" 📡 Modbus Source:    %s\n", cfg.ModbusHost)
	fmt.Printf(" ⏱️  Poll Interval:   %v\n", cfg.PollInterval)
	fmt.Printf(" ⚙️  Inverters:       Unit %d & Unit %d\n", cfg.Inverters[0].Unit, cfg.Inverters[1].Unit)
	fmt.Printf(" 🌐 Dashboard Web:    http://localhost:%d\n", cfg.ServerPort)
	fmt.Printf(" 📊 REST Snapshot:    http://localhost:%d/api/data\n", cfg.ServerPort)
	fmt.Printf(" ⚡ SSE Stream:       http://localhost:%d/api/stream\n", cfg.ServerPort)
	fmt.Printf(" 🩺 Healthcheck:      http://localhost:%d/api/health\n", cfg.ServerPort)
	fmt.Println("==================================================================")

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down gracefully...")

	// Cancel background polling
	cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("Server gracefully stopped.")
	}
}
