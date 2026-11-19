package exporter

import "fmt"

// StdoutExport imprime el payload directamente por la salida estándar
func StdoutExport(payload []byte) {
	fmt.Println(string(payload))
}
