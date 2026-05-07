package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
	calendarService "github.com/M1ralai/notly-api/internal/modules/calendar/service"
	"github.com/M1ralai/notly-api/internal/modules/course/domain"
	"github.com/M1ralai/notly-api/internal/modules/course/dto"
	"github.com/M1ralai/notly-api/internal/modules/course/repository"
	"github.com/M1ralai/notly-api/internal/modules/notification"
	notifService "github.com/M1ralai/notly-api/internal/modules/notification/service"
	semesterRepo "github.com/M1ralai/notly-api/internal/modules/semester/repository"
	userRepo "github.com/M1ralai/notly-api/internal/modules/user/repository"
)

// normalizeTime ensures time string is in HH:MM format
// Handles various input formats and converts to PostgreSQL TIME format
func normalizeTime(timeStr string) string {
	if timeStr == "" {
		return "00:00"
	}

	// Remove whitespace
	timeStr = strings.TrimSpace(timeStr)

	// Split by colon
	parts := strings.Split(timeStr, ":")
	if len(parts) < 2 {
		return "00:00"
	}

	// Parse hour and minute
	var hour, minute int
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)

	// Validate ranges
	if hour < 0 || hour > 23 {
		hour = 0
	}
	if minute < 0 || minute > 59 {
		minute = 0
	}

	// Format as HH:MM
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

type courseService struct {
	repo            repository.CourseRepository
	semesterRepo    semesterRepo.SemesterRepository
	calendarService calendarService.CalendarService
	logger          *logger.ZapLogger
	broadcaster     *notifService.Broadcaster
	userRepo        userRepo.UserRepository
}

func NewCourseService(
	repo repository.CourseRepository,
	semesterRepo semesterRepo.SemesterRepository,
	calendarService calendarService.CalendarService,
	logger *logger.ZapLogger,
	broadcaster *notifService.Broadcaster,
	userRepo userRepo.UserRepository,
) CourseService {
	return &courseService{
		repo:            repo,
		semesterRepo:    semesterRepo,
		calendarService: calendarService,
		logger:          logger,
		broadcaster:     broadcaster,
		userRepo:        userRepo,
	}
}

func (s *courseService) Create(ctx context.Context, req *dto.CreateCourseRequest, userID int) (*dto.CourseResponse, error) {
	s.logger.Info("Creating course", map[string]interface{}{
		"user_id": userID,
		"name":    req.Name,
		"code":    req.Code,
		"action":  "CREATE_COURSE",
	})

	now := time.Now()
	course := &domain.Course{
		UserID:      userID,
		Name:        req.Name,
		Code:        req.Code,
		Instructor:  req.Instructor,
		Credits:     req.Credits,
		SemesterID:  &req.SemesterID,
		Type:        req.Type,
		Color:       req.Color,
		SyllabusURL: req.SyllabusURL,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	created, err := s.repo.Create(ctx, course)
	if err != nil {
		s.logger.Error("failed to create course", err, map[string]interface{}{
			"user_id": userID,
			"name":    req.Name,
			"action":  "CREATE_COURSE_FAILED",
		})
		return nil, err
	}

	s.logger.Info("course created", map[string]interface{}{
		"course_id": created.ID,
		"user_id":   userID,
		"action":    "CREATE_COURSE",
	})

	response := dto.ToCourseResponse(created)
	if s.broadcaster != nil {
		excludeCID := utils.GetConnectionIDFromContext(ctx)
		s.broadcaster.Publish(userID, excludeCID, notification.EventCourseCreated, map[string]interface{}{
			"course_id": created.ID,
			"course":    response,
		})
		s.logger.Info("WebSocket event published", map[string]interface{}{
			"event_type":  notification.EventCourseCreated,
			"user_id":     userID,
			"entity_id":   created.ID,
			"exclude_cid": excludeCID,
			"action":      "WS_EVENT_PUBLISHED",
		})
	}

	return response, nil
}

func (s *courseService) GetByID(ctx context.Context, id, userID int) (*dto.CourseResponse, error) {
	course, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}
	if course.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Load components and schedules
	components, err := s.repo.GetComponents(ctx, id)
	if err != nil {
		s.logger.Error("failed to load components", err, map[string]interface{}{
			"course_id": id,
		})
		components = []*domain.Component{}
	}
	course.Components = components

	schedules, err := s.repo.GetSchedules(ctx, id)
	if err != nil {
		s.logger.Error("failed to load schedules", err, map[string]interface{}{
			"course_id": id,
		})
		schedules = []*domain.Schedule{}
	}
	course.Schedules = schedules

	response := dto.ToCourseResponse(course)
	s.enrichWithSemester(ctx, response)
	return response, nil
}

