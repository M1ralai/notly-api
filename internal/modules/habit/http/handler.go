package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/common/validation"
	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
	calendarService "github.com/M1ralai/notly-api/internal/modules/calendar/service"
	"github.com/M1ralai/notly-api/internal/modules/habit/dto"
	"github.com/M1ralai/notly-api/internal/modules/habit/repository"
	"github.com/M1ralai/notly-api/internal/modules/habit/service"
	notifService "github.com/M1ralai/notly-api/internal/modules/notification/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	service         service.HabitService
	repo            repository.HabitRepository
	broadcaster     *notifService.Broadcaster
	calendarService calendarService.CalendarService
	logger          *logger.ZapLogger
}

func NewHandler(
	service service.HabitService,
	repo repository.HabitRepository,
	broadcaster *notifService.Broadcaster,
	calendarSvc calendarService.CalendarService,
	logger *logger.ZapLogger,
) *Handler {
	return &Handler{
		service:         service,
		repo:            repo,
		broadcaster:     broadcaster,
		calendarService: calendarSvc,
		logger:          logger,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/habits", h.GetAll).Methods("GET")
	router.HandleFunc("/habits/active", h.GetActive).Methods("GET")
	router.HandleFunc("/habits", h.Create).Methods("POST")
	router.HandleFunc("/habits/{id}", h.GetByID).Methods("GET")
	router.HandleFunc("/habits/{id}", h.Update).Methods("PUT", "PATCH")
	router.HandleFunc("/habits/{id}", h.Delete).Methods("DELETE")
	router.HandleFunc("/habits/{id}/log", h.LogHabit).Methods("POST")
	router.HandleFunc("/habits/{id}/complete", h.Complete).Methods("POST")
	router.HandleFunc("/habits/{id}/skip", h.Skip).Methods("POST")
	router.HandleFunc("/habits/{id}/uncomplete", h.Uncomplete).Methods("POST")
}

func (h *Handler) getUserID(r *http.Request) int {
	return utils.GetUserIDFromContext(r.Context())
}

// @Summary Create
// @Tags Habit
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateHabitRequest true "Create Habit"
// @Success 201 {object} dto.HabitResponse
// @Router /api/habits [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek", err.Error())
		return
	}
	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}
	habit, err := h.service.Create(r.Context(), &req, h.getUserID(r))
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Alışkanlık oluşturulamadı", err.Error())
		return
	}
	utils.WriteJson(w, habit, http.StatusCreated, "Alışkanlık oluşturuldu")
}

// @Summary GetByID
// @Tags Habit
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} dto.HabitResponse
// @Router /api/habits/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	habit, err := h.service.GetByID(r.Context(), id, h.getUserID(r))
	if err != nil {
		if err.Error() == "habit not found" {
			utils.ReturnError(w, "NOT_FOUND", "Alışkanlık bulunamadı", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Alışkanlık getirilemedi", err.Error())
		return
	}
	utils.WriteJson(w, habit, http.StatusOK, "Alışkanlık getirildi")
}

// @Summary GetAll
// @Tags Habit
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.HabitResponse
// @Router /api/habits [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	habits, err := h.service.GetAll(r.Context(), h.getUserID(r))
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Alışkanlıklar getirilemedi", err.Error())
		return
	}
	utils.WriteJson(w, habits, http.StatusOK, "Alışkanlıklar getirildi")
}

// @Summary GetActive
// @Tags Habit
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.HabitResponse
// @Router /api/habits/active [get]
func (h *Handler) GetActive(w http.ResponseWriter, r *http.Request) {
	habits, err := h.service.GetActive(r.Context(), h.getUserID(r))
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Aktif alışkanlıklar getirilemedi", err.Error())
		return
	}
	utils.WriteJson(w, habits, http.StatusOK, "Aktif alışkanlıklar getirildi")
}

