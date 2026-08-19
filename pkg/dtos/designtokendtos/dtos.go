package designtokendtos

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

var (
	ErrTokensNotFound  = errors.New("no hay tokens configurados para este producer")
	ErrTokenJSONBroken = errors.New("uno de los campos no es JSON válido")
)

type RequestUpdateDesignTokens struct {
	Colors  json.RawMessage `json:"colors"`
	Fonts   json.RawMessage `json:"fonts"`
	Radius  json.RawMessage `json:"radius"`
	Shadows json.RawMessage `json:"shadows"`
}

// Validate garantiza que cada campo enviado sea JSON bien formado. El shape
// interno lo decide el frontend; acá solo cerramos el riesgo de guardar un
// blob roto que después reviente el render del tenant.
func (r RequestUpdateDesignTokens) Validate() error {
	for _, raw := range []json.RawMessage{r.Colors, r.Fonts, r.Radius, r.Shadows} {
		if len(raw) > 0 && !json.Valid(raw) {
			return ErrTokenJSONBroken
		}
	}
	return nil
}

type ResponseDesignTokens struct {
	ID         uint            `json:"id"`
	ProducerID uint            `json:"producer_id"`
	Colors     json.RawMessage `json:"colors,omitempty"`
	Fonts      json.RawMessage `json:"fonts,omitempty"`
	Radius     json.RawMessage `json:"radius,omitempty"`
	Shadows    json.RawMessage `json:"shadows,omitempty"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

func (r *ResponseDesignTokens) FromEntity(e entities.DesignTokens) {
	r.ID = e.ID
	r.ProducerID = e.ProducerID
	r.Colors = e.Colors
	r.Fonts = e.Fonts
	r.Radius = e.Radius
	r.Shadows = e.Shadows
	r.UpdatedAt = e.UpdatedAt
}
