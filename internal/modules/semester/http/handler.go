package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/common/validation"
	"github.com/M1ralai/notly-api/internal/modules/semester/dto"
	"github.com/M1ralai/notly-api/internal/modules/semester/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	service service.SemesterService
}

func NewHandler(service service.SemesterService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/semesters", h.GetAll).Methods("GET")
	router.HandleFunc("/semesters", h.Create).Methods("POST")
	router.HandleFunc("/semesters/{id}", h.GetByID).Methods("GET")
	router.HandleFunc("/semesters/{id}", h.Update).Methods("PUT", "PATCH")
	router.HandleFunc("/semesters/{id}", h.Delete).Methods("DELETE")
	router.HandleFunc("/semesters/current", h.GetCurrent).Methods("GET")
}

func (h *Handler) getUserID(r *http.Request) int {
	return utils.GetUserIDFromContext(r.Context())
}

// @Summary Create
// @Tags Semester
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateSemesterRequest true "Create Semester"
// @Success 201 {object} dto.SemesterResponse
// @Router /api/semesters [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSemesterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)
	semester, err := h.service.Create(r.Context(), &req, userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Dönem oluşturulamadı", err.Error())
		return
	}

	utils.WriteJson(w, semester, http.StatusCreated, "Dönem oluşturuldu")
}

// @Summary GetByID
// @Tags Semester
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} dto.SemesterResponse
// @Router /api/semesters/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	semester, err := h.service.GetByID(r.Context(), id, userID)
	if err != nil {
		if err.Error() == "semester not found" {
			utils.ReturnError(w, "NOT_FOUND", "Dönem bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Dönem getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, semester, http.StatusOK, "Dönem getirildi")
}

// @Summary GetAll
// @Tags Semester
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.SemesterResponse
// @Router /api/semesters [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	semesters, err := h.service.GetAll(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Dönemler getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, semesters, http.StatusOK, "Dönemler getirildi")
}

// @Summary GetCurrent
// @Tags Semester
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.SemesterResponse
// @Router /api/semesters/current [get]
func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	semester, err := h.service.GetCurrent(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Mevcut dönem getirilemedi", err.Error())
		return
	}

	if semester == nil {
		utils.WriteJson(w, nil, http.StatusOK, "Mevcut dönem bulunamadı")
		return
	}

	utils.WriteJson(w, semester, http.StatusOK, "Mevcut dönem getirildi")
}

// @Summary Update
// @Tags Semester
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body dto.UpdateSemesterRequest true "Update Semester"
// @Success 200 {object} dto.SemesterResponse
// @Router /api/semesters/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	var req dto.UpdateSemesterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)
	semester, err := h.service.Update(r.Context(), id, &req, userID)
	if err != nil {
		if err.Error() == "semester not found" {
			utils.ReturnError(w, "NOT_FOUND", "Dönem bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Dönem güncellenemedi", err.Error())
		return
	}

	utils.WriteJson(w, semester, http.StatusOK, "Dönem güncellendi")
}

// @Summary Delete
// @Tags Semester
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} map[string]string
// @Router /api/semesters/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		if err.Error() == "semester not found" {
			utils.ReturnError(w, "NOT_FOUND", "Dönem bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Dönem silinemedi", err.Error())
		return
	}

	utils.WriteJson(w, nil, http.StatusOK, "Dönem silindi")
}
