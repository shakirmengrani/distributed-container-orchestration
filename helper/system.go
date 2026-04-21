package helper

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type ResourceUsage struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}

// SystemResources holds CPU, memory, and disk
type SystemResources struct {
	CPU  ResourceUsage `json:"cpu"`
	Mem  ResourceUsage `json:"mem"`
	Disk ResourceUsage `json:"disk"`
}

func GetSystemResources() (*SystemResources, error) {
	// Memory
	vm, _ := mem.VirtualMemory()
	// Disk (root "/")
	ds, _ := disk.Usage("/")
	// CPU
	cpuCounts, _ := cpu.Counts(true)

	memUsage := ResourceUsage{
		Total: vm.Total,
		Used:  vm.Used,
		Free:  vm.Available,
	}

	diskUsage := ResourceUsage{
		Total: ds.Total,
		Used:  ds.Used,
		Free:  ds.Free,
	}

	// For used CPU, you could calculate utilization percentage
	cpuUsage := ResourceUsage{
		Total: uint64(cpuCounts),
		Used:  0, // Optional: calculate CPU usage %
		Free:  uint64(cpuCounts),
	}

	return &SystemResources{
		CPU:  cpuUsage,
		Mem:  memUsage,
		Disk: diskUsage,
	}, nil
}

func ComputeCapacity(sysRes SystemResources) bool {
	// Define thresholds (example: 20% free minimum)
	const minFreeMemPercent = 20.0
	const minFreeDiskPercent = 20.0
	const minFreeCPUPercent = 20.0

	// Compute capacity flags
	memFreePercent := float64(sysRes.Mem.Free) / float64(sysRes.Mem.Total) * 100
	diskFreePercent := float64(sysRes.Disk.Free) / float64(sysRes.Disk.Total) * 100
	cpuFreePercent := float64(sysRes.CPU.Free) / float64(sysRes.CPU.Total) * 100

	isCapacity := memFreePercent >= minFreeMemPercent &&
		diskFreePercent >= minFreeDiskPercent &&
		cpuFreePercent >= minFreeCPUPercent
	return isCapacity
}

func MBToB(mb ...float64) int64 {
	val := 1024.0
	if len(mb) > 0 {
		val = mb[0]
	}
	return int64(val * 1024 * 1024)
}
