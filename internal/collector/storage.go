package collector

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
)

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
