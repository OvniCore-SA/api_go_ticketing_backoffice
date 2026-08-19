package filters

type DomainFilter struct {
	ID         uint
	ProducerID uint
	Domain     string
	Type       string
}

type CommissionFilter struct {
	ID         uint
	ProducerID uint
	CurrentOnly bool
}
