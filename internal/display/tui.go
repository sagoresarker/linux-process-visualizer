package display

import (
	"fmt"
	"log"

	"github.com/sagoresarker/linux-process-visualizer/internal/metrics"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUI struct {
	app        *tview.Application
	cpuBox     *tview.TextView
	memoryBox  *tview.TextView
	processBox *tview.Table
	events     chan Event
}

type EventType int

const (
	EventQuit EventType = iota
)

type Event struct {
	Type EventType
}

// NewTUI initializes a new TUI
func NewTUI() (*TUI, error) {
	log.Println("Initializing TUI...")
	tui := &TUI{
		app:    tview.NewApplication(),
		events: make(chan Event),
	}

	// Initialize widgets
	log.Println("Setting up widgets...")
	// For TextView widgets
	tui.cpuBox = tview.NewTextView()
	tui.cpuBox.SetDynamicColors(true)
	tui.cpuBox.SetTitle("CPU Usage")
	tui.cpuBox.SetBorder(true)

	tui.memoryBox = tview.NewTextView()
	tui.memoryBox.SetDynamicColors(true)
	tui.memoryBox.SetTitle("Memory Usage")
	tui.memoryBox.SetBorder(true)

	// For Table widget
	tui.processBox = tview.NewTable()
	tui.processBox.SetBorders(true)
	tui.processBox.SetTitle("Processes")
	tui.processBox.SetBorder(true)

	// Create layout
	grid := tview.NewGrid().
		SetRows(3, 3, 0).
		SetColumns(0).
		AddItem(tui.cpuBox, 0, 0, 1, 1, 0, 0, false).
		AddItem(tui.memoryBox, 1, 0, 1, 1, 0, 0, false).
		AddItem(tui.processBox, 2, 0, 1, 1, 0, 0, false)

	// Set up input handling
	tui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc, tcell.KeyCtrlC:
			log.Println("Quit signal received")
			tui.events <- Event{Type: EventQuit}
		}
		return event
	})

	tui.app.SetRoot(grid, true)

	// Start the application in a separate goroutine
	go func() {
		log.Println("Starting TUI application...")
		if err := tui.app.Run(); err != nil {
			log.Printf("Error running application: %v\n", err)
			tui.events <- Event{Type: EventQuit}
		}
	}()

	log.Println("TUI initialization complete")
	return tui, nil
}

func (t *TUI) Close() {
	log.Println("Closing TUI...")
	t.app.Stop()
}

func (t *TUI) Events() chan Event {
	return t.events
}

func (t *TUI) Update(stats metrics.SystemStats) {
	t.app.QueueUpdateDraw(func() {
		log.Println("Updating TUI displays...")
		// Update CPU display
		t.updateCPU(stats.CPU)

		// Update memory display
		t.updateMemory(stats.Memory)

		// Update process list
		t.updateProcesses(stats.Process)
	})
}

// updateCPU updates the CPU display
func (t *TUI) updateCPU(cpu metrics.CPUStats) {
	t.cpuBox.Clear()
	fmt.Fprintf(t.cpuBox, "Total CPU Usage: %.1f%%\n", cpu.Usage)

	for i, usage := range cpu.PerCPU {
		fmt.Fprintf(t.cpuBox, "CPU%d: %.1f%%\n", i, usage)
	}
}

// updateMemory updates the memory display
func (t *TUI) updateMemory(mem metrics.MemoryStats) {
	t.memoryBox.Clear()
	totalGB := float64(mem.Total) / 1024 / 1024 / 1024
	usedGB := float64(mem.Used) / 1024 / 1024 / 1024

	fmt.Fprintf(t.memoryBox, "Total: %.1f GB\n", totalGB)
	fmt.Fprintf(t.memoryBox, "Used:  %.1f GB (%.1f%%)\n", usedGB,
		float64(mem.Used)/float64(mem.Total)*100)
}

// updateProcesses updates the process list
func (t *TUI) updateProcesses(processes []metrics.ProcessInfo) {
	t.processBox.Clear()

	// Set headers
	headers := []string{"PID", "Name", "CPU%", "Memory", "Priority", "State"}
	for i, header := range headers {
		t.processBox.SetCell(0, i,
			tview.NewTableCell(header).
				SetTextColor(tcell.ColorYellow).
				SetSelectable(false))
	}

	// Sort processes by CPU usage (simple bubble sort)
	for i := 0; i < len(processes)-1; i++ {
		for j := 0; j < len(processes)-i-1; j++ {
			if processes[j].CPU < processes[j+1].CPU {
				processes[j], processes[j+1] = processes[j+1], processes[j]
			}
		}
	}

	// Fill process data
	for i, proc := range processes {
		if i >= 100 { // Show only top 100 processes
			break
		}
		row := i + 1
		t.processBox.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("%d", proc.PID)))
		t.processBox.SetCell(row, 1, tview.NewTableCell(proc.Name))
		t.processBox.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%.1f", proc.CPU)))
		t.processBox.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%.1f", proc.CPU)))
		t.processBox.SetCell(row, 3, tview.NewTableCell(formatMemory(proc.Memory)))
		t.processBox.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%d", proc.Priority)))
		t.processBox.SetCell(row, 5, tview.NewTableCell(proc.State))
	}
}

// formatMemory formats the memory
func formatMemory(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fG", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fM", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fK", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
