package config

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	General    GeneralConfig    `yaml:"general"`
	Collectors CollectorsConfig `yaml:"collectors"`
}

// Configuracion general
type GeneralConfig struct {
	LocalTime LocalTimestamp `yaml:"local_timestamp"`
	// otras posibles config...
}

// Structs para los collectores.
type CollectorsConfig struct {
	CPU     CollectorConfig `yaml:"cpu"`
	RAM     CollectorConfig `yaml:"ram"`
	Swap    CollectorConfig `yaml:"swap"`
	Storage CollectorConfig `yaml:"storage"`
	GPU     CollectorConfig `yaml:"gpu"`
}

// Struct para activar el horario local.
type LocalTimestamp struct {
	Enable bool `yaml:"enable"`
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

		General: GeneralConfig{
			LocalTime: LocalTimestamp{Enable: false},
		},

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
