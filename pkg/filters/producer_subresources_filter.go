package filters

type ProducerModuleFilter struct {
	ProducerID uint
	ModuleID   uint
	Enabled    *bool
}

type ProducerComponentVariantFilter struct {
	ProducerID uint
	ModuleID   uint
}

type PageTemplateFilter struct {
	ProducerID uint
	PageType   string
}
