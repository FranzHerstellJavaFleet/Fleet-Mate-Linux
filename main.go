package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/javafleet/fleet-mate-linux/internal/config"
	"github.com/javafleet/fleet-mate-linux/internal/hardware"
	"github.com/javafleet/fleet-mate-linux/internal/websocket"
)

const (
	version = "1.1.0"
)

func main() {
	// Command line flags
	configFile := flag.String("config", "config.yml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		log.Printf("Fleet Mate Linux v%s", version)
		os.Exit(0)
	}

	log.Printf("Fleet Mate Linux v%s starting...", version)

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded from %s", *configFile)
	log.Printf("Mate ID: %s", cfg.Mate.ID)
	log.Printf("Mate Name: %s", cfg.Mate.Name)
	log.Printf("Navigator URL: %s", cfg.Navigator.URL)
	log.Printf("Monitoring interval: %s", cfg.Monitoring.Interval)

	// Create hardware monitor
	monitor := hardware.NewMonitor(cfg)
	log.Println("Hardware monitor initialized")

	// Create WebSocket client
	client := websocket.NewClient(cfg, monitor)
	log.Println("WebSocket client initialized")

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start client in background
	go func() {
		if err := client.Run(); err != nil {
			log.Printf("Client error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)

	// Graceful shutdown
	client.Stop()
	log.Println("Fleet Mate stopped")
}
