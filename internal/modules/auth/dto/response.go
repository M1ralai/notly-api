package dto

import "time"

type AuthResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID         int       `json:"id"`
	Email      string    `json:"email"`
	FullName   string    `json:"full_name,omitempty"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	Timezone   string    `json:"timezone"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}
