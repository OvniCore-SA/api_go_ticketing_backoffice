package templatedtos

import "errors"

var (
	ErrTemplateNotFound   = errors.New("template no encontrada")
	ErrTemplateCodeExists = errors.New("ya existe una template con ese código")
	ErrTemplateCodeReq    = errors.New("el código de la template es requerido")
	ErrTemplateNameReq    = errors.New("el nombre de la template es requerido")
	ErrPageTypeInvalid    = errors.New("el tipo de página no es válido")
	ErrModuleNotFound     = errors.New("uno o más módulos referenciados no existen")
	ErrTemplateInUse      = errors.New("la template está siendo usada por al menos un producer")
	ErrTemplateJSONBroken = errors.New("uno de los campos JSON de la template no es válido")
)
