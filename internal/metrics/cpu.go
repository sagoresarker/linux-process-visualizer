package metrics

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type CPUStats struct {
	Usage     float64
	PerCPU    []float64
	Total     CPUTimes
	PrevTotal CPUTimes
}

type CPUTimes struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	Iowait  uint64
	Irq     uint64
	Softirq uint64
	Steal   uint64
	Guest   uint64
}

// getCPUStats gets the CPU stats
func (c *Collector) getCPUStats() CPUStats {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return CPUStats{}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	stats := CPUStats{
		PerCPU: make([]float64, 0),
	}
	var cpuIndex int

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		if fields[0] == "cpu" {
			stats.Total = parseCPUTimes(fields[1:])
			stats.Usage = calculateCPUUsage(stats.Total, c.prevCPUStats.Total)
		} else if strings.HasPrefix(fields[0], "cpu") {
			times := parseCPUTimes(fields[1:])
			if cpuIndex < len(c.prevCPUStats.PerCPU) {
				usage := calculateCPUUsage(times, c.prevCPUStats.PrevTotal)
				stats.PerCPU = append(stats.PerCPU, usage)
			} else {
				stats.PerCPU = append(stats.PerCPU, 0.0)
			}
			cpuIndex++
		}
	}

	return stats
}

// parseCPUTimes parses the CPU times
func parseCPUTimes(fields []string) CPUTimes {
	var times CPUTimes
	if len(fields) >= 8 {
		times.User, _ = strconv.ParseUint(fields[0], 10, 64)
		times.Nice, _ = strconv.ParseUint(fields[1], 10, 64)
		times.System, _ = strconv.ParseUint(fields[2], 10, 64)
		times.Idle, _ = strconv.ParseUint(fields[3], 10, 64)
		times.Iowait, _ = strconv.ParseUint(fields[4], 10, 64)
		times.Irq, _ = strconv.ParseUint(fields[5], 10, 64)
		times.Softirq, _ = strconv.ParseUint(fields[6], 10, 64)
		times.Steal, _ = strconv.ParseUint(fields[7], 10, 64)
	}
	return times
}

// calculateCPUUsage calculates the CPU usage
func calculateCPUUsage(current, previous CPUTimes) float64 {
	prevIdle := previous.Idle + previous.Iowait
	idle := current.Idle + current.Iowait

	prevNonIdle := previous.User + previous.Nice + previous.System +
		previous.Irq + previous.Softirq + previous.Steal
	nonIdle := current.User + current.Nice + current.System +
		current.Irq + current.Softirq + current.Steal

	prevTotal := prevIdle + prevNonIdle
	total := idle + nonIdle

	totalDelta := total - prevTotal
	idleDelta := idle - prevIdle

	if totalDelta == 0 {
		return 0.0
	}

	return (float64(totalDelta-idleDelta) / float64(totalDelta)) * 100.0
}