// enrichWithSemester populates semester information in course response
func (s *courseService) enrichWithSemester(ctx context.Context, response *dto.CourseResponse) {
	if response.SemesterID == nil {
		return
	}

	semester, err := s.semesterRepo.GetByID(ctx, *response.SemesterID)
	if err != nil || semester == nil {
		return
	}

	response.Semester = &dto.SemesterInfo{
		ID:        semester.ID,
		Name:      semester.Name,
		StartDate: semester.StartDate.Format("2006-01-02"),
		EndDate:   semester.EndDate.Format("2006-01-02"),
		IsCurrent: semester.IsCurrent,
	}
}

func (s *courseService) GetAll(ctx context.Context, userID int) ([]*dto.CourseResponse, error) {
	courses, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Load components and schedules for all courses
	for _, course := range courses {
		components, err := s.repo.GetComponents(ctx, course.ID)
		if err != nil {
			s.logger.Error("failed to load components", err, map[string]interface{}{
				"course_id": course.ID,
			})
			components = []*domain.Component{}
		}
		course.Components = components

		schedules, err := s.repo.GetSchedules(ctx, course.ID)
		if err != nil {
			s.logger.Error("failed to load schedules", err, map[string]interface{}{
				"course_id": course.ID,
			})
			schedules = []*domain.Schedule{}
		}
		course.Schedules = schedules
	}

	responses := dto.ToCourseResponseList(courses)
	for _, response := range responses {
		s.enrichWithSemester(ctx, response)
	}
	return responses, nil
}

func (s *courseService) GetActive(ctx context.Context, userID int) ([]*dto.CourseResponse, error) {
	courses, err := s.repo.GetActiveCourses(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Load components and schedules for all active courses
	for _, course := range courses {
		components, err := s.repo.GetComponents(ctx, course.ID)
		if err != nil {
			s.logger.Error("failed to load components", err, map[string]interface{}{
				"course_id": course.ID,
			})
			components = []*domain.Component{}
		}
		course.Components = components

		schedules, err := s.repo.GetSchedules(ctx, course.ID)
		if err != nil {
			s.logger.Error("failed to load schedules", err, map[string]interface{}{
				"course_id": course.ID,
			})
			schedules = []*domain.Schedule{}
		}
		course.Schedules = schedules
	}

	responses := dto.ToCourseResponseList(courses)
	for _, response := range responses {
		s.enrichWithSemester(ctx, response)
	}
	return responses, nil
}

func (s *courseService) Update(ctx context.Context, id int, req *dto.UpdateCourseRequest, userID int) (*dto.CourseResponse, error) {
	s.logger.Info("Updating course", map[string]interface{}{
		"user_id":   userID,
		"course_id": id,
		"action":    "UPDATE_COURSE",
	})

	course, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}
	if course.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if req.Name != nil {
		course.Name = *req.Name
	}
	if req.Code != nil {
		course.Code = *req.Code
	}
	if req.Instructor != nil {
		course.Instructor = *req.Instructor
	}
	if req.Credits != nil {
		course.Credits = *req.Credits
	}
	if req.SemesterID != nil {
		course.SemesterID = req.SemesterID
	}
	if req.Type != nil {
		course.Type = *req.Type
	}
	if req.Color != nil {
		course.Color = *req.Color
	}
	if req.SyllabusURL != nil {
		course.SyllabusURL = *req.SyllabusURL
	}
	if req.FinalGrade != nil {
		course.FinalGrade = *req.FinalGrade
	}
	if req.IsActive != nil {
		course.IsActive = *req.IsActive
	}

	course.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, course); err != nil {
		s.logger.Error("failed to update course", err, map[string]interface{}{
			"course_id": id,
			"user_id":   userID,
			"action":    "UPDATE_COURSE_FAILED",
		})
		return nil, err
	}

	s.logger.Info("course updated", map[string]interface{}{
		"course_id": id,
		"user_id":   userID,
		"action":    "UPDATE_COURSE",
	})

	response := dto.ToCourseResponse(course)
	if s.broadcaster != nil {
		excludeCID := utils.GetConnectionIDFromContext(ctx)
		s.broadcaster.Publish(userID, excludeCID, notification.EventCourseUpdated, map[string]interface{}{
			"course_id": id,
			"course":    response,
		})
		s.logger.Info("WebSocket event published", map[string]interface{}{
			"event_type":  notification.EventCourseUpdated,
			"user_id":     userID,
			"entity_id":   id,
			"exclude_cid": excludeCID,
			"action":      "WS_EVENT_PUBLISHED",
		})
	}

	return response, nil
}

