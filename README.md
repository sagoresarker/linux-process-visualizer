
A minimal system monitoring tool written in Go, similar to htop. This tool provides real-time system metrics including CPU usage, memory usage, and process information.

## Features

- Real-time CPU usage monitoring (total and per-core)
- Memory usage statistics
- Process list with details (PID, name, CPU%, memory usage, priority, state)
- Terminal user interface
- Sort processes by CPU usage
- Updates every second


## Installation

```bash
git clone https://github.com/sagoresarker/linux-process-visualizer.git
cd linux-process-visualizer
go build ./cmd/linux-process-visualizer
```

## Usage

```bash
./linux-process-visualizer
```

## Building from Source

```bash
go build -o gohtop ./cmd/linux-process-visualizer
```
