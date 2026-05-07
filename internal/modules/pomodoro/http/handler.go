package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/common/validation"
	"github.com/M1ralai/notly-api/internal/modules/pomodoro/dto"
	"github.com/M1ralai/notly-api/internal/modules/pomodoro/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/pomodoros", h.CreateSession).Methods("POST")
	router.HandleFunc("/pomodoro/sessions", h.GetSessions).Methods("GET")
	router.HandleFunc("/pomodoros/settings", h.GetSettings).Methods("GET")
	router.HandleFunc("/pomodoros/settings", h.UpdateSettings).Methods("PUT")
	router.HandleFunc("/courses/{courseId:[0-9]+}/sessions", h.GetCourseSessions).Methods("GET")
}

// @Summary CreateSession
// @Tags Pomodoro
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreatePomodoroSessionRequest true "Create Session"
// @Success 201 {object} dto.PomodoroSessionResponse
// @Router /api/pomodoros [post]
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	if userID == 0 {
		utils.ReturnError(w, "UNAUTHORIZED", "User not found in context", "Authentication required")
		return
	}

	var req dto.CreatePomodoroSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid request body", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Validation failed", validation.FormatErr(err))
		return
	}

	session, err := h.service.CreateSession(r.Context(), userID, req)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Failed to create session", err.Error())
		return
	}

	utils.WriteJson(w, session, http.StatusCreated, "Session created successfully")
}

// @Summary GetSessions
// @Tags Pomodoro
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.PomodoroSessionResponse
// @Router /api/pomodoro/sessions [get]
func (h *Handler) GetSessions(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	if userID == 0 {
		utils.ReturnError(w, "UNAUTHORIZED", "User not found in context", "Authentication required")
		return
	}

	sessions, err := h.service.GetSessions(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Failed to get sessions", err.Error())
		return
	}

	utils.WriteJson(w, sessions, http.StatusOK, "Sessions retrieved successfully")
}

// @Summary GetCourseSessions
// @Tags Pomodoro
// @Security BearerAuth
// @Produce json
// @Param courseId path int true "ID"
// @Success 200 {array} dto.PomodoroSessionResponse
// @Router /api/courses/{courseId}/sessions [get]
func (h *Handler) GetCourseSessions(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	if userID == 0 {
		utils.ReturnError(w, "UNAUTHORIZED", "User not found in context", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	courseID, err := strconv.Atoi(vars["courseId"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid course ID", "Course ID must be an integer")
		return
	}

	sessions, err := h.service.GetCourseSessions(r.Context(), userID, courseID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Failed to get course sessions", err.Error())
		return
	}

	utils.WriteJson(w, sessions, http.StatusOK, "Course sessions retrieved successfully")
}

// @Summary GetSettings
// @Tags Pomodoro
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.PomodoroSettingsResponse
// @Router /api/pomodoros/settings [get]
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	if userID == 0 {
		utils.ReturnError(w, "UNAUTHORIZED", "User not found in context", "Authentication required")
		return
	}

	settings, err := h.service.GetSettings(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Failed to get settings", err.Error())
		return
	}

	utils.WriteJson(w, settings, http.StatusOK, "Settings retrieved successfully")
}

// @Summary UpdateSettings
// @Tags Pomodoro
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdatePomodoroSettingsRequest true "Update Settings"
// @Success 200 {object} dto.PomodoroSettingsResponse
// @Router /api/pomodoros/settings [put]
func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	if userID == 0 {
		utils.ReturnError(w, "UNAUTHORIZED", "User not found in context", "Authentication required")
		return
	}

	var req dto.UpdatePomodoroSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid request body", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Validation failed", validation.FormatErr(err))
		return
	}

	settings, err := h.service.UpdateSettings(r.Context(), userID, req)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Failed to update settings", err.Error())
		return
	}

	utils.WriteJson(w, settings, http.StatusOK, "Settings updated successfully")
}