// @Summary Update
// @Tags Habit
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body dto.UpdateHabitRequest true "Update Habit"
// @Success 200 {object} dto.HabitResponse
// @Router /api/habits/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var req dto.UpdateHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek", err.Error())
		return
	}
	habit, err := h.service.Update(r.Context(), id, &req, h.getUserID(r))
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Alışkanlık güncellenemedi", err.Error())
		return
	}
	utils.WriteJson(w, habit, http.StatusOK, "Alışkanlık güncellendi")
}

// @Summary Delete
// @Tags Habit
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} map[string]string
// @Router /api/habits/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.service.Delete(r.Context(), id, h.getUserID(r)); err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Alışkanlık silinemedi", err.Error())
		return
	}
	utils.WriteJson(w, nil, http.StatusOK, "Alışkanlık silindi")
}

// @Summary LogHabit
// @Tags Habit
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body dto.LogHabitRequest true "Log Habit"
// @Success 200 {object} map[string]string
// @Router /api/habits/{id}/log [post]
func (h *Handler) LogHabit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var req dto.LogHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek", err.Error())
		return
	}
	if err := h.service.LogHabit(r.Context(), id, &req, h.getUserID(r)); err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Alışkanlık kaydedilemedi", err.Error())
		return
	}
	utils.WriteJson(w, nil, http.StatusOK, "Alışkanlık kaydedildi")
}

// @Summary Skip
// @Tags Habit
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} map[string]string
// @Router /api/habits/{id}/skip [post]
func (h *Handler) Skip(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	userID := h.getUserID(r)

	if err := h.service.SkipHabit(r.Context(), id, userID); err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Alışkanlık atlanamadı", err.Error())
		return
	}
	utils.WriteJson(w, nil, http.StatusOK, "Alışkanlık atlandı")
}

// @Summary Complete
// @Tags Habit
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body dto.LogHabitRequest false "Log Habit Optional"
// @Success 200 {object} map[string]string
// @Router /api/habits/{id}/complete [post]
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var req dto.LogHabitRequest
	json.NewDecoder(r.Body).Decode(&req) // Ignore error if empty body

	userID := h.getUserID(r)

	if err := h.service.Complete(r.Context(), id, &req, userID); err != nil {
		if err.Error() == "habit already completed today" {
			utils.ReturnError(w, "BAD_REQUEST", "Bu alışkanlık bugün zaten tamamlandı", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Alışkanlık tamamlanamadı", err.Error())
		return
	}

	// Sync to Google Calendar (non-blocking, fire and forget)
	if h.calendarService != nil {
		habit, _ := h.service.GetByID(r.Context(), id, userID)
		if habit != nil {
			go func() {
				if err := h.calendarService.MarkDone(r.Context(), userID, id, "habit", habit.Title, time.Now()); err != nil {
					h.logger.Error("Failed to sync habit to calendar", err, map[string]interface{}{
						"habit_id": id, "user_id": userID,
					})
				}
			}()
		}
	}

	utils.WriteJson(w, nil, http.StatusOK, "Alışkanlık tamamlandı")
}

// Uncomplete removes a habit completion (marks as undone)
// @Summary Uncomplete
// @Tags Habit
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} map[string]string
// @Router /api/habits/{id}/uncomplete [post]
func (h *Handler) Uncomplete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	userID := h.getUserID(r)

	// Delete today's log from database
	if err := h.repo.DeleteTodayLog(r.Context(), id, userID); err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Alışkanlık tamamlama geri alınamadı", err.Error())
		return
	}

	// Remove from Google Calendar (non-blocking)
	if h.calendarService != nil {
		go func() {
			if err := h.calendarService.MarkUndone(r.Context(), userID, id, "habit", time.Now()); err != nil {
				h.logger.Error("Failed to remove habit from calendar", err, map[string]interface{}{
					"habit_id": id, "user_id": userID,
				})
			}
		}()
	}

	utils.WriteJson(w, nil, http.StatusOK, "Alışkanlık tamamlama geri alındı")
}
