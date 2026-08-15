package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config es el árbol completo leído del .yml. Por ahora solo describe, por
// cada collector, si está habilitado y cada cuánto se lee — nada de
// puertos, rutas de log ni flags de dashboard todavía; eso se suma cuando
// existan los exporters que realmente lo necesiten.
type Config struct {
	Collectors CollectorsConfig `yaml:"collectors"`
}

type CollectorsConfig struct {
	CPU     CollectorConfig `yaml:"cpu"`
	RAM     CollectorConfig `yaml:"ram"`
	Swap    CollectorConfig `yaml:"swap"`
	Storage CollectorConfig `yaml:"storage"`
	GPU     CollectorConfig `yaml:"gpu"`
}

type CollectorConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Interval Duration `yaml:"interval"`
}

// Duration envuelve time.Duration para poder escribir "5s" o "1m" en el
// .yml en vez de nanosegundos crudos. yaml.v3 no sabe parsear ese formato
// para time.Duration por sí solo (mismo problema que tiene encoding/json);
// este wrapper con UnmarshalYAML es el punto exacto donde se resuelve, sin
// tener que tocar el resto del struct.
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var raw string
	if err := value.Decode(&raw); err != nil {
		return err
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("config: invalid interval %q: %w", raw, err)
	}

	d.Duration = parsed
	return nil
}

// DefaultConfig es la configuración básica hardcodeada.
func DefaultConfig() Config {
	defaultInterval := Duration{5 * time.Second}

	return Config{
		Collectors: CollectorsConfig{
			CPU:     CollectorConfig{Enabled: true, Interval: defaultInterval},
			RAM:     CollectorConfig{Enabled: true, Interval: defaultInterval},
			Swap:    CollectorConfig{Enabled: true, Interval: defaultInterval},
			Storage: CollectorConfig{Enabled: true, Interval: defaultInterval},
			GPU:     CollectorConfig{Enabled: false, Interval: defaultInterval},
		},
	}
}

// Load lee el .yml en path y lo aplica sobre DefaultConfig(): si el archivo
// no existe, devuelve el default junto con ErrConfigNotFound para que el
// caller decida qué hacer.
func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, ErrConfigNotFound
		}
		return cfg, fmt.Errorf("config: reading %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config: parsing %s: %w", path, err)
	}

	return cfg, nil
}
