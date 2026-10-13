package main

import (
	"fmt"
	"os"

	"github.com/JoseM656/gotelemetry/internal/config"
	"github.com/JoseM656/gotelemetry/internal/runner"
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

	// Cargar configuracion en runner
	runner.Run(cfg)

}
