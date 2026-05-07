package dto

type CreateSemesterRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=255"`
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
	IsCurrent *bool  `json:"is_current,omitempty"`
}

type UpdateSemesterRequest struct {
	Name      *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
	IsCurrent *bool   `json:"is_current,omitempty"`
}
