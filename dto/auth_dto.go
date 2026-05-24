package dto

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

type GoogleLoginRequest struct {
	Token string `json:"id_token" validate:"required"`
}

type RefreshTokenRequest struct {
	Token string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	Token string `json:"refresh_token" validate:"required"`
}
