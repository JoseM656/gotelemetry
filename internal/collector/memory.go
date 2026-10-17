package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

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
