package filters

type PlanFilter struct {
	ID       uint
	Code     string
	IsActive *bool
	Search   string
}
