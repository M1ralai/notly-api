package dto

import (
	"time"

	"github.com/M1ralai/notly-api/internal/modules/semester/domain"
)

type SemesterResponse struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	IsCurrent bool      `json:"is_current"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToSemesterResponse(s *domain.Semester) *SemesterResponse {
	if s == nil {
		return nil
	}
	return &SemesterResponse{
		ID:        s.ID,
		UserID:    s.UserID,
		Name:      s.Name,
		StartDate: s.StartDate.Format("2006-01-02"),
		EndDate:   s.EndDate.Format("2006-01-02"),
		IsCurrent: s.IsCurrent,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func ToSemesterResponseList(semesters []*domain.Semester) []*SemesterResponse {
	if semesters == nil {
		return nil
	}
	result := make([]*SemesterResponse, len(semesters))
	for i, s := range semesters {
		result[i] = ToSemesterResponse(s)
	}
	return result
}