func (s *courseService) Delete(ctx context.Context, id, userID int) error {
	s.logger.Info("Deleting course", map[string]interface{}{
		"user_id":   userID,
		"course_id": id,
		"action":    "DELETE_COURSE",
	})

	course, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.New("course not found")
	}
	if course.UserID != userID {
		return errors.New("unauthorized")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete course", err, map[string]interface{}{
			"course_id": id,
			"user_id":   userID,
			"action":    "DELETE_COURSE_FAILED",
		})
		return err
	}

	s.logger.Info("course deleted", map[string]interface{}{
		"course_id": id,
		"user_id":   userID,
		"action":    "DELETE_COURSE",
	})

	if s.broadcaster != nil {
		excludeCID := utils.GetConnectionIDFromContext(ctx)
		s.broadcaster.Publish(userID, excludeCID, notification.EventCourseDeleted, map[string]interface{}{
			"course_id": id,
			"name":      course.Name,
		})
		s.logger.Info("WebSocket event published", map[string]interface{}{
			"event_type":  notification.EventCourseDeleted,
			"user_id":     userID,
			"entity_id":   id,
			"exclude_cid": excludeCID,
			"action":      "WS_EVENT_PUBLISHED",
		})
	}

	return nil
}

func (s *courseService) CreateComponent(ctx context.Context, req *dto.CreateComponentRequest, userID int) (*dto.ComponentResponse, error) {
	s.logger.Info("Creating course component", map[string]interface{}{
		"user_id":   userID,
		"course_id": req.CourseID,
		"name":      req.Name,
		"type":      req.Type,
		"action":    "CREATE_COMPONENT",
	})

	// Verify course belongs to user
	course, err := s.repo.GetByID(ctx, req.CourseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}
	if course.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	now := time.Now()
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, errors.New("invalid due_date format, use YYYY-MM-DD")
		}
		dueDate = &parsed
	}

	var weight float64
	if req.Weight != nil {
		weight = *req.Weight
	}

	var maxScore float64
	if req.MaxScore != nil {
		maxScore = *req.MaxScore
	}

	component := &domain.Component{
		CourseID:      req.CourseID,
		Type:          req.Type,
		Name:          req.Name,
		Weight:        weight,
		MaxScore:      maxScore,
		AchievedScore: req.AchievedScore,
		DueDate:       dueDate,
		IsCompleted:   false,
		Notes:         req.Notes,
		DisplayOrder:  req.DisplayOrder,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	created, err := s.repo.CreateComponent(ctx, component)
	if err != nil {
		s.logger.Error("failed to create component", err, map[string]interface{}{
			"user_id":   userID,
			"course_id": req.CourseID,
			"action":    "CREATE_COMPONENT_FAILED",
		})
		return nil, err
	}

	s.logger.Info("component created", map[string]interface{}{
		"component_id": created.ID,
		"course_id":    req.CourseID,
		"user_id":      userID,
		"action":       "CREATE_COMPONENT",
	})

	response := dto.ToComponentResponse(created)
	if s.broadcaster != nil {
		excludeCID := utils.GetConnectionIDFromContext(ctx)
		s.broadcaster.Publish(userID, excludeCID, notification.EventComponentCreated, map[string]interface{}{
			"component_id": created.ID,
			"course_id":    req.CourseID,
			"component":    response,
		})
		s.logger.Info("WebSocket event published", map[string]interface{}{
			"event_type":  notification.EventComponentCreated,
			"user_id":     userID,
			"entity_id":   created.ID,
			"exclude_cid": excludeCID,
			"action":      "WS_EVENT_PUBLISHED",
		})
	}

	return response, nil
}

func (s *courseService) GetComponents(ctx context.Context, courseID, userID int) ([]*dto.ComponentResponse, error) {
	// Verify course belongs to user
	course, err := s.repo.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}
	if course.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	components, err := s.repo.GetComponents(ctx, courseID)
	if err != nil {
		return nil, err
	}

	return dto.ToComponentResponseList(components), nil
}

