package main

import (
	"log"
	"os"
	"time"

	"github.com/sagoresarker/linux-process-visualizer/internal/display"
	"github.com/sagoresarker/linux-process-visualizer/internal/metrics"
)

func main() {
	logFile, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Starting Linux Process Visualizer...")

	// Initialize the TUI
	tui, err := display.NewTUI()
	if err != nil {
		log.Fatalf("Failed to initialize TUI: %v", err)
	}
	defer tui.Close()

	// Create metrics collector
	collector := metrics.NewCollector()
	log.Println("Metrics collector initialized")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	log.Println("Entering main update loop")
	for {
		select {
		case <-ticker.C:
			stats := collector.Collect()
			tui.Update(stats)
		case event := <-tui.Events():
			if event.Type == display.EventQuit {
				log.Println("Quit event received, shutting down...")
				return
			}
		}
	}
}
