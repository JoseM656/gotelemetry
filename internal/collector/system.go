package collector

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// ---------- CPU ----------

// Devuelve los datos del CPU en crudo.
type CPUStats struct {
	User, Nice, System, Idle, Iowait, IRQ, SoftIRQ, Steal uint64
}

func ReadCPU() (CPUStats, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return CPUStats{}, fmt.Errorf("read cpu: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return CPUStats{}, fmt.Errorf("read cpu: /proc/stat empty")
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) < 9 || fields[0] != "cpu" {
		return CPUStats{}, fmt.Errorf("read cpu: unexpected format in /proc/stat")
	}

	vals := make([]uint64, 8)
	for i := 0; i < 8; i++ {
		v, err := strconv.ParseUint(fields[i+1], 10, 64)
		if err != nil {
			return CPUStats{}, fmt.Errorf("read cpu: parsing field %d: %w", i, err)
		}
		vals[i] = v
	}

	return CPUStats{
		User:    vals[0],
		Nice:    vals[1],
		System:  vals[2],
		Idle:    vals[3],
		Iowait:  vals[4],
		IRQ:     vals[5],
		SoftIRQ: vals[6],
		Steal:   vals[7],
	}, nil
}

// ---------- RAM ----------

type RAMStats struct {
	TotalKB, FreeKB, AvailableKB, BuffersKB, CachedKB uint64
}

func ReadRAM() (RAMStats, error) {
	fields, err := parseMeminfo()
	if err != nil {
		return RAMStats{}, fmt.Errorf("read ram: %w", err)
	}

	return RAMStats{
		TotalKB:     fields["MemTotal"],
		FreeKB:      fields["MemFree"],
		AvailableKB: fields["MemAvailable"],
		BuffersKB:   fields["Buffers"],
		CachedKB:    fields["Cached"],
	}, nil
}

// ---------- Swap ----------

type SwapStats struct {
	TotalKB, FreeKB uint64
}

func ReadSwap() (SwapStats, error) {
	fields, err := parseMeminfo()
	if err != nil {
		return SwapStats{}, fmt.Errorf("read swap: %w", err)
	}

	return SwapStats{
		TotalKB: fields["SwapTotal"],
		FreeKB:  fields["SwapFree"],
	}, nil
}

// parseMeminfo lee /proc/meminfo una sola vez y devuelve todos los campos
// como mapa. RAM y Swap comparten el mismo archivo fuente.
func parseMeminfo() (map[string]uint64, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]uint64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		val, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue // ignora líneas no numéricas.
		}
		result[key] = val
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// ---------- Almacenamiento ----------

type StorageStats struct {
	MountPoint string
	TotalBytes uint64
	FreeBytes  uint64
	UsedBytes  uint64
}

// pseudoFilesystems son tipos que no representan almacenamiento físico real
// (procfs, sysfs, tmpfs en RAM, cgroups, etc.) y que /proc/mounts siempre
// lista junto a los discos reales. Filtrarlos acá evita reportar
// "almacenamiento" que en realidad es memoria o metadata del kernel.
var pseudoFilesystems = map[string]bool{
	"proc": true, "sysfs": true, "tmpfs": true, "devtmpfs": true,
	"devpts": true, "cgroup": true, "cgroup2": true, "overlay": true,
	"squashfs": true, "debugfs": true, "tracefs": true, "mqueue": true,
	"pstore": true, "bpf": true, "autofs": true, "rpc_pipefs": true,
}

func ReadStorage() ([]StorageStats, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("read storage: %w", err)
	}
	defer f.Close()

	var results []StorageStats
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		mountPoint, fsType := fields[1], fields[2]

		if pseudoFilesystems[fsType] {
			continue
		}

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mountPoint, &stat); err != nil {
			continue // punto de montaje no accesible (permisos, montaje roto); no aborta el resto
		}

		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bfree * uint64(stat.Bsize)

		results = append(results, StorageStats{
			MountPoint: mountPoint,
			TotalBytes: total,
			FreeBytes:  free,
			UsedBytes:  total - free,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read storage: %w", err)
	}

	return results, nil
}

// ---------- GPU ----------

// Como hay varios fabricantes primero se guarda con un sync.Once para no consultar en cada lectura.

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
