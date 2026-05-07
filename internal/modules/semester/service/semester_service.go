package service

import (
	"context"

	"github.com/M1ralai/notly-api/internal/modules/semester/dto"
)

type SemesterService interface {
	Create(ctx context.Context, req *dto.CreateSemesterRequest, userID int) (*dto.SemesterResponse, error)
	GetByID(ctx context.Context, id, userID int) (*dto.SemesterResponse, error)
	GetAll(ctx context.Context, userID int) ([]*dto.SemesterResponse, error)
	GetCurrent(ctx context.Context, userID int) (*dto.SemesterResponse, error)
	Update(ctx context.Context, id int, req *dto.UpdateSemesterRequest, userID int) (*dto.SemesterResponse, error)
	Delete(ctx context.Context, id, userID int) error
}
