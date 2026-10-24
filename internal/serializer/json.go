package serializer

import (
	"encoding/json"
	"time"
)

type SerializerPayload struct {
	Timestamp string `json:"timestamp"`
	Component string `json:"component"`
	Interval  string `json:"interval"`
	Data      any    `json:"data,omitempty"`  // Se omite si hay error
	Error     string `json:"error,omitempty"` // Se omite si todo salió bien
}

// FormatPayload toma los datos crudos del colector y los envuelve en el JSON estándar
func FormatPayload(component string, interval time.Duration, data any) ([]byte, error) {
	payload := SerializerPayload{
		Timestamp: time.Now().Format(time.RFC3339), // El serializador pone la hora exacta
		Component: component,
		Interval:  interval.String(),
		Data:      data,
	}

	// Convierte la estructura completa a JSON en memoria ([]byte)
	return json.Marshal(payload)
}

// idem, para JSON que muevan errores de lectura.
func FormatPayloadError(name string, interval time.Duration, err error) ([]byte, error) {

	payload := SerializerPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Component: name,
		Interval:  interval.String(),
		Error:     err.Error(),
	}

	return json.Marshal(payload)
}
