package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Collectors CollectorsConfig `yaml:"collectors"`
}

// Misma declaracion de structs pero diferentes datos, reutilizacion correcta.
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
// .yml en vez de nanosegundos crudos.
type Duration struct {
	time.Duration
}

// DefaultConfig es la configuración básica hardcodeada.
func DefaultConfig() Config {
	defaultInterval := Duration{5 * time.Second}

	return Config{
		Collectors: CollectorsConfig{
			CPU:     CollectorConfig{Enabled: true, Interval: defaultInterval},
			RAM:     CollectorConfig{Enabled: true, Interval: defaultInterval},
			Swap:    CollectorConfig{Enabled: true, Interval: defaultInterval},
			Storage: CollectorConfig{Enabled: false, Interval: defaultInterval},
			GPU:     CollectorConfig{Enabled: false, Interval: defaultInterval},
		},
	}
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
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

// saveDefault escribe la estructura DefaultConfig() en el disco codificada como YAML
func saveDefault(path string) error {

	defaultCfg := DefaultConfig()

	data, err := yaml.Marshal(defaultCfg)
	if err != nil {
		return fmt.Errorf("config: Error serializing default configuration: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0775); err != nil {
		return fmt.Errorf("config: error recreating %s: %w", path, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("config: error recreating %s: %w", path, err)
	}

	return nil
}

// Load lee la configuración. Si no existe, intenta regenerarla.
// Si fue regenerada con éxito, devuelve (cfg, nil, true).
func Load(path string) (cfg Config, created bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = DefaultConfig()
			if saveErr := saveDefault(path); saveErr != nil {
				return cfg, false, fmt.Errorf("config: file missing and recreation failed: %w", saveErr)
			}
			return cfg, true, nil // true indica que se creó un archivo nuevo
		}
		return DefaultConfig(), false, fmt.Errorf("config: reading %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), false, fmt.Errorf("config: parsing %s: %w", path, err)
	}

	return cfg, false, nil
}
