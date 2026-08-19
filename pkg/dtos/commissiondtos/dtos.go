package commissiondtos

import (
	"errors"
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

var (
	ErrPercentageInvalid = errors.New("el porcentaje debe estar entre 0 y 100")
	ErrValidFromRequired = errors.New("valid_from es requerido")
)

type RequestCreateCommission struct {
	Percentage float64   `json:"percentage"`
	ValidFrom  time.Time `json:"valid_from"`
}

func (r *RequestCreateCommission) Validate() error {
	if r.Percentage < 0 || r.Percentage > 100 {
		return ErrPercentageInvalid
	}
	if r.ValidFrom.IsZero() {
		return ErrValidFromRequired
	}
	return nil
}

type ResponseCommission struct {
	ID         uint       `json:"id"`
	ProducerID uint       `json:"producer_id"`
	Percentage float64    `json:"percentage"`
	ValidFrom  time.Time  `json:"valid_from"`
	ValidTo    *time.Time `json:"valid_to,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (r *ResponseCommission) FromEntity(e entities.Commission) {
	r.ID = e.ID
	r.ProducerID = e.ProducerID
	r.Percentage = e.Percentage
	r.ValidFrom = e.ValidFrom
	r.ValidTo = e.ValidTo
	r.CreatedAt = e.CreatedAt
}

type ResponseCommissions struct {
	Commissions []ResponseCommission `json:"commissions"`
}

func (r *ResponseCommissions) FromEntities(list []entities.Commission) {
	r.Commissions = make([]ResponseCommission, len(list))
	for i, e := range list {
		r.Commissions[i].FromEntity(e)
	}
}
