package producerdtos

import "errors"

var (
	ErrProducerNotFound = errors.New("producer no encontrado")
	ErrSlugRequired     = errors.New("el slug es requerido")
	ErrNameRequired     = errors.New("el nombre es requerido")
	ErrSlugExists       = errors.New("ya existe un producer con ese slug")
	ErrInvalidStatus    = errors.New("el estado no es válido")
	ErrTemplateNotFound = errors.New("la template referenciada no existe")
	ErrPlanNotFound     = errors.New("el plan referenciado no existe")
	ErrEmailInvalid     = errors.New("el email de contacto no es válido")
	ErrSlugInvalid      = errors.New("el slug debe contener solo letras minúsculas, números y guiones; debe empezar y terminar con letra o número")
)
