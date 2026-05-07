package service

import (
	"context"
	"time"

	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
	courseRepo "github.com/M1ralai/notly-api/internal/modules/course/repository"
	eventRepo "github.com/M1ralai/notly-api/internal/modules/event/repository"
	goalRepo "github.com/M1ralai/notly-api/internal/modules/goal/repository"
	habitRepo "github.com/M1ralai/notly-api/internal/modules/habit/repository"
	lifeareaRepo "github.com/M1ralai/notly-api/internal/modules/lifearea/repository"
	semesterRepo "github.com/M1ralai/notly-api/internal/modules/semester/repository"
	"github.com/M1ralai/notly-api/internal/modules/sync/dto"
	taskRepo "github.com/M1ralai/notly-api/internal/modules/task/repository"
	noteRepo "github.com/M1ralai/notly-api/internal/modules/note/repository"
	"golang.org/x/sync/errgroup"
)

type SyncService interface {
	GetDelta(ctx context.Context, userID int, since time.Time) (*dto.DeltaSyncResponse, error)
}

type syncService struct {
	taskRepo     taskRepo.TaskRepository
	habitRepo    habitRepo.HabitRepository
	lifeareaRepo lifeareaRepo.LifeAreaRepository
	goalRepo     goalRepo.GoalRepository
	courseRepo   courseRepo.CourseRepository
	eventRepo    eventRepo.EventRepository
	semesterRepo semesterRepo.SemesterRepository
	noteRepo     noteRepo.NoteRepository
	logger       *logger.ZapLogger
}

func NewSyncService(
	taskRepo taskRepo.TaskRepository,
	habitRepo habitRepo.HabitRepository,
	lifeareaRepo lifeareaRepo.LifeAreaRepository,
	goalRepo goalRepo.GoalRepository,
	courseRepo courseRepo.CourseRepository,
	eventRepo eventRepo.EventRepository,
	semesterRepo semesterRepo.SemesterRepository,
	noteRepo noteRepo.NoteRepository,
	logger *logger.ZapLogger,
) SyncService {
	return &syncService{
		taskRepo:     taskRepo,
		habitRepo:    habitRepo,
		lifeareaRepo: lifeareaRepo,
		goalRepo:     goalRepo,
		courseRepo:   courseRepo,
		eventRepo:    eventRepo,
		semesterRepo: semesterRepo,
		noteRepo:     noteRepo,
		logger:       logger,
	}
}

func (s *syncService) GetDelta(ctx context.Context, userID int, since time.Time) (*dto.DeltaSyncResponse, error) {
	// G, ctx := errgroup.WithContext(ctx) can be used to run repo queries in parallel

	g, ctx := errgroup.WithContext(ctx)

	response := &dto.DeltaSyncResponse{
		Timestamp: time.Now().UTC(),
		Changes:   dto.ChangesDelta{},
	}

	// 1. Fetch Tasks
	g.Go(func() error {
		updated, err := s.taskRepo.GetUpdatedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Tasks.Updated", err, nil)
			return err
		}
		deleted, err := s.taskRepo.GetDeletedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Tasks.Deleted", err, nil)
			return err
		}

		response.Changes.Tasks = dto.ModuleDelta{
			Updated: updated,
			Deleted: deleted,
		}
		return nil
	})

	// 2. Fetch Habits
	g.Go(func() error {
		updated, err := s.habitRepo.GetUpdatedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Habits.Updated", err, nil)
			return err
		}
		deleted, err := s.habitRepo.GetDeletedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Habits.Deleted", err, nil)
			return err
		}

		response.Changes.Habits = dto.ModuleDelta{
			Updated: updated,
			Deleted: deleted,
		}
		return nil
	})

	// 3. Fetch LifeAreas
	g.Go(func() error {
		updated, err := s.lifeareaRepo.GetUpdatedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.LifeAreas.Updated", err, nil)
			return err
		}
		deleted, err := s.lifeareaRepo.GetDeletedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.LifeAreas.Deleted", err, nil)
			return err
		}
		response.Changes.LifeAreas = dto.ModuleDelta{Updated: updated, Deleted: deleted}
		return nil
	})

	// 4. Fetch Goals
	g.Go(func() error {
		updated, err := s.goalRepo.GetUpdatedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Goals.Updated", err, nil)
			return err
		}
		deleted, err := s.goalRepo.GetDeletedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Goals.Deleted", err, nil)
			return err
		}
		response.Changes.Goals = dto.ModuleDelta{Updated: updated, Deleted: deleted}
		return nil
	})

	// 5. Fetch Courses
	g.Go(func() error {
		updated, err := s.courseRepo.GetUpdatedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Courses.Updated", err, nil)
			return err
		}
		deleted, err := s.courseRepo.GetDeletedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Courses.Deleted", err, nil)
			return err
		}
		response.Changes.Courses = dto.ModuleDelta{Updated: updated, Deleted: deleted}
		return nil
	})

	// 6. Fetch Events
	g.Go(func() error {
		updated, err := s.eventRepo.GetUpdatedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Events.Updated", err, nil)
			return err
		}
		deleted, err := s.eventRepo.GetDeletedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Events.Deleted", err, nil)
			return err
		}
		response.Changes.Events = dto.ModuleDelta{Updated: updated, Deleted: deleted}
		return nil
	})

	// 7. Fetch Semesters
	g.Go(func() error {
		updated, err := s.semesterRepo.GetUpdatedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Semesters.Updated", err, nil)
			return err
		}
		deleted, err := s.semesterRepo.GetDeletedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Semesters.Deleted", err, nil)
			return err
		}
		response.Changes.Semesters = dto.ModuleDelta{Updated: updated, Deleted: deleted}
		return nil
	})

	// 8. Fetch Notes
	g.Go(func() error {
		updated, err := s.noteRepo.GetUpdatedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Notes.Updated", err, nil)
			return err
		}
		deleted, err := s.noteRepo.GetDeletedSince(ctx, userID, since)
		if err != nil {
			s.logger.Error("SyncService.GetDelta.Notes.Deleted", err, nil)
			return err
		}
		response.Changes.Notes = dto.ModuleDelta{Updated: updated, Deleted: deleted}
		return nil
	})

	// Wait for all queries to resolve
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return response, nil
}
