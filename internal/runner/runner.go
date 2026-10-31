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

var (
	cachedHostname string
	cachedVersion  string
	localTimestamp string
)

func Run(cfg config.Config) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Evalua si agregar la hora local al resultado
	if cfg.General.LocalTime.Enable {
		localTimestamp = time.Now().Format(time.RFC3339)
	}

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

func runCollector[T any](ctx context.Context, wg *sync.WaitGroup, name string, interval time.Duration, readFn func() (T, error)) {

	defer wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Función interna para ejecutar la lectura y procesarla
	executeRead := func() {
		stats, err := readFn()
		if err != nil {
			// Captura y serializa un error de lectura.
			payload, serializeErr := handleCollectorError(name, interval, err)
			if serializeErr != nil {
				fmt.Printf("[%s] error serializing error payload: %v\n", name, serializeErr)
				return
			}

			// PROVISIONAL - TODO: Issue #3, por ahora se imprime como debug.
			fmt.Println(string(payload))
			return
		}

		// Flujo de serializacion normal.
		payload, err := serializer.FormatPayload(cachedHostname, cachedVersion, name, interval, stats, localTimestamp)
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

// Convierte los errores en un JSON para que queden en registro tambien.
func handleCollectorError(name string, interval time.Duration, err error) ([]byte, error) {

	// En este caso es mejor desactivar el componente en configuración.
	if errors.Is(err, collector.ErrCapabilityUnavailable) {
		fmt.Printf("[%s] on %s: functionality unavailable in this hardware (interval: %s)\n", name, cachedHostname, interval)
	}

	// Crear JSON de error con serializer.
	return serializer.FormatPayloadError(cachedHostname, cachedVersion, name, interval, err, localTimestamp)
}

// Consigue información extra una sola vez.
func SetMetadata(hostname, version string) {
	cachedHostname = hostname
	cachedVersion = version
}
