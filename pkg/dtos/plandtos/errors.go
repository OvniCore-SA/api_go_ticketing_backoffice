package plandtos

import "errors"

var (
	ErrPlanNotFound      = errors.New("plan no encontrado")
	ErrPlanCodeExists    = errors.New("ya existe un plan con ese código")
	ErrPlanHasProducers  = errors.New("el plan está siendo usado por al menos un producer")
	ErrPlanCodeRequired  = errors.New("el código del plan es requerido")
	ErrPlanNameRequired  = errors.New("el nombre del plan es requerido")
)
