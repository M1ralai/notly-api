package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/common/validation"
	"github.com/M1ralai/notly-api/internal/infrastructure/jobs"
	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
	calendarService "github.com/M1ralai/notly-api/internal/modules/calendar/service"
	jobimpl "github.com/M1ralai/notly-api/internal/modules/job/jobs"
	notifService "github.com/M1ralai/notly-api/internal/modules/notification/service"
	"github.com/M1ralai/notly-api/internal/modules/task/dto"
	"github.com/M1ralai/notly-api/internal/modules/task/repository"
	"github.com/M1ralai/notly-api/internal/modules/task/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	service         service.TaskService
	jobPool         *jobs.WorkerPool
	repo            repository.TaskRepository
	broadcaster     *notifService.Broadcaster
	calendarService calendarService.CalendarService
	logger          *logger.ZapLogger
}

func NewHandler(
	service service.TaskService,
	jobPool *jobs.WorkerPool,
	repo repository.TaskRepository,
	broadcaster *notifService.Broadcaster,
	calendarSvc calendarService.CalendarService,
	logger *logger.ZapLogger,
) *Handler {
	return &Handler{
		service:         service,
		jobPool:         jobPool,
		repo:            repo,
		broadcaster:     broadcaster,
		calendarService: calendarSvc,
		logger:          logger,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Specific routes must be registered BEFORE parameterized routes
	router.HandleFunc("/tasks/stats", h.GetStats).Methods("GET")
	router.HandleFunc("/tasks/parent", h.GetParentTasks).Methods("GET")
	router.HandleFunc("/tasks/{id}/subtasks", h.GetSubtasks).Methods("GET")
	router.HandleFunc("/tasks/{id}/complete", h.CompleteSubtask).Methods("POST")
	router.HandleFunc("/tasks/{id}/uncomplete", h.UncompleteTask).Methods("POST")

	// General routes
	router.HandleFunc("/tasks", h.GetAll).Methods("GET")
	router.HandleFunc("/tasks", h.Create).Methods("POST")

	// Parameterized routes (must be last)
	router.HandleFunc("/tasks/{id}", h.GetByID).Methods("GET")
	router.HandleFunc("/tasks/{id}", h.Update).Methods("PUT", "PATCH")
	router.HandleFunc("/tasks/{id}", h.Delete).Methods("DELETE")
}

func (h *Handler) getUserID(r *http.Request) int {
	return utils.GetUserIDFromContext(r.Context())
}

// @Summary Create Task
// @Tags Task
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateTaskRequest true "Create Task"
// @Success 201 {object} dto.TaskResponse
// @Router /api/tasks [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)
	task, err := h.service.Create(r.Context(), &req, userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Görev oluşturulamadı", err.Error())
		return
	}

	utils.WriteJson(w, task, http.StatusCreated, "Görev oluşturuldu")
}

