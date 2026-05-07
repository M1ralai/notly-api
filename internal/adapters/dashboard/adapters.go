package adapters

import (
	"context"

	dashboardDto "github.com/M1ralai/notly-api/internal/modules/dashboard/dto"
	habitService "github.com/M1ralai/notly-api/internal/modules/habit/service"
	lifeareaService "github.com/M1ralai/notly-api/internal/modules/lifearea/service"
	taskService "github.com/M1ralai/notly-api/internal/modules/task/service"
)

// TaskAdapter implements dashboard.TaskFetcher
type TaskAdapter struct {
	Service taskService.TaskService
}

func NewTaskAdapter(s taskService.TaskService) *TaskAdapter {
	return &TaskAdapter{Service: s}
}

func (a *TaskAdapter) FetchTasks(ctx context.Context, userID int) ([]*dashboardDto.TaskResponse, error) {
	tasks, err := a.Service.GetAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Map taskDto to dashboardDto
	result := make([]*dashboardDto.TaskResponse, len(tasks))
	for i, t := range tasks {
		result[i] = &dashboardDto.TaskResponse{
			ID:                   t.ID,
			UserID:               t.UserID,
			ParentTaskID:         t.ParentTaskID,
			Title:                t.Title,
			Description:          t.Description,
			DueDate:              t.DueDate,
			EstimatedStart:       t.EstimatedStart,
			EstimatedEnd:         t.EstimatedEnd,
			ActualStart:          t.ActualStart,
			ActualEnd:            t.ActualEnd,
			Priority:             t.Priority,
			IsCompleted:          t.IsCompleted,
			CompletedAt:          t.CompletedAt,
			ProgressPercentage:   t.ProgressPercentage,
			CompletedSubtasks:    t.CompletedSubtasks,
			TotalSubtasks:        t.TotalSubtasks,
			EstimatedDurationMin: t.EstimatedDurationMin,
			ActualDurationMin:    t.ActualDurationMin,
			IsOverdue:            t.IsOverdue,
			CreatedAt:            t.CreatedAt,
			UpdatedAt:            t.UpdatedAt,
		}
	}
	return result, nil
}

func (a *TaskAdapter) FetchStats(ctx context.Context, userID int) (*dashboardDto.TaskStatsResponse, error) {
	stats, err := a.Service.GetStats(ctx, userID)
	if err != nil {
		return nil, err
	}
	if stats == nil {
		return nil, nil
	}
	return &dashboardDto.TaskStatsResponse{
		CompletedToday: stats.CompletedToday,
		DueToday:       stats.DueToday,
		DueTomorrow:    stats.DueTomorrow,
		Overdue:        stats.Overdue,
	}, nil
}

// HabitAdapter implements dashboard.HabitFetcher
type HabitAdapter struct {
	Service habitService.HabitService
}

func NewHabitAdapter(s habitService.HabitService) *HabitAdapter {
	return &HabitAdapter{Service: s}
}

func (a *HabitAdapter) FetchHabits(ctx context.Context, userID int) ([]*dashboardDto.HabitResponse, error) {
	habits, err := a.Service.GetAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*dashboardDto.HabitResponse, len(habits))
	for i, h := range habits {
		result[i] = &dashboardDto.HabitResponse{
			ID:             h.ID,
			UserID:         h.UserID,
			LifeAreaID:     h.LifeAreaID,
			Title:          h.Title,
			Icon:           h.Icon,
			Description:    h.Description,
			Frequency:      h.Frequency,
			FrequencyDays:  h.FrequencyDays,
			IntervalDays:   h.IntervalDays,
			TargetCount:    h.TargetCount,
			CurrentStreak:  h.CurrentStreak,
			BestStreak:     h.BestStreak,
			LongestStreak:  h.LongestStreak,
			IsActive:       h.IsActive,
			CompletedToday: h.CompletedToday,
			SkippedToday:   h.SkippedToday,
			TimeOfDay:      h.TimeOfDay,
			ReminderTime:   h.ReminderTime,
			CreatedAt:      h.CreatedAt,
			UpdatedAt:      h.UpdatedAt,
		}
	}
	return result, nil
}

// LifeAreaAdapter implements dashboard.LifeAreaFetcher
type LifeAreaAdapter struct {
	Service lifeareaService.LifeAreaService
}

func NewLifeAreaAdapter(s lifeareaService.LifeAreaService) *LifeAreaAdapter {
	return &LifeAreaAdapter{Service: s}
}

func (a *LifeAreaAdapter) FetchLifeAreas(ctx context.Context, userID int) ([]*dashboardDto.LifeAreaResponse, error) {
	areas, err := a.Service.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*dashboardDto.LifeAreaResponse, len(areas))
	for i, la := range areas {
		result[i] = &dashboardDto.LifeAreaResponse{
			ID:           la.ID,
			UserID:       la.UserID,
			Name:         la.Name,
			Icon:         la.Icon,
			Color:        la.Color,
			DisplayOrder: la.DisplayOrder,
			CreatedAt:    la.CreatedAt,
		}
	}
	return result, nil
}
