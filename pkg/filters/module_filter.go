package filters

type ModuleFilter struct {
	ID       uint
	Code     string
	Category string
	IsCore   *bool
	Search   string
}

type ComponentVariantFilter struct {
	ID       uint
	Code     string
	ModuleID uint
}
