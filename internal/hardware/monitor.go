package hardware

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/javafleet/fleet-mate-linux/internal/config"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Stats represents all hardware statistics
type Stats struct {
	Timestamp   time.Time          `json:"timestamp"`
	MateID      string             `json:"mate_id"`
	CPU         *CPUStats          `json:"cpu,omitempty"`
	Memory      *MemoryStats       `json:"memory,omitempty"`
	Disk        []DiskStats        `json:"disk,omitempty"`
	Temperature *TemperatureStats  `json:"temperature,omitempty"`
	Network     []NetworkStats     `json:"network,omitempty"`
	GPU         []GPUStats         `json:"gpu,omitempty"`
	System      *SystemStats       `json:"system,omitempty"`
}

// CPUStats contains CPU information
type CPUStats struct {
	UsagePercent float64   `json:"usage_percent"`
	PerCore      []float64 `json:"per_core,omitempty"`
	Cores        int       `json:"cores"`
	Model        string    `json:"model"`
	MHz          float64   `json:"mhz"`
}

// MemoryStats contains memory information
type MemoryStats struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	SwapTotal   uint64  `json:"swap_total,omitempty"`
	SwapUsed    uint64  `json:"swap_used,omitempty"`
	SwapPercent float64 `json:"swap_percent,omitempty"`
}

// DiskStats contains disk information
type DiskStats struct {
	MountPoint  string  `json:"mount_point"`
	Device      string  `json:"device"`
	FSType      string  `json:"fs_type"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

// TemperatureStats contains temperature information
type TemperatureStats struct {
	Sensors []SensorTemp `json:"sensors"`
}

// SensorTemp represents a temperature sensor
type SensorTemp struct {
	Name        string  `json:"name"`
	Temperature float64 `json:"temperature"`
	High        float64 `json:"high,omitempty"`
	Critical    float64 `json:"critical,omitempty"`
}

// NetworkStats contains network interface information
type NetworkStats struct {
	Interface   string `json:"interface"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	Errin       uint64 `json:"errin"`
	Errout      uint64 `json:"errout"`
}

// SystemStats contains system information
type SystemStats struct {
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	Uptime          uint64 `json:"uptime"`
}

// GPUStats contains GPU information
type GPUStats struct {
	Index            int     `json:"index"`
	Name             string  `json:"name"`
	UtilizationGPU   float64 `json:"utilization_gpu"`
	MemoryTotal      uint64  `json:"memory_total"`      // in MB
	MemoryUsed       uint64  `json:"memory_used"`       // in MB
	MemoryFree       uint64  `json:"memory_free"`       // in MB
	MemoryUsedPercent float64 `json:"memory_used_percent"`
	Temperature      float64 `json:"temperature"`       // in Celsius
}

// Monitor handles hardware monitoring
type Monitor struct {
	config *config.Config
}

// NewMonitor creates a new hardware monitor
func NewMonitor(cfg *config.Config) *Monitor {
	return &Monitor{
		config: cfg,
	}
}

// Collect gathers all enabled hardware statistics
func (m *Monitor) Collect() (*Stats, error) {
	stats := &Stats{
		Timestamp: time.Now(),
		MateID:    m.config.Mate.ID,
	}

	// Always collect system info
	if sysStats, err := m.collectSystem(); err == nil {
		stats.System = sysStats
	}

	// CPU
	if m.config.Monitoring.Enabled.CPU {
		if cpuStats, err := m.collectCPU(); err == nil {
			stats.CPU = cpuStats
		}
	}

	// Memory
	if m.config.Monitoring.Enabled.Memory {
		if memStats, err := m.collectMemory(); err == nil {
			stats.Memory = memStats
		}
	}

	// Disk
	if m.config.Monitoring.Enabled.Disk {
		if diskStats, err := m.collectDisk(); err == nil {
			stats.Disk = diskStats
		}
	}

	// Temperature
	if m.config.Monitoring.Enabled.Temperature {
		if tempStats, err := m.collectTemperature(); err == nil {
			stats.Temperature = tempStats
		}
	}

	// Network
	if m.config.Monitoring.Enabled.Network {
		if netStats, err := m.collectNetwork(); err == nil {
			stats.Network = netStats
		}
	}

	// GPU
	if m.config.Monitoring.Enabled.GPU {
		if gpuStats, err := m.collectGPU(); err == nil {
			stats.GPU = gpuStats
		}
	}

	return stats, nil
}

