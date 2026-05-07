package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/common/validation"
	"github.com/M1ralai/notly-api/internal/modules/lifearea/dto"
	"github.com/M1ralai/notly-api/internal/modules/lifearea/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	service service.LifeAreaService
}

func NewHandler(service service.LifeAreaService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/life-areas", h.GetAll).Methods("GET")
	router.HandleFunc("/life-areas", h.Create).Methods("POST")
	router.HandleFunc("/life-areas/{id}", h.GetByID).Methods("GET")
	router.HandleFunc("/life-areas/{id}", h.Update).Methods("PUT", "PATCH")
	router.HandleFunc("/life-areas/{id}", h.Delete).Methods("DELETE")
}

func (h *Handler) getUserID(r *http.Request) int {
	return utils.GetUserIDFromContext(r.Context())
}

// @Summary Create
// @Tags Lifearea
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateLifeAreaRequest true "Create Life Area"
// @Success 201 {object} dto.LifeAreaResponse
// @Router /api/life-areas [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateLifeAreaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)
	lifeArea, err := h.service.Create(r.Context(), &req, userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Hayat alanı oluşturulamadı", err.Error())
		return
	}

	utils.WriteJson(w, lifeArea, http.StatusCreated, "Hayat alanı oluşturuldu")
}

// @Summary GetByID
// @Tags Lifearea
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} dto.LifeAreaResponse
// @Router /api/life-areas/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	lifeArea, err := h.service.GetByID(r.Context(), id, userID)
	if err != nil {
		if err.Error() == "life area not found" {
			utils.ReturnError(w, "NOT_FOUND", "Hayat alanı bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Hayat alanı getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, lifeArea, http.StatusOK, "Hayat alanı getirildi")
}

// @Summary GetAll
// @Tags Lifearea
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.LifeAreaResponse
// @Router /api/life-areas [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	areas, err := h.service.GetByUserID(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Hayat alanları getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, areas, http.StatusOK, "Hayat alanları getirildi")
}

// @Summary Update
// @Tags Lifearea
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body dto.UpdateLifeAreaRequest true "Update Life Area"
// @Success 200 {object} dto.LifeAreaResponse
// @Router /api/life-areas/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	var req dto.UpdateLifeAreaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)
	lifeArea, err := h.service.Update(r.Context(), id, &req, userID)
	if err != nil {
		if err.Error() == "life area not found" {
			utils.ReturnError(w, "NOT_FOUND", "Hayat alanı bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Hayat alanı güncellenemedi", err.Error())
		return
	}

	utils.WriteJson(w, lifeArea, http.StatusOK, "Hayat alanı güncellendi")
}

// @Summary Delete
// @Tags Lifearea
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} map[string]string
// @Router /api/life-areas/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		if err.Error() == "life area not found" {
			utils.ReturnError(w, "NOT_FOUND", "Hayat alanı bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Hayat alanı silinemedi", err.Error())
		return
	}

	utils.WriteJson(w, nil, http.StatusOK, "Hayat alanı silindi")
}
