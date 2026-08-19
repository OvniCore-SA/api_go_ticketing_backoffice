package entities

import (
	"time"

	"gorm.io/gorm"
)

// Commission representa el porcentaje de comisión vigente para un Producer.
// El histórico se conserva: cada alta cierra la vigencia anterior (ValidTo)
// dentro de una transacción, así siempre hay a lo sumo un registro con
// ValidTo NULL por producer.
type Commission struct {
	gorm.Model

	ProducerID uint       `gorm:"column:producer_id;not null"`
	Percentage float64    `gorm:"column:percentage;not null"`
	ValidFrom  time.Time  `gorm:"column:valid_from;not null"`
	ValidTo    *time.Time `gorm:"column:valid_to"`

	Producer Producer `gorm:"foreignKey:ProducerID"`
}
