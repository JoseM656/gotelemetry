package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// saveDefault escribe la estructura DefaultConfig() en el disco codificada como YAML
func SaveDefault(path string) error {

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
			if saveErr := SaveDefault(path); saveErr != nil {
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
