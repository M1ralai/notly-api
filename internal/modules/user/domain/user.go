package domain

import "time"

type User struct {
	ID               int
	Email            string
	PasswordHash     string
	FullName         string
	AvatarURL        string
	Timezone         string
	IsVerified       bool
	IsPremium        bool
	PremiumPlan      string
	PremiumExpiresAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (u *User) IsValid() bool {
	return u.Email != "" && u.PasswordHash != ""
}

func (u *User) HasPremiumAccess() bool {
	if !u.IsPremium {
		return false
	}
	if u.PremiumExpiresAt == nil {
		return true
	}
	return u.PremiumExpiresAt.After(time.Now())
}
