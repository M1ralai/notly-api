package dto

import (
	"time"

	"github.com/M1ralai/notly-api/internal/modules/user/domain"
)

type UserResponse struct {
	ID               int        `json:"id"`
	Email            string     `json:"email"`
	FullName         string     `json:"full_name,omitempty"`
	AvatarURL        string     `json:"avatar_url,omitempty"`
	Timezone         string     `json:"timezone"`
	IsPremium        bool       `json:"is_premium"`
	PremiumPlan      string     `json:"premium_plan"`
	PremiumExpiresAt *time.Time `json:"premium_expires_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func ToUserResponse(u *domain.User) *UserResponse {
	return &UserResponse{
		ID:               u.ID,
		Email:            u.Email,
		FullName:         u.FullName,
		AvatarURL:        u.AvatarURL,
		Timezone:         u.Timezone,
		IsPremium:        u.HasPremiumAccess(),
		PremiumPlan:      u.PremiumPlan,
		PremiumExpiresAt: u.PremiumExpiresAt,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
	}
}

func ToUserResponseList(users []*domain.User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i, u := range users {
		result[i] = ToUserResponse(u)
	}
	return result
}
