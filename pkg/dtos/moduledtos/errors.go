package moduledtos

import "errors"

var (
	ErrModuleNotFound  = errors.New("módulo no encontrado")
	ErrVariantNotFound = errors.New("variante no encontrada")
)
