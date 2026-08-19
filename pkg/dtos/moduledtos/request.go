package moduledtos

type RequestListModules struct {
	Search   string `query:"search"`
	Category string `query:"category"`
	IsCore   *bool  `query:"is_core"`
}
