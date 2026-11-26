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

	// Canal compartido runner <-> dispacher
	metricsChan := make(chan []byte, 100)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Iniciar el dispacher y su lectura del canal como una gorutina.
	// Tiene grupo propio para poder ceerarlo de forma propia.
	var dispWg sync.WaitGroup
	dispWg.Go(func() {
		StartDispatcher(ctx, cfg, metricsChan, cancel)
	})

	// Variables opcionales segun la configuración leidas al arrancar.
	EnableLocalTimestamp := cfg.General.LocalTime.Enabled
	EnablePrettyInterval := cfg.General.PrettyOutput.Enabled

	// Lanza cada trabajador segun su propia configuracion.
	if cfg.Collectors.CPU.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "cpu", cfg.Collectors.CPU.Interval.Duration, metricsChan, EnableLocalTimestamp, EnablePrettyInterval, collector.ReadCPU)
	}

	if cfg.Collectors.RAM.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "ram", cfg.Collectors.RAM.Interval.Duration, metricsChan, EnableLocalTimestamp, EnablePrettyInterval, collector.ReadRAM)
	}

	if cfg.Collectors.Swap.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "swap", cfg.Collectors.Swap.Interval.Duration, metricsChan, EnableLocalTimestamp, EnablePrettyInterval, collector.ReadSwap)
	}

	if cfg.Collectors.Storage.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "storage", cfg.Collectors.Storage.Interval.Duration, metricsChan, EnableLocalTimestamp, EnablePrettyInterval, collector.ReadStorage)
	}

	if cfg.Collectors.GPU.Enabled {
		wg.Add(1)
		go runCollector(ctx, &wg, "gpu", cfg.Collectors.GPU.Interval.Duration, metricsChan, EnableLocalTimestamp, EnablePrettyInterval, collector.ReadGPU)
	}

	fmt.Fprintln(os.Stderr, "Running on background. Use Ctrl+C or SIGTERM for exit.")

	// Cerrar el contexto de runner si hay una señal o un error interno.
	select {
	case sig := <-sigChan:
		fmt.Fprintf(os.Stderr, "\nSignal %v received. Exiting...\n", sig)
	case <-ctx.Done():
		fmt.Fprintln(os.Stderr, "Context cancelled internally. Exiting...")
	}

	fmt.Fprintln(os.Stderr, "collectors shutting down...")

	// === Secuencia de apagado ===

	// Collectores paran
	cancel()

	// Esperar que se terminen los escitores
	wg.Wait()

	// Cerrar el canal principal
	close(metricsChan)

	// Esperar a que dispacher drene el buffer y cierre
	dispWg.Wait()
}

func runCollector[T any](
	ctx context.Context,
	wg *sync.WaitGroup,
	name string,
	interval time.Duration,
	metricsChan chan<- []byte,
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
				fmt.Fprintf(os.Stderr, "[%s] error serializing error payload: %v\n", name, serializeErr)
				return
			}
			// Debug.
			// fmt.Println(string(payload))

			// Enviar al canal que dispacher lee.
			select {
			case metricsChan <- payload:
			default:
				fmt.Fprintln(os.Stderr, "error: internal channel is full, dropping payload to prevent blocking")
			}
			return
		}

		// 3. Flujo de serialización normal
		payload, err := serializer.FormatPayload(cachedHostname, cachedVersion, name, interval, opts, stats)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] error serializing payload: %v\n", name, err)
			return
		}

		// Debug.
		// fmt.Println(string(payload))

		// Enviar al canal que dispacher lee.
		select {
		case metricsChan <- payload:
		default:
			fmt.Fprintln(os.Stderr, "error: internal channel is full, dropping payload to prevent blocking")
		}

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
		fmt.Fprintf(os.Stderr, "[%s] on %s: functionality unavailable in this hardware (interval: %s)\n", name, cachedHostname, interval)
	}

	return serializer.FormatPayloadError(cachedHostname, cachedVersion, name, interval, opts, err)
}

// Consigue información extra una sola vez.
func SetMetadata(hostname, version string) {
	cachedHostname = hostname
	cachedVersion = version
}