func (s *courseService) UpdateComponent(ctx context.Context, id int, req *dto.UpdateComponentRequest, userID int) (*dto.ComponentResponse, error) {
	s.logger.Info("Updating course component", map[string]interface{}{
		"user_id":      userID,
		"component_id": id,
		"action":       "UPDATE_COMPONENT",
	})

	component, err := s.repo.GetComponentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if component == nil {
		return nil, errors.New("component not found")
	}

	// Verify course belongs to user
	course, err := s.repo.GetByID(ctx, component.CourseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}
	if course.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Update fields
	if req.Type != nil {
		component.Type = *req.Type
	}
	if req.Name != nil {
		component.Name = *req.Name
	}
	if req.Weight != nil {
		component.Weight = *req.Weight
	}
	if req.MaxScore != nil {
		component.MaxScore = *req.MaxScore
	}
	if req.AchievedScore != nil {
		component.AchievedScore = req.AchievedScore
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			component.DueDate = nil
		} else {
			parsed, err := time.Parse("2006-01-02", *req.DueDate)
			if err != nil {
				return nil, errors.New("invalid due_date format, use YYYY-MM-DD")
			}
			component.DueDate = &parsed
		}
	}
	if req.CompletionDate != nil {
		if *req.CompletionDate == "" {
			component.CompletionDate = nil
		} else {
			parsed, err := time.Parse("2006-01-02", *req.CompletionDate)
			if err != nil {
				return nil, errors.New("invalid completion_date format, use YYYY-MM-DD")
			}
			component.CompletionDate = &parsed
		}
	}
	if req.IsCompleted != nil {
		component.IsCompleted = *req.IsCompleted
	}
	if req.Notes != nil {
		component.Notes = *req.Notes
	}
	if req.DisplayOrder != nil {
		component.DisplayOrder = *req.DisplayOrder
	}

	component.UpdatedAt = time.Now()

	if err := s.repo.UpdateComponent(ctx, component); err != nil {
		s.logger.Error("failed to update component", err, map[string]interface{}{
			"component_id": id,
			"user_id":      userID,
			"action":       "UPDATE_COMPONENT_FAILED",
		})
		return nil, err
	}

	s.logger.Info("component updated", map[string]interface{}{
		"component_id": id,
		"user_id":      userID,
		"action":       "UPDATE_COMPONENT",
	})

	response := dto.ToComponentResponse(component)

	// Check if grade was updated - broadcast grade change
	if req.AchievedScore != nil && s.broadcaster != nil {
		// Recalculate course grade
		course, _ := s.repo.GetByID(ctx, component.CourseID)
		if course != nil {
			components, _ := s.repo.GetComponents(ctx, component.CourseID)
			var totalWeight, weightedScore float64
			for _, comp := range components {
				if comp.AchievedScore != nil && comp.IsCompleted {
					totalWeight += comp.Weight
					weightedScore += (*comp.AchievedScore / comp.MaxScore) * comp.Weight
				}
			}
			var newGrade float64
			if totalWeight > 0 {
				newGrade = (weightedScore / totalWeight) * 100
			}

			excludeCID := utils.GetConnectionIDFromContext(ctx)
			s.broadcaster.Publish(userID, excludeCID, notification.EventComponentGraded, map[string]interface{}{
				"component_id": id,
				"course_id":    component.CourseID,
				"component":    response,
				"new_grade":    newGrade,
			})
			s.logger.Info("WebSocket event published", map[string]interface{}{
				"event_type":  notification.EventComponentGraded,
				"user_id":     userID,
				"entity_id":   id,
				"exclude_cid": excludeCID,
				"action":      "WS_EVENT_PUBLISHED",
			})
		}
	}

	if s.broadcaster != nil {
		excludeCID := utils.GetConnectionIDFromContext(ctx)
		s.broadcaster.Publish(userID, excludeCID, notification.EventComponentUpdated, map[string]interface{}{
			"component_id": id,
			"course_id":    component.CourseID,
			"component":    response,
		})
		s.logger.Info("WebSocket event published", map[string]interface{}{
			"event_type":  notification.EventComponentUpdated,
			"user_id":     userID,
			"entity_id":   id,
			"exclude_cid": excludeCID,
			"action":      "WS_EVENT_PUBLISHED",
		})
	}

	return response, nil
}

func (s *courseService) DeleteComponent(ctx context.Context, id, userID int) error {
	s.logger.Info("Deleting course component", map[string]interface{}{
		"user_id":      userID,
		"component_id": id,
		"action":       "DELETE_COMPONENT",
	})

	component, err := s.repo.GetComponentByID(ctx, id)
	if err != nil {
		return err
	}
	if component == nil {
		return errors.New("component not found")
	}

	// Verify course belongs to user
	course, err := s.repo.GetByID(ctx, component.CourseID)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.New("course not found")
	}
	if course.UserID != userID {
		return errors.New("unauthorized")
	}

	if err := s.repo.DeleteComponent(ctx, id); err != nil {
		s.logger.Error("failed to delete component", err, map[string]interface{}{
			"component_id": id,
			"user_id":      userID,
			"action":       "DELETE_COMPONENT_FAILED",
		})
		return err
	}

	s.logger.Info("component deleted", map[string]interface{}{
		"component_id": id,
		"user_id":      userID,
		"action":       "DELETE_COMPONENT",
	})

	return nil
}

