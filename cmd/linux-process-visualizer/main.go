package main

import (
	"github.com/sagoresarker/linux-process-visualizer/internal/display"
	"github.com/sagoresarker/linux-process-visualizer/internal/metrics"
	"log"
	"time"
)

func main() {
	// Initialize the TUI
	tui, err := display.NewTUI()
	if err != nil {
		log.Fatalf("Failed to initialize TUI: %v", err)
	}
	defer tui.Close()

	// Create metrics collector
	collector := metrics.NewCollector()

	// Update loop
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := collector.Collect()
			tui.Update(stats)
		case event := <-tui.Events():
			if event.Type == display.EventQuit {
				return
			}
		}
	}
}
