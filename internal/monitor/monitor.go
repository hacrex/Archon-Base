// Package monitor collects local host resources for the student self-hosting profile.
package monitor

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	MinimumMemoryBytes = uint64(8 * 1024 * 1024 * 1024)
	MinimumCPUCores    = 2
	MinimumDiskBytes   = uint64(20 * 1024 * 1024 * 1024)
	MinimumGPUMemoryMB = 1024
)

type Snapshot struct {
	CollectedAt time.Time   `json:"collectedAt"`
	Profile     string      `json:"profile"`
	CPU         CPUStatus   `json:"cpu"`
	Memory      MemoryStats `json:"memory"`
	Disk        DiskStats   `json:"disk"`
	GPU         *GPUStats   `json:"gpu,omitempty"`
	Compliance  Compliance  `json:"compliance"`
}

type CPUStatus struct {
	Cores     int     `json:"cores"`
	Load1m    float64 `json:"load1m"`
	LoadRatio float64 `json:"loadRatio"`
}

type MemoryStats struct {
	TotalBytes     uint64  `json:"totalBytes"`
	AvailableBytes uint64  `json:"availableBytes"`
	UsedBytes      uint64  `json:"usedBytes"`
	UsedRatio      float64 `json:"usedRatio"`
}

type DiskStats struct {
	Path           string  `json:"path"`
	TotalBytes     uint64  `json:"totalBytes"`
	AvailableBytes uint64  `json:"availableBytes"`
	UsedRatio      float64 `json:"usedRatio"`
}

type GPUStats struct {
	Name          string  `json:"name,omitempty"`
	TotalMemoryMB uint64  `json:"totalMemoryMB"`
	UsedMemoryMB  uint64  `json:"usedMemoryMB"`
	Utilization   float64 `json:"utilizationPercent"`
}

type Compliance struct {
	Ready       bool     `json:"ready"`
	GPUOptional bool     `json:"gpuOptional"`
	Warnings    []string `json:"warnings,omitempty"`
}

type Collector struct {
	DiskPath string
	GPUQuery func(context.Context) (*GPUStats, error)
}

func NewCollector(diskPath string) Collector {
	if strings.TrimSpace(diskPath) == "" {
		diskPath = "/var/lib/rancher/k3s"
	}
	return Collector{DiskPath: diskPath, GPUQuery: queryNVIDIA}
}

func (c Collector) Collect(ctx context.Context) (Snapshot, error) {
	memory, err := readMemory()
	if err != nil {
		return Snapshot{}, fmt.Errorf("read memory: %w", err)
	}
	disk, err := readDisk(c.DiskPath)
	if err != nil {
		return Snapshot{}, fmt.Errorf("read disk: %w", err)
	}
	load, err := readLoad1m()
	if err != nil {
		return Snapshot{}, fmt.Errorf("read load: %w", err)
	}
	cores := runtime.NumCPU()
	snapshot := Snapshot{
		CollectedAt: time.Now().UTC(), Profile: "enthusiast-k3s",
		CPU:    CPUStatus{Cores: cores, Load1m: load, LoadRatio: load / float64(max(cores, 1))},
		Memory: memory, Disk: disk,
		Compliance: Compliance{GPUOptional: true},
	}
	if c.GPUQuery != nil {
		gpuCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		gpu, gpuErr := c.GPUQuery(gpuCtx)
		cancel()
		if gpuErr == nil && gpu != nil {
			snapshot.GPU = gpu
		}
	}
	snapshot.Compliance = Evaluate(snapshot)
	return snapshot, nil
}

func Evaluate(snapshot Snapshot) Compliance {
	compliance := Compliance{GPUOptional: true, Ready: true}
	if snapshot.CPU.Cores < MinimumCPUCores {
		compliance.Ready = false
		compliance.Warnings = append(compliance.Warnings, fmt.Sprintf("at least %d CPU cores are recommended", MinimumCPUCores))
	}
	if snapshot.Memory.TotalBytes < MinimumMemoryBytes {
		compliance.Ready = false
		compliance.Warnings = append(compliance.Warnings, "at least 8 GB RAM is recommended")
	}
	if snapshot.Disk.TotalBytes < MinimumDiskBytes {
		compliance.Ready = false
		compliance.Warnings = append(compliance.Warnings, "at least 20 GB free disk capacity is recommended")
	}
	if snapshot.GPU == nil {
		compliance.Warnings = append(compliance.Warnings, "no NVIDIA GPU detected; CPU-only K3s practice remains supported")
	} else if snapshot.GPU.TotalMemoryMB < MinimumGPUMemoryMB {
		compliance.Warnings = append(compliance.Warnings, "1 GB graphics memory is recommended for GPU experiments")
	}
	return compliance
}

func readMemory() (MemoryStats, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemoryStats{}, err
	}
	defer file.Close()
	values := map[string]uint64{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, parseErr := strconv.ParseUint(fields[1], 10, 64)
		if parseErr == nil {
			// Linux reports meminfo values in KiB.
			values[strings.TrimSuffix(fields[0], ":")] = value * 1024
		}
	}
	if err := scanner.Err(); err != nil {
		return MemoryStats{}, err
	}
	total := values["MemTotal"]
	available := values["MemAvailable"]
	if total == 0 {
		return MemoryStats{}, errors.New("MemTotal is missing")
	}
	used := total - min(total, available)
	return MemoryStats{TotalBytes: total, AvailableBytes: available, UsedBytes: used, UsedRatio: float64(used) / float64(total)}, nil
}

func readDisk(path string) (DiskStats, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return DiskStats{}, err
	}
	total := stat.Blocks * uint64(stat.Bsize)
	available := stat.Bavail * uint64(stat.Bsize)
	used := total - min(total, available)
	ratio := float64(0)
	if total > 0 {
		ratio = float64(used) / float64(total)
	}
	return DiskStats{Path: path, TotalBytes: total, AvailableBytes: available, UsedRatio: ratio}, nil
}

func readLoad1m() (float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0, errors.New("load average is empty")
	}
	return strconv.ParseFloat(fields[0], 64)
}

func queryNVIDIA(ctx context.Context) (*GPUStats, error) {
	command := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=name,memory.total,memory.used,utilization.gpu", "--format=csv,noheader,nounits")
	output, err := command.Output()
	if err != nil {
		return nil, err
	}
	line := strings.TrimSpace(strings.SplitN(string(output), "\n", 2)[0])
	fields := strings.Split(line, ",")
	if len(fields) < 4 {
		return nil, errors.New("unexpected nvidia-smi output")
	}
	parseMB := func(value string) (uint64, error) { return strconv.ParseUint(strings.TrimSpace(value), 10, 64) }
	total, err := parseMB(fields[1])
	if err != nil {
		return nil, err
	}
	used, err := parseMB(fields[2])
	if err != nil {
		return nil, err
	}
	utilization, err := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64)
	if err != nil {
		return nil, err
	}
	return &GPUStats{Name: strings.TrimSpace(fields[0]), TotalMemoryMB: total, UsedMemoryMB: used, Utilization: utilization}, nil
}

func WriteJSON(path string, snapshot Snapshot) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, append(data, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
