package display

import (
	"fmt"
	"log"
	"sort"

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

	// Initialize widgets with improved styling
	log.Println("Setting up widgets...")

	// CPU Box styling
	tui.cpuBox = tview.NewTextView()
	tui.cpuBox.SetDynamicColors(true)
	tui.cpuBox.SetBorder(true)
	tui.cpuBox.SetTitle("💻 CPU Usage")
	tui.cpuBox.SetTitleColor(tcell.ColorGreen)
	tui.cpuBox.SetBorderColor(tcell.ColorGreen)
	tui.cpuBox.SetTitleAlign(tview.AlignLeft)

	// Memory Box styling
	tui.memoryBox = tview.NewTextView()
	tui.memoryBox.SetDynamicColors(true)
	tui.memoryBox.SetBorder(true)
	tui.memoryBox.SetTitle("🧠 Memory Usage")
	tui.memoryBox.SetTitleColor(tcell.ColorBlue)
	tui.memoryBox.SetBorderColor(tcell.ColorBlue)
	tui.memoryBox.SetTitleAlign(tview.AlignLeft)

	// Process Table styling
	tui.processBox = tview.NewTable()
	tui.processBox.SetBorders(true)
	tui.processBox.SetBorder(true)
	tui.processBox.SetTitle("📊 Processes")
	tui.processBox.SetTitleColor(tcell.ColorYellow)
	tui.processBox.SetBorderColor(tcell.ColorYellow)
	tui.processBox.SetTitleAlign(tview.AlignLeft)

	// Create layout with better proportions
	grid := tview.NewGrid().
		SetRows(6, 6, 0). // Increased height for CPU and Memory boxes
		SetColumns(0).
		SetBorders(false).
		AddItem(tui.cpuBox, 0, 0, 1, 1, 0, 0, false).
		AddItem(tui.memoryBox, 1, 0, 1, 1, 0, 0, false).
		AddItem(tui.processBox, 2, 0, 1, 1, 0, 0, true) // Make process box focusable

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
		t.updateCPU(stats.CPU)

		t.updateMemory(stats.Memory)

		t.updateProcesses(stats.Process)
	})
}

// updateCPU updates the CPU display with improved formatting
func (t *TUI) updateCPU(cpu metrics.CPUStats) {
	t.cpuBox.Clear()
	fmt.Fprintf(t.cpuBox, "[white]Total CPU Usage: [red]%.1f%%[white]\n\n", cpu.Usage)

	// Display CPU cores in columns
	numCores := len(cpu.PerCPU)
	columns := 4 // Number of columns for CPU display
	for i := 0; i < numCores; i += columns {
		for j := 0; j < columns && i+j < numCores; j++ {
			usage := cpu.PerCPU[i+j]
			color := getColorForUsage(usage)
			fmt.Fprintf(t.cpuBox, "[white]CPU%-2d: [%s]%5.1f%%[white]    ", i+j, color, usage)
		}
		fmt.Fprintf(t.cpuBox, "\n")
	}
}

// updateMemory updates the memory display with improved formatting
func (t *TUI) updateMemory(mem metrics.MemoryStats) {
	t.memoryBox.Clear()
	totalGB := float64(mem.Total) / 1024 / 1024 / 1024
	usedGB := float64(mem.Used) / 1024 / 1024 / 1024
	freeGB := float64(mem.Free) / 1024 / 1024 / 1024
	cachedGB := float64(mem.Cached) / 1024 / 1024 / 1024
	buffersGB := float64(mem.Buffers) / 1024 / 1024 / 1024

	usagePercent := float64(mem.Used) / float64(mem.Total) * 100
	color := getColorForUsage(usagePercent)

	fmt.Fprintf(t.memoryBox, "[white]Total:   [blue]%.1f GB[white]\n", totalGB)
	fmt.Fprintf(t.memoryBox, "[white]Used:    [%s]%.1f GB[white] ([%s]%.1f%%[white])\n", color, usedGB, color, usagePercent)
	fmt.Fprintf(t.memoryBox, "[white]Free:    [green]%.1f GB[white]\n", freeGB)
	fmt.Fprintf(t.memoryBox, "[white]Cached:  [yellow]%.1f GB[white]\n", cachedGB)
	fmt.Fprintf(t.memoryBox, "[white]Buffers: [yellow]%.1f GB[white]\n", buffersGB)
}

// updateProcesses updates the process list with improved formatting
func (t *TUI) updateProcesses(processes []metrics.ProcessInfo) {
	t.processBox.Clear()

	// Set up headers with improved styling
	headers := []string{"PID", "Name", "CPU%", "Memory", "Priority", "State"}
	colors := []tcell.Color{
		tcell.ColorYellow,
		tcell.ColorYellow,
		tcell.ColorYellow,
		tcell.ColorYellow,
		tcell.ColorYellow,
		tcell.ColorYellow,
	}
	for i, header := range headers {
		t.processBox.SetCell(0, i,
			tview.NewTableCell(header).
				SetTextColor(colors[i]).
				SetSelectable(false).
				SetAlign(tview.AlignCenter).
				SetExpansion(1))
	}

	// Sort processes by CPU usage (using sort.Slice for better performance)
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].CPU > processes[j].CPU
	})

	// Add processes with alternating row colors
	for i, proc := range processes {
		if i >= 100 { // Show only top 100 processes
			break
		}
		row := i + 1
		bgColor := tcell.ColorDefault
		if i%2 == 0 {
			bgColor = tcell.ColorDarkSlateGray
		}

		cpuColor := getColorForUsage(proc.CPU)

		cells := []struct {
			text  string
			color tcell.Color
		}{
			{fmt.Sprintf("%d", proc.PID), tcell.ColorWhite},
			{proc.Name, tcell.ColorWhite},
			{fmt.Sprintf("%.1f", proc.CPU), tcell.GetColor(cpuColor)},
			{formatMemory(proc.Memory), tcell.ColorWhite},
			{fmt.Sprintf("%d", proc.Priority), tcell.ColorWhite},
			{proc.State, tcell.ColorWhite},
		}

		for col, cell := range cells {
			t.processBox.SetCell(row, col,
				tview.NewTableCell(cell.text).
					SetTextColor(cell.color).
					SetBackgroundColor(bgColor).
					SetAlign(tview.AlignCenter).
					SetExpansion(1))
		}
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

// Helper function to determine color based on usage percentage
func getColorForUsage(usage float64) string {
	switch {
	case usage >= 90:
		return "red"
	case usage >= 70:
		return "orange"
	case usage >= 50:
		return "yellow"
	default:
		return "green"
	}
}
