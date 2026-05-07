package service

import (
	"context"

	"github.com/M1ralai/notly-api/internal/modules/dashboard/dto"
	"golang.org/x/sync/errgroup"
)

// Ports (Interfaces that the dashboard module expects to be fulfilled)
type TaskFetcher interface {
	FetchTasks(ctx context.Context, userID int) ([]*dto.TaskResponse, error)
	FetchStats(ctx context.Context, userID int) (*dto.TaskStatsResponse, error)
}

type HabitFetcher interface {
	FetchHabits(ctx context.Context, userID int) ([]*dto.HabitResponse, error)
}

type LifeAreaFetcher interface {
	FetchLifeAreas(ctx context.Context, userID int) ([]*dto.LifeAreaResponse, error)
}

type DashboardService struct {
	taskFetcher     TaskFetcher
	habitFetcher    HabitFetcher
	lifeAreaFetcher LifeAreaFetcher
}

func NewDashboardService(
	taskFetcher TaskFetcher,
	habitFetcher HabitFetcher,
	lifeAreaFetcher LifeAreaFetcher,
) *DashboardService {
	return &DashboardService{
		taskFetcher:     taskFetcher,
		habitFetcher:    habitFetcher,
		lifeAreaFetcher: lifeAreaFetcher,
	}
}

func (s *DashboardService) GetDashboardData(ctx context.Context, userID int) (*dto.DashboardResponse, error) {
	resp := &dto.DashboardResponse{
		Tasks:     make([]*dto.TaskResponse, 0),
		Habits:    make([]*dto.HabitResponse, 0),
		LifeAreas: make([]*dto.LifeAreaResponse, 0),
		Stats:     &dto.TaskStatsResponse{},
	}

	g, groupCtx := errgroup.WithContext(ctx)

	// Fetch Tasks
	g.Go(func() error {
		tasks, err := s.taskFetcher.FetchTasks(groupCtx, userID)
		if err != nil {
			return err
		}
		resp.Tasks = tasks
		return nil
	})

	// Fetch Stats
	g.Go(func() error {
		stats, err := s.taskFetcher.FetchStats(groupCtx, userID)
		if err != nil {
			return err
		}
		resp.Stats = stats
		return nil
	})

	// Fetch Habits
	g.Go(func() error {
		habits, err := s.habitFetcher.FetchHabits(groupCtx, userID)
		if err != nil {
			return err
		}
		resp.Habits = habits
		return nil
	})

	// Fetch LifeAreas
	g.Go(func() error {
		areas, err := s.lifeAreaFetcher.FetchLifeAreas(groupCtx, userID)
		if err != nil {
			return err
		}
		resp.LifeAreas = areas
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return resp, nil
}