func (s *courseService) CreateSchedule(ctx context.Context, req *dto.CreateScheduleRequest, userID int) (*dto.ScheduleResponse, error) {
	s.logger.Info("Creating course schedule", map[string]interface{}{
		"user_id":   userID,
		"course_id": req.CourseID,
		"day":       req.DayOfWeek,
		"action":    "CREATE_SCHEDULE",
	})

	// Verify course belongs to user
	course, err := s.repo.GetByID(ctx, req.CourseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}
	if course.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Validate and normalize time format (HH:MM)
	startTime := normalizeTime(req.StartTime)
	endTime := normalizeTime(req.EndTime)

	schedule := &domain.Schedule{
		CourseID:             req.CourseID,
		DayOfWeek:            req.DayOfWeek,
		StartTime:            startTime,
		EndTime:              endTime,
		Location:             req.Location,
		NotificationsEnabled: req.NotificationsEnabled,
		NotificationType:     req.NotificationType,
		ReminderTime:         req.ReminderTime,
		CreatedAt:            time.Now(),
	}

	created, err := s.repo.CreateSchedule(ctx, schedule)
	if err != nil {
		s.logger.Error("failed to create schedule", err, map[string]interface{}{
			"user_id":   userID,
			"course_id": req.CourseID,
			"action":    "CREATE_SCHEDULE_FAILED",
		})
		return nil, err
	}

	s.logger.Info("schedule created", map[string]interface{}{
		"schedule_id": created.ID,
		"course_id":   req.CourseID,
		"user_id":     userID,
		"action":      "CREATE_SCHEDULE",
	})

	// Sync to Google Calendar if requested
	if req.SyncToCalendar && s.calendarService != nil {
		s.logger.Info("syncing schedule to Google Calendar", map[string]interface{}{
			"schedule_id": created.ID,
			"course_id":   req.CourseID,
			"user_id":     userID,
			"action":      "SYNC_TO_CALENDAR",
		})

		// Get course details for event title
		courseDetails, err := s.repo.GetByID(ctx, req.CourseID)
		if err == nil && courseDetails != nil {
			eventTitle := fmt.Sprintf("%s - %s", courseDetails.Code, courseDetails.Name)
			eventDescription := fmt.Sprintf("Location: %s\nInstructor: %s", req.Location, courseDetails.Instructor)

			// Parse start time
			// Parse start time
			startTimeParts := strings.Split(startTime, ":")
			if len(startTimeParts) == 2 {
				hour, _ := time.Parse("15", startTimeParts[0])
				minute, _ := time.Parse("04", startTimeParts[1])

				endTimeParts := strings.Split(endTime, ":")
				if len(endTimeParts) == 2 {
					endHour, _ := time.Parse("15", endTimeParts[0])
					endMinute, _ := time.Parse("04", endTimeParts[1])

					// Calculate first event date and RRULE
					var firstEventDate time.Time
					var recurrence []string

					// Try to get semester info
					var semesterEndDate time.Time
					var baseDate time.Time
					if courseDetails.SemesterID != nil {
						semester, err := s.semesterRepo.GetByID(ctx, *courseDetails.SemesterID)
						if err == nil && semester != nil {
							semesterEndDate = semester.EndDate

							// Determine start date (max of Now or Semester Start)
							// We want to generate events starting from the semester start, but if it's already passed,
							// starting from Now avoids cluttering the past.
							// However, for completeness, let's start from valid semester start if active,
							// or Now if not specified.
							// User requested "start end semesterdan al".
							// Let's use Semester.StartDate if available, otherwise Now.
							baseDate = time.Now()
							if !semester.StartDate.IsZero() {
								baseDate = semester.StartDate
							}

							firstEventDate = calculateNextDayOfWeek(baseDate, req.DayOfWeek)
						} else {
							// Fallback if semester not found
							firstEventDate = calculateNextDayOfWeek(time.Now(), req.DayOfWeek)
						}
					} else {
						// Fallback if no semester ID
						firstEventDate = calculateNextDayOfWeek(time.Now(), req.DayOfWeek)
					}

					// Get user timezone
					var loc *time.Location
					user, err := s.userRepo.GetByID(ctx, userID)
					if err == nil && user != nil && user.Timezone != "" {
						loc, err = time.LoadLocation(user.Timezone)
						if err != nil {
							s.logger.Error("failed to load user timezone", err, map[string]interface{}{"timezone": user.Timezone})
							loc = time.UTC
						}
					} else {
						loc = time.UTC
					}

					// Construct timestamps using user's location
					firstEventStart := time.Date(
						firstEventDate.Year(), firstEventDate.Month(), firstEventDate.Day(),
						hour.Hour(), minute.Minute(), 0, 0, loc,
					)
					firstEventEnd := time.Date(
						firstEventDate.Year(), firstEventDate.Month(), firstEventDate.Day(),
						endHour.Hour(), endMinute.Minute(), 0, 0, loc,
					)

					// Construct RRULE if semester end date is available
					if !semesterEndDate.IsZero() {
						// RRULE:FREQ=WEEKLY;UNTIL=20240601T235959Z;BYDAY=MO
						// Translate DayOfWeek to RRULE format
						rruleDay := mapDayToRRULE(req.DayOfWeek)
						if rruleDay != "" {
							// Format UNTIL date in UTC
							untilStr := semesterEndDate.UTC().Format("20060102T150405Z")
							recurrence = []string{
								fmt.Sprintf("RRULE:FREQ=WEEKLY;UNTIL=%s;BYDAY=%s", untilStr, rruleDay),
							}
							s.logger.Info("RRULE generated", map[string]interface{}{
								"rrule":             recurrence[0],
								"until":             untilStr,
								"base_date":         baseDate.Format(time.RFC3339),
								"first_event_date":  firstEventDate.Format(time.RFC3339),
								"semester_end_date": semesterEndDate.Format(time.RFC3339),
							})
						}
					} else {
						s.logger.Info("RRULE skipped: semester end date is zero", nil)
					}

					// Create the event
					err = s.calendarService.CreateTimedEvent(
						ctx, userID, created.ID, "course_schedule",
						eventTitle, eventDescription,
						firstEventStart, firstEventEnd,
						recurrence,
						req.NotificationsEnabled, req.NotificationType, req.ReminderTime,
					)
					if err != nil {
						s.logger.Error("failed to sync schedule to Google Calendar", err, map[string]interface{}{
							"schedule_id": created.ID,
							"user_id":     userID,
						})
						// Don't fail the whole operation if calendar sync fails
					} else {
						s.logger.Info("schedule synced to Google Calendar", map[string]interface{}{
							"schedule_id": created.ID,
							"user_id":     userID,
							"rrule":       recurrence,
						})
					}
				} else {
					s.logger.Error("invalid end time format for calendar sync", nil, map[string]interface{}{
						"end_time": endTime,
					})
				}
			} else {
				s.logger.Error("invalid start time format for calendar sync", nil, map[string]interface{}{
					"start_time": startTime,
				})
			}
		}
	}

	response := dto.ToScheduleResponse(created)
	if s.broadcaster != nil {
		excludeCID := utils.GetConnectionIDFromContext(ctx)
		s.broadcaster.Publish(userID, excludeCID, notification.EventScheduleCreated, map[string]interface{}{
			"schedule_id": created.ID,
			"course_id":   req.CourseID,
			"schedule":    response,
		})
		s.logger.Info("WebSocket event published", map[string]interface{}{
			"event_type":  notification.EventScheduleCreated,
			"user_id":     userID,
			"entity_id":   created.ID,
			"exclude_cid": excludeCID,
			"action":      "WS_EVENT_PUBLISHED",
		})
	}

	return response, nil
}

