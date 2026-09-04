package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/JoseM656/gotelemetry/internal/collector"
	"github.com/JoseM656/gotelemetry/internal/config"
)

var BuildVersion = ""

func loadPath(configPath string) config.Config {

	cfg, created, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("config: error loading %q: %v\n", configPath, err)
		os.Exit(1)
	}

	if created {
		fmt.Printf("config: %q not found. Regenerating...\n", configPath)
	} else {
		fmt.Printf("config loaded. %q\n", configPath)
	}

	return cfg

}

func main() {

	// === Procesar argumentos desde cli.go ===
	args := ParseFlags()

	if args.ShowVersion {
		fmt.Printf("go telemetry version %v\n", BuildVersion)
		os.Exit(0)
	}

	// Carga en config.go - PROVISIONAL TODO: Cambiar la ubicacion.
	cfg := loadPath("/etc/gotelemetry/config.yml")

	// ===========================================

	// Procesar señales del sistema.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Lanza cada trabajador segun su propia configuracion.
	if cfg.Collectors.CPU.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "cpu", cfg.Collectors.CPU.Interval.Duration, collector.ReadCPU)
	}

	if cfg.Collectors.RAM.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "ram", cfg.Collectors.RAM.Interval.Duration, collector.ReadRAM)
	}

	if cfg.Collectors.Swap.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "swap", cfg.Collectors.Swap.Interval.Duration, collector.ReadSwap)
	}

	if cfg.Collectors.Storage.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "storage", cfg.Collectors.Storage.Interval.Duration, collector.ReadStorage)
	}

	if cfg.Collectors.GPU.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "gpu", cfg.Collectors.GPU.Interval.Duration, collector.ReadGPU)
	}

	fmt.Println("Running on background. Use Ctrl+C for exit.")

	// Bloquea main hasta recibir la señal SIGINT/SIGTERM
	sig := <-sigChan
	fmt.Printf("\nSignal %v recibed. Exiting...", sig)

	cancel()

	// Espera a que todas las goroutines de recolectores limpien y salgan
	wg.Wait()
}

// runCollector ejecuta la función de lectura respetando el intervalo específico de su ticker
func runCollector[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	name string,
	interval time.Duration,
	readFn func() (T, error),
) {
	defer wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Ejecutar una primera lectura inmediata al arrancar
	stats, err := readFn()
	printResult(name, interval, stats, err)

	for {
		select {
		case <-ctx.Done():
			// Se canceló el contexto
			return
		case <-ticker.C:
			// Se cumplió el intervalo del componente
			stats, err := readFn()
			printResult(name, interval, stats, err)
		}
	}
}

// PROVISIONAL - TODO: Funcion de logger.
func printResult(name string, interval time.Duration, stats any, err error) {
	if err != nil {
		if errors.Is(err, collector.ErrCapabilityUnavailable) {
			fmt.Printf("[%s] no disponible en este hardware (intervalo: %s)\n", name, interval)
		} else {
			fmt.Printf("[%s] error: %v\n", name, err)
		}
		return
	}

	fmt.Printf("[%s] (intervalo: %s) => %+v\n", name, interval, stats)
}
