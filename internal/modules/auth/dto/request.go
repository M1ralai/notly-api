package dto

type LoginRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	TurnstileToken string `json:"turnstile_token,omitempty"`
}

type RegisterRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	FullName       string `json:"full_name,omitempty"`
	Timezone       string `json:"timezone,omitempty"`
	TurnstileToken string `json:"turnstile_token,omitempty"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

type ResendCodeRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