func (s *courseService) GetSchedules(ctx context.Context, courseID, userID int) ([]*dto.ScheduleResponse, error) {
	// Verify course belongs to user
	course, err := s.repo.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}
	if course.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	schedules, err := s.repo.GetSchedules(ctx, courseID)
	if err != nil {
		return nil, err
	}

	return dto.ToScheduleResponseList(schedules), nil
}

func (s *courseService) UpdateSchedule(ctx context.Context, id int, req *dto.UpdateScheduleRequest, userID int) (*dto.ScheduleResponse, error) {
	s.logger.Info("Updating course schedule", map[string]interface{}{
		"user_id":     userID,
		"schedule_id": id,
		"action":      "UPDATE_SCHEDULE",
	})

	schedule, err := s.repo.GetScheduleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if schedule == nil {
		return nil, errors.New("schedule not found")
	}

	// Verify course belongs to user
	course, err := s.repo.GetByID(ctx, schedule.CourseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, errors.New("course not found")
	}
	if course.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Update fields
	if req.DayOfWeek != nil {
		schedule.DayOfWeek = *req.DayOfWeek
	}
	if req.StartTime != nil {
		normalizedStart := normalizeTime(*req.StartTime)
		s.logger.Info("Normalizing start time", map[string]interface{}{
			"original":   *req.StartTime,
			"normalized": normalizedStart,
		})
		schedule.StartTime = normalizedStart
	}
	if req.EndTime != nil {
		normalizedEnd := normalizeTime(*req.EndTime)
		s.logger.Info("Normalizing end time", map[string]interface{}{
			"original":   *req.EndTime,
			"normalized": normalizedEnd,
		})
		schedule.EndTime = normalizedEnd
	}
	if req.Location != nil {
		schedule.Location = *req.Location
	}
	if req.NotificationsEnabled != nil {
		schedule.NotificationsEnabled = *req.NotificationsEnabled
	}
	if req.NotificationType != nil {
		schedule.NotificationType = *req.NotificationType
	}
	if req.ReminderTime != nil {
		schedule.ReminderTime = *req.ReminderTime
	}

	if err := s.repo.UpdateSchedule(ctx, schedule); err != nil {
		s.logger.Error("failed to update schedule", err, map[string]interface{}{
			"schedule_id": id,
			"user_id":     userID,
			"action":      "UPDATE_SCHEDULE_FAILED",
		})
		return nil, err
	}

	s.logger.Info("schedule updated", map[string]interface{}{
		"schedule_id": id,
		"user_id":     userID,
		"action":      "UPDATE_SCHEDULE",
	})

	// Sync update to Google Calendar
	if s.calendarService != nil {
		// Calculate new dates/RRULE similar to CreateSchedule
		// We need course details and semester info again
		// Ideally refactor date calc logic, but for now duplicate/inline to fit constraint

		var recurrence []string
		var firstEventDate time.Time

		// Parse times
		startTimeParts := strings.Split(schedule.StartTime, ":")
		endTimeParts := strings.Split(schedule.EndTime, ":")

		if len(startTimeParts) == 2 && len(endTimeParts) == 2 {
			hour, _ := time.Parse("15", startTimeParts[0])
			minute, _ := time.Parse("04", startTimeParts[1])
			endHour, _ := time.Parse("15", endTimeParts[0])
			endMinute, _ := time.Parse("04", endTimeParts[1])

			// Get semester info
			var semesterEndDate time.Time
			if course.SemesterID != nil {
				semester, err := s.semesterRepo.GetByID(ctx, *course.SemesterID)
				if err == nil && semester != nil {
					semesterEndDate = semester.EndDate

					baseDate := time.Now()
					if !semester.StartDate.IsZero() {
						baseDate = semester.StartDate
					}

					firstEventDate = calculateNextDayOfWeek(baseDate, schedule.DayOfWeek)
				} else {
					firstEventDate = calculateNextDayOfWeek(time.Now(), schedule.DayOfWeek)
				}
			} else {
				firstEventDate = calculateNextDayOfWeek(time.Now(), schedule.DayOfWeek)
			}

			// Get user timezone
			var loc *time.Location
			user, err := s.userRepo.GetByID(ctx, userID)
			if err == nil && user != nil && user.Timezone != "" {
				loc, err = time.LoadLocation(user.Timezone)
				if err != nil {
					s.logger.Error("failed to load user timezone", err, map[string]interface{}{"timezone": user.Timezone})
					loc = time.UTC
				}
			} else {
				loc = time.UTC
			}

			firstEventStart := time.Date(
				firstEventDate.Year(), firstEventDate.Month(), firstEventDate.Day(),
				hour.Hour(), minute.Minute(), 0, 0, loc,
			)
			firstEventEnd := time.Date(
				firstEventDate.Year(), firstEventDate.Month(), firstEventDate.Day(),
				endHour.Hour(), endMinute.Minute(), 0, 0, loc,
			)

			if !semesterEndDate.IsZero() {
				rruleDay := mapDayToRRULE(schedule.DayOfWeek)
				if rruleDay != "" {
					untilStr := semesterEndDate.UTC().Format("20060102T150405Z")
					recurrence = []string{
						fmt.Sprintf("RRULE:FREQ=WEEKLY;UNTIL=%s;BYDAY=%s", untilStr, rruleDay),
					}
				}
			}

			eventTitle := fmt.Sprintf("%s - %s", course.Code, course.Name)
			eventDescription := fmt.Sprintf("Location: %s\nInstructor: %s", schedule.Location, course.Instructor)

			err = s.calendarService.UpdateTimedEvent(
				ctx, userID, id, "course_schedule",
				eventTitle, eventDescription,
				firstEventStart, firstEventEnd,
				recurrence,
				schedule.NotificationsEnabled, schedule.NotificationType, schedule.ReminderTime,
			)
			if err != nil {
				s.logger.Error("failed to update Google Calendar event", err, nil)
			} else {
				s.logger.Info("Google Calendar event updated", nil)
			}
		}
	}

	response := dto.ToScheduleResponse(schedule)
	if s.broadcaster != nil {
		excludeCID := utils.GetConnectionIDFromContext(ctx)
		s.broadcaster.Publish(userID, excludeCID, notification.EventScheduleUpdated, map[string]interface{}{
			"schedule_id": id,
			"course_id":   schedule.CourseID,
			"schedule":    response,
		})
		s.logger.Info("WebSocket event published", map[string]interface{}{
			"event_type":  notification.EventScheduleUpdated,
			"user_id":     userID,
			"entity_id":   id,
			"exclude_cid": excludeCID,
			"action":      "WS_EVENT_PUBLISHED",
		})
	}

	return response, nil
}

