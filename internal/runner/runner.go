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
)

func Run(cfg config.Config) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Variables opcionales segun la configuración leidas al arrancar.
	EnableLocalTimestamp := cfg.Output.LocalTime.Enabled
	EnablePrettyInterval := cfg.Output.PrettyOutput.Enabled

	// Lanza cada trabajador segun su propia configuracion.
	if cfg.Collectors.CPU.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "cpu", cfg.Collectors.CPU.Interval.Duration, EnableLocalTimestamp, EnablePrettyInterval, collector.ReadCPU)
	}

	if cfg.Collectors.RAM.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "ram", cfg.Collectors.RAM.Interval.Duration, EnableLocalTimestamp, EnablePrettyInterval, collector.ReadRAM)
	}

	if cfg.Collectors.Swap.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "swap", cfg.Collectors.Swap.Interval.Duration, EnableLocalTimestamp, EnablePrettyInterval, collector.ReadSwap)
	}

	if cfg.Collectors.Storage.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "storage", cfg.Collectors.Storage.Interval.Duration, EnableLocalTimestamp, EnablePrettyInterval, collector.ReadStorage)
	}

	if cfg.Collectors.GPU.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "gpu", cfg.Collectors.GPU.Interval.Duration, EnableLocalTimestamp, EnablePrettyInterval, collector.ReadGPU)
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
	EnableLocalTimestamp bool, EnablePrettyInterval bool,
	readFn func() (T, error),
) {

	defer wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	executeRead := func() {
		// 1. Evaluar las variables opcionales en cada tick.
		var localTs string
		if EnableLocalTimestamp {
			localTs = time.Now().Format(time.RFC3339)
		}

		var prettyInt string
		if EnablePrettyInterval {
			prettyInt = interval.String() // ej: "5s"
		}

		// Agrupar las opciones para el serializador.
		opts := serializer.PayloadOptions{
			LocalTimestamp: localTs,
			PrettyInterval: prettyInt,
		}

		stats, err := readFn()
		if err != nil {
			// 2. Si hay error, se pasan los opcionales al error tambien.
			payload, serializeErr := handleCollectorError(name, interval, err, opts)
			if serializeErr != nil {
				fmt.Printf("[%s] error serializing error payload: %v\n", name, serializeErr)
				return
			}
			// PROVISIONAL - TODO: Issue #3, por ahora se imprime como debug.
			fmt.Println(string(payload))
			return
		}

		// 3. Flujo de serialización normal
		payload, err := serializer.FormatPayload(cachedHostname, cachedVersion, name, interval, opts, stats)
		if err != nil {
			fmt.Printf("[%s] error serializing payload: %v\n", name, err)
			return
		}

		// PROVISIONAL - TODO: Issue #3, por ahora se imprime como debug.
		fmt.Println(string(payload))
	}

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
func handleCollectorError(name string, interval time.Duration, err error, opts serializer.PayloadOptions) ([]byte, error) {
	if errors.Is(err, collector.ErrCapabilityUnavailable) {
		fmt.Printf("[%s] on %s: functionality unavailable in this hardware (interval: %s)\n", name, cachedHostname, interval)
	}

	return serializer.FormatPayloadError(cachedHostname, cachedVersion, name, interval, opts, err)
}

// Consigue información extra una sola vez.
func SetMetadata(hostname, version string) {
	cachedHostname = hostname
	cachedVersion = version
}
