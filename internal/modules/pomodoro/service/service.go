package service

import (
	"context"

	"github.com/M1ralai/notly-api/internal/modules/pomodoro/domain"
	"github.com/M1ralai/notly-api/internal/modules/pomodoro/dto"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSession(ctx context.Context, userID int, req dto.CreatePomodoroSessionRequest) (*dto.PomodoroSessionResponse, error) {
	session := &domain.PomodoroSession{
		UserID:          userID,
		CourseID:        req.CourseID,
		DurationMinutes: req.DurationMinutes,
		Notes:           req.Notes,
	}

	if err := s.repo.Save(ctx, session); err != nil {
		return nil, err
	}

	return dto.ToSessionResponse(session), nil
}

func (s *Service) GetSessions(ctx context.Context, userID int) ([]*dto.PomodoroSessionResponse, error) {
	sessions, err := s.repo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.ToSessionResponseList(sessions), nil
}

func (s *Service) GetCourseSessions(ctx context.Context, userID, courseID int) ([]*dto.PomodoroSessionResponse, error) {
	sessions, err := s.repo.FindByCourseID(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	return dto.ToSessionResponseList(sessions), nil
}

func (s *Service) GetSettings(ctx context.Context, userID int) (*dto.PomodoroSettingsResponse, error) {
	settings, err := s.repo.GetSettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	return dto.ToSettingsResponse(settings), nil
}

func (s *Service) UpdateSettings(ctx context.Context, userID int, req dto.UpdatePomodoroSettingsRequest) (*dto.PomodoroSettingsResponse, error) {
	settings := &domain.PomodoroSettings{
		UserID:       userID,
		StudyPresets: req.StudyPresets,
		StudyColor:   req.StudyColor,
	}

	if err := s.repo.UpdateSettings(ctx, settings); err != nil {
		return nil, err
	}

	return dto.ToSettingsResponse(settings), nil
}
