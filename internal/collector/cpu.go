package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

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