func (s *courseService) DeleteSchedule(ctx context.Context, id, userID int) error {
	s.logger.Info("Deleting course schedule", map[string]interface{}{
		"user_id":     userID,
		"schedule_id": id,
		"action":      "DELETE_SCHEDULE",
	})

	schedule, err := s.repo.GetScheduleByID(ctx, id)
	if err != nil {
		return err
	}
	if schedule == nil {
		return errors.New("schedule not found")
	}

	// Verify course belongs to user
	course, err := s.repo.GetByID(ctx, schedule.CourseID)
	if err != nil {
		return err
	}
	if course == nil {
		return errors.New("course not found")
	}
	if course.UserID != userID {
		return errors.New("unauthorized")
	}

	if err := s.repo.DeleteSchedule(ctx, id); err != nil {
		s.logger.Error("failed to delete schedule", err, map[string]interface{}{
			"schedule_id": id,
			"user_id":     userID,
			"action":      "DELETE_SCHEDULE_FAILED",
		})
		return err
	}

	// Sync deletion to Google Calendar
	if s.calendarService != nil {
		if err := s.calendarService.DeleteEvent(ctx, userID, id, "course_schedule"); err != nil {
			s.logger.Error("failed to delete schedule from Google Calendar", err, nil)
		}
	}

	s.logger.Info("schedule deleted", map[string]interface{}{
		"schedule_id": id,
		"user_id":     userID,
		"action":      "DELETE_SCHEDULE",
	})

	if s.broadcaster != nil {
		excludeCID := utils.GetConnectionIDFromContext(ctx)
		s.broadcaster.Publish(userID, excludeCID, notification.EventScheduleDeleted, map[string]interface{}{
			"schedule_id": id,
			"course_id":   schedule.CourseID,
			"day_of_week": schedule.DayOfWeek,
		})
		s.logger.Info("WebSocket event published", map[string]interface{}{
			"event_type":  notification.EventScheduleDeleted,
			"user_id":     userID,
			"entity_id":   id,
			"exclude_cid": excludeCID,
			"action":      "WS_EVENT_PUBLISHED",
		})
	}

	return nil
}

