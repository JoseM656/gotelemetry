package runner

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
	"github.com/JoseM656/gotelemetry/internal/serializer"
)

func Run(cfg config.Config) {

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

	// Función interna para ejecutar la lectura y procesarla
	executeRead := func() {
		stats, err := readFn()
		if err != nil {
			// El supervisor maneja el error sin romper la app
			handleCollectorError(name, interval, err)
			return
		}

		// Si todo salió bien, recién aquí interviene el serializador
		payload, err := serializer.FormatPayload(name, interval, stats)
		if err != nil {
			fmt.Printf("[%s] error serializing payload: %v\n", name, err)
			return
		}

		// PROVISIONAL - TODO: Issue #3, por ahora se imprime como debug.
		fmt.Println(string(payload))
	}

	// Primera ejecución inmediata
	executeRead()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			executeRead()
		}
	}
}

// PROVISIONAL - TODO: Issue #28, estos errores son parte de la lectura tambien.
// Función auxiliar para gestionar los errores de forma limpia
func handleCollectorError(name string, interval time.Duration, err error) {
	if errors.Is(err, collector.ErrCapabilityUnavailable) {
		fmt.Printf("[%s] funcionality unavailable in this hardware (interval: %s)\n", name, interval)
	} else {
		fmt.Printf("[%s] error of lecture: %v (interval: %s)\n", name, err, interval)
	}
}
