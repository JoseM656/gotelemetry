package runner

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/JoseM656/gotelemetry/internal/config"
	"github.com/JoseM656/gotelemetry/internal/exporter"
)

func runExporterWorker(
	ctx context.Context,
	ch <-chan []byte,
	exportFn func([]byte) error,
	errChan chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case payload, ok := <-ch:
			if !ok {
				return
			}
			if err := exportFn(payload); err != nil && errChan != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}
		}
	}
}

func StartDispatcher(ctx context.Context, cfg config.Config, metricsChan <-chan []byte, cancel context.CancelFunc) {
	stdoutEnabled := cfg.Output.STDOUT.Enabled
	loggerEnabled := cfg.Output.LOG.Enabled

	// 1. Validación fail-fast al arrancar
	if !stdoutEnabled && !loggerEnabled {
		fmt.Fprintln(os.Stderr, "error: there isn't selected an exporter for telemetry. Check config file. Exiting...")
		cancel()
		return
	}

	if loggerEnabled && cfg.Output.LOG.Path == "" {
		fmt.Fprintf(os.Stderr, "error: log path is empty, add one or disable this exporter. Exiting...\n")
		cancel()
		return
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	var stdoutChan chan []byte
	if stdoutEnabled {
		fmt.Fprintln(os.Stderr, "exporting using stdout...")
		stdoutChan = make(chan []byte, 64)
		wg.Add(1)
		go runExporterWorker(
			ctx,
			stdoutChan,
			func(b []byte) error {
				exporter.StdoutExport(b)
				return nil
			},
			nil,
			&wg,
		)
	}

	var logChan chan []byte
	if loggerEnabled {
		fmt.Fprintf(os.Stderr, "exporting logs in %s\n", cfg.Output.LOG.Path)
		logChan = make(chan []byte, 64)
		wg.Add(1)
		go runExporterWorker(
			ctx,
			logChan,
			func(b []byte) error {
				if err := exporter.LogExport(b, cfg.Output.LOG.Path); err != nil {
					return fmt.Errorf("log exporter: %w", err)
				}
				return nil
			},
			errChan,
			&wg,
		)
	}

	// Cierre limpio de canales propios y espera de workers al salir del dispatcher
	defer func() {
		if stdoutChan != nil {
			close(stdoutChan)
		}
		if logChan != nil {
			close(logChan)
		}
		wg.Wait()

		// Cerrar el archivo si logger esta trabajando.
		if loggerEnabled {
			if err := exporter.CloseLogFile(); err != nil {
				fmt.Fprintf(os.Stderr, "error closing log file: %v\n", err)
			}
		}
	}()

	// 2. Bucle despachador principal
	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr, "exporters shutting down...")
			return

		case err := <-errChan:
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			cancel()
			return

		case payload, ok := <-metricsChan:
			if !ok {
				cancel()
				return
			}

			if stdoutEnabled {
				select {
				case stdoutChan <- payload:
				case <-ctx.Done():
					return
				}
			}

			if loggerEnabled {
				select {
				case logChan <- payload:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
