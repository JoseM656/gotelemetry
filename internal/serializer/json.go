package serializer

import (
	"encoding/json"
	"time"
)

type SerializerPayload struct {
	Hostname       string `json:"hostname"`
	Version        string `json:"version"`
	Timestamp      string `json:"timestamp"` // UTC por default.
	LocalTimestamp string `json:"local_timestamp,omitempty"`
	Component      string `json:"component"`
	Interval       string `json:"interval"`
	Data           any    `json:"data,omitempty"`  // Se omite si hay error
	Error          string `json:"error,omitempty"` // Se omite si todo salió bien
}

// FormatPayload toma los datos crudos del colector y los envuelve en el JSON estándar
func FormatPayload(hostname string, version string, component string, interval time.Duration, data any, localtimestamp string) ([]byte, error) {
	payload := SerializerPayload{
		Hostname:       hostname,
		Version:        version,
		Timestamp:      time.Now().UTC().Format(time.RFC3339), // UTC hora default
		LocalTimestamp: localtimestamp,
		Component:      component,
		Interval:       interval.String(),
		Data:           data,
	}

	// Convierte la estructura completa a JSON en memoria ([]byte)
	return json.Marshal(payload)
}

// idem, para JSON que muevan errores de lectura.
func FormatPayloadError(hostname string, version string, name string, interval time.Duration, err error, localtimestamp string) ([]byte, error) {

	payload := SerializerPayload{
		Hostname:       hostname,
		Version:        version,
		Timestamp:      time.Now().UTC().Format(time.RFC3339), // UTC hora default
		LocalTimestamp: localtimestamp,
		Component:      name,
		Interval:       interval.String(),
		Error:          err.Error(),
	}

	return json.Marshal(payload)
}