// collectCPU collects CPU statistics
func (m *Monitor) collectCPU() (*CPUStats, error) {
	// CPU usage
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// Per-core usage if enabled
	var perCore []float64
	if m.config.Hardware.CPU.CollectPerCore {
		perCore, err = cpu.Percent(time.Second, true)
		if err != nil {
			perCore = nil
		}
	}

	// CPU info
	info, err := cpu.Info()
	if err != nil || len(info) == 0 {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	// Count logical cores
	cores, _ := cpu.Counts(true)

	return &CPUStats{
		UsagePercent: percentages[0],
		PerCore:      perCore,
		Cores:        cores,
		Model:        info[0].ModelName,
		MHz:          info[0].Mhz,
	}, nil
}

// collectMemory collects memory statistics
func (m *Monitor) collectMemory() (*MemoryStats, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	stats := &MemoryStats{
		Total:       vmem.Total,
		Available:   vmem.Available,
		Used:        vmem.Used,
		UsedPercent: vmem.UsedPercent,
	}

	// Swap memory if enabled
	if m.config.Hardware.Memory.IncludeSwap {
		swap, err := mem.SwapMemory()
		if err == nil {
			stats.SwapTotal = swap.Total
			stats.SwapUsed = swap.Used
			stats.SwapPercent = swap.UsedPercent
		}
	}

	return stats, nil
}

// collectDisk collects disk statistics
func (m *Monitor) collectDisk() ([]DiskStats, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	var stats []DiskStats

	// Filter mount points if specified
	mountPoints := m.config.Hardware.Disk.MountPoints
	shouldCollect := func(mountPoint string) bool {
		if len(mountPoints) == 0 {
			return true
		}
		for _, mp := range mountPoints {
			if mp == mountPoint {
				return true
			}
		}
		return false
	}

	for _, partition := range partitions {
		if !shouldCollect(partition.Mountpoint) {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		stats = append(stats, DiskStats{
			MountPoint:  partition.Mountpoint,
			Device:      partition.Device,
			FSType:      partition.Fstype,
			Total:       usage.Total,
			Free:        usage.Free,
			Used:        usage.Used,
			UsedPercent: usage.UsedPercent,
		})
	}

	return stats, nil
}

// collectTemperature collects temperature statistics
func (m *Monitor) collectTemperature() (*TemperatureStats, error) {
	temps, err := host.SensorsTemperatures()
	if err != nil {
		return nil, fmt.Errorf("failed to get temperature: %w", err)
	}

	stats := &TemperatureStats{
		Sensors: make([]SensorTemp, 0),
	}

	// Filter sensors if specified
	sensors := m.config.Hardware.Temperature.Sensors
	shouldCollect := func(name string) bool {
		if len(sensors) == 0 {
			return true
		}
		for _, s := range sensors {
			if s == name {
				return true
			}
		}
		return false
	}

	for _, temp := range temps {
		if !shouldCollect(temp.SensorKey) {
			continue
		}

		stats.Sensors = append(stats.Sensors, SensorTemp{
			Name:        temp.SensorKey,
			Temperature: temp.Temperature,
			High:        temp.High,
			Critical:    temp.Critical,
		})
	}

	return stats, nil
}

// collectNetwork collects network statistics
func (m *Monitor) collectNetwork() ([]NetworkStats, error) {
	counters, err := net.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get network stats: %w", err)
	}

	var stats []NetworkStats

	// Filter interfaces if specified
	interfaces := m.config.Hardware.Network.Interfaces
	shouldCollect := func(name string) bool {
		if len(interfaces) == 0 {
			return true
		}
		for _, iface := range interfaces {
			if iface == name {
				return true
			}
		}
		return false
	}

	for _, counter := range counters {
		if !shouldCollect(counter.Name) {
			continue
		}

		stats = append(stats, NetworkStats{
			Interface:   counter.Name,
			BytesSent:   counter.BytesSent,
			BytesRecv:   counter.BytesRecv,
			PacketsSent: counter.PacketsSent,
			PacketsRecv: counter.PacketsRecv,
			Errin:       counter.Errin,
			Errout:      counter.Errout,
		})
	}

	return stats, nil
}

// collectSystem collects system information
func (m *Monitor) collectSystem() (*SystemStats, error) {
	info, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get system info: %w", err)
	}

	return &SystemStats{
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		Uptime:          info.Uptime,
	}, nil
}

// collectGPU collects GPU statistics using nvidia-smi
func (m *Monitor) collectGPU() ([]GPUStats, error) {
	// Check if nvidia-smi is available
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,gpu_name,utilization.gpu,memory.total,memory.used,memory.free,temperature.gpu", "--format=csv,noheader,nounits")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run nvidia-smi: %w", err)
	}

	var gpuStats []GPUStats

	// Parse output line by line (one line per GPU)
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse CSV: index, name, utilization, mem_total, mem_used, mem_free, temperature
		fields := strings.Split(line, ", ")
		if len(fields) < 7 {
			continue
		}

		index, _ := strconv.Atoi(strings.TrimSpace(fields[0]))
		name := strings.TrimSpace(fields[1])
		utilization, _ := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
		memTotal, _ := strconv.ParseUint(strings.TrimSpace(fields[3]), 10, 64)
		memUsed, _ := strconv.ParseUint(strings.TrimSpace(fields[4]), 10, 64)
		memFree, _ := strconv.ParseUint(strings.TrimSpace(fields[5]), 10, 64)
		temperature, _ := strconv.ParseFloat(strings.TrimSpace(fields[6]), 64)

		// Calculate memory usage percentage
		var memUsedPercent float64
		if memTotal > 0 {
			memUsedPercent = float64(memUsed) / float64(memTotal) * 100.0
		}

		gpuStats = append(gpuStats, GPUStats{
			Index:             index,
			Name:              name,
			UtilizationGPU:    utilization,
			MemoryTotal:       memTotal,
			MemoryUsed:        memUsed,
			MemoryFree:        memFree,
			MemoryUsedPercent: memUsedPercent,
			Temperature:       temperature,
		})
	}

	if len(gpuStats) == 0 {
		return nil, fmt.Errorf("no GPUs found")
	}

	return gpuStats, nil
}
