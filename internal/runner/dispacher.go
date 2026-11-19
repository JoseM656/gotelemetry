package runner

import (
	"context"
	"fmt"

	"github.com/JoseM656/gotelemetry/internal/config"
	"github.com/JoseM656/gotelemetry/internal/exporter"
)

// StartDispatcher recibe el contexto, la configuración y el canal compartido.
func StartDispatcher(ctx context.Context, cfg config.Config, metricsChan <-chan []byte, cancel context.CancelFunc) {

	// 1. Validar si hay al menos un exportador activo
	stdoutEnabled := cfg.Output.STDOUT.Enabled
	loggerEnabled := cfg.Output.LOG.Enabled

	// PROVISIONAL: cuando hayan mas opciones capaz un bulque for o un switch.
	if !stdoutEnabled && !loggerEnabled {
		fmt.Println("error: there isn't selected an exporter for telemetry. Check config file. Exiting...")
		cancel()
		return

	} else if stdoutEnabled {
		fmt.Println("exporting using stdout...")

	} else if loggerEnabled {

		if cfg.Output.LOG.Path == "" {
			fmt.Printf("error: log path is empty, add one or disable this exporter. Exiting...\n")
			cancel()
			return
		}

		fmt.Printf("exporting logs in %s\n", cfg.Output.LOG.Path)
	}

	// 2. El bucle consumidor que lee continuamente
	for {
		select {
		case <-ctx.Done():
			fmt.Println("exporters shutting down...")
			return

		case payload := <-metricsChan:

			if stdoutEnabled {
				exporter.StdoutExport(payload)
			}

			if loggerEnabled {

				err := exporter.LogExport(payload, cfg.Output.LOG.Path)
				if err != nil {
					fmt.Printf("error: failed to write log: %v\n", err)
					cancel()
					return
				}
			}
		}
	}

}