func calculateNextDayOfWeek(t time.Time, day string) time.Time {
	days := map[string]time.Weekday{
		"Sunday":    time.Sunday,
		"Monday":    time.Monday,
		"Tuesday":   time.Tuesday,
		"Wednesday": time.Wednesday,
		"Thursday":  time.Thursday,
		"Friday":    time.Friday,
		"Saturday":  time.Saturday,
	}

	target, ok := days[day]
	if !ok {
		return t // Fallback
	}

	// Calculate days until target
	daysUntil := int(target - t.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	// If we want "next" occurrence literally, <= 0 is correct.
	// However, if today is Monday and schedule is Monday, do we want today or next week?
	// "start end semesterdan al" -> if semester started today, and course is today, it should include today.
	// So daysUntil should be >= 0.
	// If daysUntil == 0, it wraps to 7 if we use <= 0 logic from some snippets.
	// But my previous snippet used `if daysUntil < 0 { daysUntil += 7 }`.
	// This allows 0 (today). This is what we want.
	// Let's revert to `< 0`.

	daysUntil = int(target - t.Weekday())
	if daysUntil < 0 {
		daysUntil += 7
	}

	return t.AddDate(0, 0, daysUntil)
}

func mapDayToRRULE(day string) string {
	mapping := map[string]string{
		"Sunday":    "SU",
		"Monday":    "MO",
		"Tuesday":   "TU",
		"Wednesday": "WE",
		"Thursday":  "TH",
		"Friday":    "FR",
		"Saturday":  "SA",
	}
	return mapping[day]
}
