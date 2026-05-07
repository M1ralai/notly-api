package repository

import (
	"time"

	"github.com/M1ralai/notly-api/internal/modules/semester/domain"
)

type SemesterModel struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	Name      string    `db:"name"`
	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
	IsCurrent bool      `db:"is_current"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (m *SemesterModel) ToDomain() *domain.Semester {
	return &domain.Semester{
		ID:        m.ID,
		UserID:    m.UserID,
		Name:      m.Name,
		StartDate: m.StartDate,
		EndDate:   m.EndDate,
		IsCurrent: m.IsCurrent,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func FromDomain(s *domain.Semester) *SemesterModel {
	return &SemesterModel{
		ID:        s.ID,
		UserID:    s.UserID,
		Name:      s.Name,
		StartDate: s.StartDate,
		EndDate:   s.EndDate,
		IsCurrent: s.IsCurrent,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
