package serializer

import (
	"encoding/json"
	"time"
)

type SerializerPayload struct {
	Timestamp string `json:"timestamp"`
	Component string `json:"component"`
	Interval  string `json:"interval"`
	Data      any    `json:"data"`
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
