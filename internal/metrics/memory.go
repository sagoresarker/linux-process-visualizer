package metrics

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type MemoryStats struct {
	Total     uint64
	Used      uint64
	Free      uint64
	Shared    uint64
	Buffers   uint64
	Cached    uint64
	Available uint64
}

// getMemoryStats gets the memory stats
func (c *Collector) getMemoryStats() MemoryStats {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemoryStats{}
	}
	defer file.Close()

	var stats MemoryStats
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		// Convert from KB to bytes
		value *= 1024

		switch fields[0] {
		case "MemTotal:":
			stats.Total = value
		case "MemFree:":
			stats.Free = value
		case "MemAvailable:":
			stats.Available = value
		case "Buffers:":
			stats.Buffers = value
		case "Cached:":
			stats.Cached = value
		case "Shmem:":
			stats.Shared = value
		}
	}

	stats.Used = stats.Total - stats.Free - stats.Buffers - stats.Cached

	return stats
}
