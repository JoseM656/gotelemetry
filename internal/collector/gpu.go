package collector

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type GPUStats struct {
	UsagePercent uint64
}

type gpuVendor int

const (
	gpuVendorUnknown gpuVendor = iota
	gpuVendorNone
	gpuVendorAMD
	gpuVendorIntel
	gpuVendorNvidia
)

var (
	detectedGPUVendor gpuVendor
	gpuDetectOnce     sync.Once
)

const gpuDRMDevicePath = "/sys/class/drm/card0/device"

// detectGPUVendor lee el vendor ID PCI expuesto por el kernel en sysfs para
// decidir qué estrategia de lectura usar. IDs estándar: AMD/ATI = 0x1002,
// Intel = 0x8086, Nvidia = 0x10de.
func detectGPUVendor() gpuVendor {
	data, err := os.ReadFile(gpuDRMDevicePath + "/vendor")
	if err != nil {
		return gpuVendorNone
	}

	switch strings.TrimSpace(string(data)) {
	case "0x1002":
		return gpuVendorAMD
	case "0x8086":
		return gpuVendorIntel
	case "0x10de":
		if _, err := exec.LookPath("nvidia-smi"); err == nil {
			return gpuVendorNvidia
		}
		return gpuVendorNone
	default:
		return gpuVendorNone
	}
}

func ReadGPU() (GPUStats, error) {
	gpuDetectOnce.Do(func() {
		detectedGPUVendor = detectGPUVendor()
	})

	switch detectedGPUVendor {
	case gpuVendorAMD:
		return readAMDGPU()
	case gpuVendorIntel:
		return readIntelGPU()
	case gpuVendorNvidia:
		return readNvidiaGPU()
	default:
		return GPUStats{}, ErrCapabilityUnavailable
	}
}

func readAMDGPU() (GPUStats, error) {
	data, err := os.ReadFile(gpuDRMDevicePath + "/gpu_busy_percent")
	if err != nil {
		return GPUStats{}, ErrCapabilityUnavailable
	}

	busy, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return GPUStats{}, fmt.Errorf("read gpu (amd): %w", err)
	}

	return GPUStats{UsagePercent: busy}, nil
}

// TODO: Leer la GPU de intel es una tarea indepenidente por si misma.
func readIntelGPU() (GPUStats, error) {
	return GPUStats{}, ErrCapabilityUnavailable
}

func readNvidiaGPU() (GPUStats, error) {
	out, err := exec.Command(
		"nvidia-smi",
		"--query-gpu=utilization.gpu",
		"--format=csv,noheader,nounits",
	).Output()
	if err != nil {
		return GPUStats{}, fmt.Errorf("read gpu (nvidia): %w", err)
	}

	busy, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return GPUStats{}, fmt.Errorf("read gpu (nvidia): %w", err)
	}

	return GPUStats{UsagePercent: busy}, nil
}
