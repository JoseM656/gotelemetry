package serializer

import (
	"encoding/json"
	"time"
)

// agrupa los campos opcionales del contrato
type PayloadOptions struct {
	LocalTimestamp string
	PrettyInterval string
}

type SerializerPayload struct {
	Hostname       string `json:"hostname"`
	Version        string `json:"version"`
	Timestamp      string `json:"timestamp"` // UTC por default.
	LocalTimestamp string `json:"local_timestamp,omitempty"`
	Component      string `json:"component"`
	Interval       int64  `json:"interval"`                  // intervalo en milisegundos
	PrettyInterval string `json:"pretty_interval,omitempty"` // Intervalo formateado de forma bonita.
	Data           any    `json:"data,omitempty"`            // Se omite si hay error
	Error          string `json:"error,omitempty"`           // Se omite si todo salió bien
}

// FormatPayload toma los datos crudos del colector y los envuelve en el JSON estándar
func FormatPayload(hostname string, version string, component string, interval time.Duration, opts PayloadOptions, data any) ([]byte, error) {
	payload := SerializerPayload{
		Hostname:       hostname,
		Version:        version,
		Timestamp:      time.Now().UTC().Format(time.RFC3339), // UTC hora default
		LocalTimestamp: opts.LocalTimestamp,
		Component:      component,
		Interval:       interval.Milliseconds(),
		PrettyInterval: opts.PrettyInterval,
		Data:           data,
	}

	// Convierte la estructura completa a JSON en memoria ([]byte)
	return json.Marshal(payload)
}

// idem, para JSON que muevan errores de lectura.
func FormatPayloadError(hostname string, version string, name string, interval time.Duration, opts PayloadOptions, err error) ([]byte, error) {

	payload := SerializerPayload{
		Hostname:       hostname,
		Version:        version,
		Timestamp:      time.Now().UTC().Format(time.RFC3339), // UTC hora default
		LocalTimestamp: opts.LocalTimestamp,
		Component:      name,
		Interval:       interval.Milliseconds(),
		PrettyInterval: opts.PrettyInterval,
		Error:          err.Error(),
	}

	return json.Marshal(payload)
}
