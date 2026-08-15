package main

import (
	"errors"
	"fmt"

	// TODO: reemplazar "telemetrygo" por el module path real de tu go.mod
	// (el que definiste al correr `go mod init`).
	"github.com/JoseM656/TelemetryGo/internal/collector"
	"github.com/JoseM656/TelemetryGo/internal/config"
)

// PROVISIONAL UNTIL THERE IS AN ACTUAL OUTPUT
func main() {

	// RUTA PROVISIONAL ANTES DEL MAKEFILE
	const configPath = "config.yml"

	cfg, err := config.Load(configPath)
	if err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			fmt.Printf("config: %q not found, using standar configuration\n", configPath)
		} else {
			fmt.Printf("config: error loading %q: %v\n", configPath, err)
		}
	}

	fmt.Printf("config loaded: %+v\n\n", cfg)

	if cfg.Collectors.CPU.Enabled {
		stats, err := collector.ReadCPU()
		printResult("cpu", cfg.Collectors.CPU.Interval, stats, err)
	}

	if cfg.Collectors.RAM.Enabled {
		stats, err := collector.ReadRAM()
		printResult("ram", cfg.Collectors.RAM.Interval, stats, err)
	}

	if cfg.Collectors.Swap.Enabled {
		stats, err := collector.ReadSwap()
		printResult("swap", cfg.Collectors.Swap.Interval, stats, err)
	}

	if cfg.Collectors.Storage.Enabled {
		stats, err := collector.ReadStorage()
		printResult("storage", cfg.Collectors.Storage.Interval, stats, err)
	}

	if cfg.Collectors.GPU.Enabled {
		stats, err := collector.ReadGPU()
		printResult("gpu", cfg.Collectors.GPU.Interval, stats, err)
	}
}

// PROVISIONAL PARA PROBAR
func printResult(name string, interval config.Duration, stats any, err error) {
	if err != nil {
		if errors.Is(err, collector.ErrCapabilityUnavailable) {
			fmt.Printf("[%s] no disponible en este hardware (intervalo configurado: %s)\n", name, interval.Duration)
		} else {
			fmt.Printf("[%s] error: %v\n", name, err)
		}
		return
	}

	fmt.Printf("[%s] (intervalo configurado: %s) => %+v\n", name, interval.Duration, stats)
}
