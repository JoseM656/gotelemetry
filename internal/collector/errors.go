package collector

import "errors"

// Errores sentinelas, sriven para definir errores globales que se pueden traer en varios archivos.

// En caso de que componentes como una GPU no esten disponibles.
var ErrCapabilityUnavailable = errors.New("capability not available on this hardware")
