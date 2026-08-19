package filters

type UserFilter struct {
	ID           uint
	Email        string
	RoleCode     string
	SelectFields []string
}
