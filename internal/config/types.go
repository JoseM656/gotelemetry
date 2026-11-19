package config

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	General    GeneralConfig    `yaml:"general"`
	Output     OutputConfig     `yaml:"output"`
	Collectors CollectorsConfig `yaml:"collectors"`
}

// Configuracion de salida y demas.
type GeneralConfig struct {
	LocalTime    BooleanStandar `yaml:"local_timestamp"`        // Struct para activar el horario local.
	PrettyOutput BooleanStandar `yaml:"pretty_interval_output"` // Struct para activar una salida mas legible del intervalo en el JSON.
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

// Elegir que exporters usar.
type OutputConfig struct {
	LOG    LogExporter    `yaml:"logger"`
	STDOUT BooleanStandar `yaml:"stdout"`
}

// Struct generico para todos los que solo requieran un booleano.
type BooleanStandar struct {
	Enabled bool `yaml:"enabled"`
}

type CollectorConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Interval Duration `yaml:"interval"`
}

type LogExporter struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
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
			LocalTime:    BooleanStandar{Enabled: false},
			PrettyOutput: BooleanStandar{Enabled: true},
		},

		Collectors: CollectorsConfig{
			CPU:     CollectorConfig{Enabled: true, Interval: defaultInterval},
			RAM:     CollectorConfig{Enabled: true, Interval: defaultInterval},
			Swap:    CollectorConfig{Enabled: true, Interval: defaultInterval},
			Storage: CollectorConfig{Enabled: false, Interval: defaultInterval},
			GPU:     CollectorConfig{Enabled: false, Interval: defaultInterval},
		},

		Output: OutputConfig{
			LOG:    LogExporter{Enabled: false, Path: ""},
			STDOUT: BooleanStandar{Enabled: true},
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
