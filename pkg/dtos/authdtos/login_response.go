package authdtos

type LoginResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         ResponseUser `json:"user"`
}
