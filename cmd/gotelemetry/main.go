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
var ConfigPath = ""

// Resuelve encontrar el path de la configuración
func fallbackPath() string {
	// 1. Inyectado desde el Makefile
	if ConfigPath != "" {
		return ConfigPath
	}

	// 2. Por env
	if envPath := os.Getenv("GOTELEMETRY_CONFIG_PATH"); envPath != "" {
		return envPath
	}

	// 3. Default
	fmt.Println("fallback: There is not a selected config path or a env, using default /etc/gotelemetry/config.yml,\nthis will require your permission if the file not exist. :D")
	return "/etc/gotelemetry/config.yml"
}

// Carga o crea la configuración segun el path
func loadPath(path string) config.Config {

	cfg, created, err := config.Load(path)
	if err != nil {
		fmt.Printf("error loading %q: %v\n", path, err)
		os.Exit(1)
	}

	if created {
		fmt.Printf("config: %q not found. Regenerating...\n", path)

	} else {
		fmt.Printf("config loaded. %q\n", path)
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

	if args.RegenConfig {
		// Gestiona el fallback al regenerar
		ConfigPath = fallbackPath()
		err := config.SaveDefault(ConfigPath)
		if err != nil {
			fmt.Printf("error regenerating in %q: %v\n", ConfigPath, err)
			os.Exit(1)
		}

		fmt.Printf("config regenerated in %v\n", ConfigPath)
		os.Exit(0)

	}

	// == Fin de procesar argumentos desde cli.go ===

	// Cargar path
	ConfigPath = fallbackPath()
	cfg := loadPath(ConfigPath)

	// Procesar señales del sistema.
	ctx, cancel := context.WithCancel(context.Background())
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
	fmt.Printf("Signal %v recibed. Exiting...\n", sig)

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
