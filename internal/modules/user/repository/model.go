package repository

import (
	"time"

	"github.com/M1ralai/notly-api/internal/modules/user/domain"
)

type UserModel struct {
	ID               int        `db:"id"`
	Email            string     `db:"email"`
	PasswordHash     string     `db:"password_hash"`
	FullName         *string    `db:"full_name"`
	AvatarURL        *string    `db:"avatar_url"`
	Timezone         string     `db:"timezone"`
	IsVerified       bool       `db:"is_verified"`
	IsPremium        bool       `db:"is_premium"`
	PremiumPlan      string     `db:"premium_plan"`
	PremiumExpiresAt *time.Time `db:"premium_expires_at"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

func (m *UserModel) ToDomain() *domain.User {
	fullName := ""
	if m.FullName != nil {
		fullName = *m.FullName
	}
	avatarURL := ""
	if m.AvatarURL != nil {
		avatarURL = *m.AvatarURL
	}
	premiumPlan := m.PremiumPlan
	if premiumPlan == "" {
		premiumPlan = "free"
	}

	return &domain.User{
		ID:               m.ID,
		Email:            m.Email,
		PasswordHash:     m.PasswordHash,
		FullName:         fullName,
		AvatarURL:        avatarURL,
		Timezone:         m.Timezone,
		IsVerified:       m.IsVerified,
		IsPremium:        m.IsPremium,
		PremiumPlan:      premiumPlan,
		PremiumExpiresAt: m.PremiumExpiresAt,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func FromDomain(u *domain.User) *UserModel {
	var fullName, avatarURL *string
	if u.FullName != "" {
		fullName = &u.FullName
	}
	if u.AvatarURL != "" {
		avatarURL = &u.AvatarURL
	}

	return &UserModel{
		ID:               u.ID,
		Email:            u.Email,
		PasswordHash:     u.PasswordHash,
		FullName:         fullName,
		AvatarURL:        avatarURL,
		Timezone:         u.Timezone,
		IsVerified:       u.IsVerified,
		IsPremium:        u.IsPremium,
		PremiumPlan:      u.PremiumPlan,
		PremiumExpiresAt: u.PremiumExpiresAt,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
	}
}
