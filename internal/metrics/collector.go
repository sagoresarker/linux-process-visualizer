package metrics

type Collector struct {
	prevCPUStats CPUStats
}

type SystemStats struct {
	CPU     CPUStats
	Memory  MemoryStats
	Process []ProcessInfo
}

func NewCollector() *Collector {
	return &Collector{}
}

// Collect collects the system stats
func (c *Collector) Collect() SystemStats {
	stats := SystemStats{
		CPU:     c.getCPUStats(),
		Memory:  c.getMemoryStats(),
		Process: c.getProcessStats(),
	}

	c.prevCPUStats = stats.CPU

	return stats
}
