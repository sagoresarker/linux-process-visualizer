package metrics

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ProcessInfo struct {
	PID      int
	Name     string
	State    string
	Memory   uint64
	CPU      float64
	Command  string
	User     string
	Priority int
}

// getProcessStats gets the process stats
func (c *Collector) getProcessStats() []ProcessInfo {
	processes := []ProcessInfo{}

	// Read all directories in /proc
	dirs, err := os.ReadDir("/proc")
	if err != nil {
		return processes
	}

	for _, dir := range dirs {
		// Check if the directory name is a number (PID)
		pid, err := strconv.Atoi(dir.Name())
		if err != nil {
			continue
		}

		process := c.readProcessInfo(pid)
		if process != nil {
			processes = append(processes, *process)
		}
	}

	return processes
}

// readProcessInfo reads the process info
func (c *Collector) readProcessInfo(pid int) *ProcessInfo {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)

	// Read stat file
	statFile, err := os.Open(statPath)
	if err != nil {
		return nil
	}
	defer statFile.Close()

	scanner := bufio.NewScanner(statFile)
	if !scanner.Scan() {
		return nil
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) < 24 {
		return nil
	}

	// Read command line
	cmdline, _ := os.ReadFile(cmdlinePath)
	cmd := strings.ReplaceAll(string(cmdline), "\x00", " ")
	if cmd == "" {
		cmd = fields[1]
	}

	// Parse process stats
	utime, _ := strconv.ParseUint(fields[13], 10, 64)
	stime, _ := strconv.ParseUint(fields[14], 10, 64)
	vsize, _ := strconv.ParseUint(fields[22], 10, 64)
	priority, _ := strconv.Atoi(fields[17])

	return &ProcessInfo{
		PID:      pid,
		Name:     strings.Trim(fields[1], "()"),
		State:    fields[2],
		Memory:   vsize,
		CPU:      float64(utime+stime) / float64(100), // Simple CPU usage calculation
		Command:  cmd,
		Priority: priority,
	}
}
