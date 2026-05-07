package service

import (
	"context"
	"errors"
	"time"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
	notifService "github.com/M1ralai/notly-api/internal/modules/notification/service"
	"github.com/M1ralai/notly-api/internal/modules/semester/domain"
	"github.com/M1ralai/notly-api/internal/modules/semester/dto"
	"github.com/M1ralai/notly-api/internal/modules/semester/repository"
)

type semesterServiceImpl struct {
	repo        repository.SemesterRepository
	logger      *logger.ZapLogger
	broadcaster *notifService.Broadcaster
}

func NewSemesterService(
	repo repository.SemesterRepository,
	logger *logger.ZapLogger,
	broadcaster *notifService.Broadcaster,
) SemesterService {
	return &semesterServiceImpl{
		repo:        repo,
		logger:      logger,
		broadcaster: broadcaster,
	}
}

func (s *semesterServiceImpl) Create(ctx context.Context, req *dto.CreateSemesterRequest, userID int) (*dto.SemesterResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("invalid start_date format, expected YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.New("invalid end_date format, expected YYYY-MM-DD")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("end_date must be after start_date")
	}

	isCurrent := false
	if req.IsCurrent != nil {
		isCurrent = *req.IsCurrent
	}

	semester := &domain.Semester{
		UserID:    userID,
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
		IsCurrent: isCurrent,
	}

	created, err := s.repo.Create(ctx, semester)
	if err != nil {
		s.logger.Error("Failed to create semester", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, err
	}

	s.logger.Info("Semester created", map[string]interface{}{
		"semester_id": created.ID,
		"user_id":     userID,
		"name":        created.Name,
	})

	// Broadcast real-time notification
	if s.broadcaster != nil {
		excludeCID := utils.GetConnectionIDFromContext(ctx)
		s.broadcaster.Publish(userID, excludeCID, "semester_created", map[string]interface{}{
			"semester_id": created.ID,
			"name":        created.Name,
		})
	}

	return dto.ToSemesterResponse(created), nil
}

func (s *semesterServiceImpl) GetByID(ctx context.Context, id, userID int) (*dto.SemesterResponse, error) {
	semester, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if semester == nil {
		return nil, errors.New("semester not found")
	}

	if semester.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return dto.ToSemesterResponse(semester), nil
}

func (s *semesterServiceImpl) GetAll(ctx context.Context, userID int) ([]*dto.SemesterResponse, error) {
	semesters, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get semesters", err, map[string]interface{}{
			"user_id": userID,
		})
		return nil, err
	}

	return dto.ToSemesterResponseList(semesters), nil
}

func (s *semesterServiceImpl) GetCurrent(ctx context.Context, userID int) (*dto.SemesterResponse, error) {
	semester, err := s.repo.GetCurrent(ctx, userID)
	if err != nil {
		return nil, err
	}

	if semester == nil {
		return nil, nil
	}

	return dto.ToSemesterResponse(semester), nil
}

func (s *semesterServiceImpl) Update(ctx context.Context, id int, req *dto.UpdateSemesterRequest, userID int) (*dto.SemesterResponse, error) {
	semester, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if semester == nil {
		return nil, errors.New("semester not found")
	}

	if semester.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Update fields if provided
	if req.Name != nil {
		semester.Name = *req.Name
	}

	if req.StartDate != nil {
		startDate, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, errors.New("invalid start_date format, expected YYYY-MM-DD")
		}
		semester.StartDate = startDate
	}

	if req.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, errors.New("invalid end_date format, expected YYYY-MM-DD")
		}
		semester.EndDate = endDate
	}

	if req.IsCurrent != nil {
		semester.IsCurrent = *req.IsCurrent
	}

	// Validate dates
	if semester.EndDate.Before(semester.StartDate) {
		return nil, errors.New("end_date must be after start_date")
	}

	if err := s.repo.Update(ctx, semester); err != nil {
		s.logger.Error("Failed to update semester", err, map[string]interface{}{
			"semester_id": id,
			"user_id":     userID,
		})
		return nil, err
	}

	s.logger.Info("Semester updated", map[string]interface{}{
		"semester_id": id,
		"user_id":     userID,
	})

	// Broadcast real-time notification
	if s.broadcaster != nil {
		excludeCID := utils.GetConnectionIDFromContext(ctx)
		s.broadcaster.Publish(userID, excludeCID, "semester_updated", map[string]interface{}{
			"semester_id": id,
		})
	}

	return dto.ToSemesterResponse(semester), nil
}

func (s *semesterServiceImpl) Delete(ctx context.Context, id, userID int) error {
	semester, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if semester == nil {
		return errors.New("semester not found")
	}

	if semester.UserID != userID {
		return errors.New("unauthorized")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete semester", err, map[string]interface{}{
			"semester_id": id,
			"user_id":     userID,
		})
		return err
	}

	s.logger.Info("Semester deleted", map[string]interface{}{
		"semester_id": id,
		"user_id":     userID,
	})

	// Broadcast real-time notification
	if s.broadcaster != nil {
		excludeCID := utils.GetConnectionIDFromContext(ctx)
		s.broadcaster.Publish(userID, excludeCID, "semester_deleted", map[string]interface{}{
			"semester_id": id,
		})
	}

	return nil
}