// @Summary Get Task By ID
// @Tags Task
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} dto.TaskResponse
// @Router /api/tasks/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	task, err := h.service.GetByID(r.Context(), id, userID)
	if err != nil {
		if err.Error() == "task not found" {
			utils.ReturnError(w, "NOT_FOUND", "Görev bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Görev getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, task, http.StatusOK, "Görev getirildi")
}

// @Summary Get All Tasks
// @Tags Task
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.TaskResponse
// @Router /api/tasks [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	tasks, err := h.service.GetAll(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Görevler getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, tasks, http.StatusOK, "Görevler getirildi")
}

// @Summary Get Parent Tasks
// @Tags Task
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.TaskResponse
// @Router /api/tasks/parent [get]
func (h *Handler) GetParentTasks(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	tasks, err := h.service.GetParentTasks(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Ana görevler getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, tasks, http.StatusOK, "Ana görevler getirildi")
}

// @Summary Get Subtasks
// @Tags Task
// @Security BearerAuth
// @Produce json
// @Param id path string true "id"
// @Success 200 {array} dto.TaskResponse
// @Router /api/tasks/{id}/subtasks [get]
func (h *Handler) GetSubtasks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	subtasks, err := h.service.GetSubtasks(r.Context(), parentID, userID)
	if err != nil {
		if err.Error() == "parent task not found" {
			utils.ReturnError(w, "NOT_FOUND", "Ana görev bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Alt görevler getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, subtasks, http.StatusOK, "Alt görevler getirildi")
}

// @Summary Update Task
// @Tags Task
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param request body dto.UpdateTaskRequest true "Update Task"
// @Success 200 {object} dto.TaskResponse
// @Router /api/tasks/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	var req dto.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)

	// Submit job to pool asynchronously
	if h.jobPool != nil {
		updateJob := jobimpl.NewTaskUpdateJob(h.logger, h.repo, h.broadcaster, id, userID, &req)
		if err := h.jobPool.SubmitAsync(updateJob); err != nil {
			h.logger.Error("Failed to submit task update job", err, map[string]interface{}{
				"task_id": id,
				"user_id": userID,
				"action":  "TASK_UPDATE_JOB_SUBMIT_FAILED",
			})
			// Fallback to synchronous update if job submission fails
			task, err := h.service.Update(r.Context(), id, &req, userID)
			if err != nil {
				utils.ReturnError(w, "INTERNAL_ERROR", "Görev güncellenemedi", err.Error())
				return
			}
			utils.WriteJson(w, task, http.StatusOK, "Görev güncellendi")
			return
		}

		// Return immediately - job will process in background
		utils.WriteJson(w, map[string]interface{}{
			"message": "Task update job submitted",
			"task_id": id,
		}, http.StatusAccepted, "Görev güncellemesi işleme alındı")
		return
	}

	// Fallback to synchronous update if job pool is not available
	task, err := h.service.Update(r.Context(), id, &req, userID)
	if err != nil {
		if err.Error() == "task not found" {
			utils.ReturnError(w, "NOT_FOUND", "Görev bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Görev güncellenemedi", err.Error())
		return
	}

	utils.WriteJson(w, task, http.StatusOK, "Görev güncellendi")
}

// @Summary Delete Task
// @Tags Task
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} map[string]string
// @Router /api/tasks/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		if err.Error() == "task not found" {
			utils.ReturnError(w, "NOT_FOUND", "Görev bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Görev silinemedi", err.Error())
		return
	}

	utils.WriteJson(w, nil, http.StatusOK, "Görev silindi")
}

// @Summary Complete Subtask
// @Tags Task
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} map[string]string
// @Router /api/tasks/{id}/complete [post]
func (h *Handler) CompleteSubtask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	if err := h.service.CompleteSubtask(r.Context(), id, userID); err != nil {
		if err.Error() == "subtask not found" {
			utils.ReturnError(w, "NOT_FOUND", "Alt görev bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Alt görev tamamlanamadı", err.Error())
		return
	}

	// Sync to Google Calendar (non-blocking)
	if h.calendarService != nil {
		task, _ := h.service.GetByID(r.Context(), id, userID)
		if task != nil {
			go func() {
				if err := h.calendarService.MarkDone(r.Context(), userID, id, "task", task.Title, time.Now()); err != nil {
					h.logger.Error("Failed to sync task to calendar", err, map[string]interface{}{
						"task_id": id, "user_id": userID,
					})
				}
			}()
		}
	}

	utils.WriteJson(w, nil, http.StatusOK, "Alt görev tamamlandı")
}

// UncompleteTask marks a task as not completed and removes from Google Calendar
// @Summary Uncomplete Task
// @Tags Task
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} map[string]string
// @Router /api/tasks/{id}/uncomplete [post]
func (h *Handler) UncompleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)

	// Update task status to not completed
	isCompleted := false
	req := &dto.UpdateTaskRequest{
		IsCompleted: &isCompleted,
	}
	_, err = h.service.Update(r.Context(), id, req, userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Görev tamamlama geri alınamadı", err.Error())
		return
	}

	// Remove from Google Calendar (non-blocking)
	if h.calendarService != nil {
		go func() {
			if err := h.calendarService.MarkUndone(r.Context(), userID, id, "task", time.Now()); err != nil {
				h.logger.Error("Failed to remove task from calendar", err, map[string]interface{}{
					"task_id": id, "user_id": userID,
				})
			}
		}()
	}

	utils.WriteJson(w, nil, http.StatusOK, "Görev tamamlama geri alındı")
}

// @Summary Get Task Statistics
// @Tags Task
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.TaskStatsResponse
// @Router /api/tasks/stats [get]
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	stats, err := h.service.GetStats(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "İstatistikler getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, stats, http.StatusOK, "İstatistikler getirildi")
}
