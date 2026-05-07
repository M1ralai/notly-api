package http

import (
	"encoding/json"
	"net/http"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/common/validation"
	"github.com/M1ralai/notly-api/internal/modules/calendar/dto"
	"github.com/M1ralai/notly-api/internal/modules/calendar/service"
	"github.com/gorilla/mux"
)

type Handler struct{ service service.CalendarService }

func NewHandler(service service.CalendarService) *Handler { return &Handler{service: service} }

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/calendar/google/auth-url", h.GetGoogleAuthURL).Methods("GET")
	router.HandleFunc("/calendar/google/callback", h.HandleGoogleCallback).Methods("POST")
	router.HandleFunc("/calendar/google/disconnect", h.DisconnectGoogle).Methods("POST")
	router.HandleFunc("/calendar/google/sync", h.SyncGoogle).Methods("POST")
	router.HandleFunc("/calendar/status", h.GetSyncStatus).Methods("GET")
	router.HandleFunc("/calendar/integrations", h.GetIntegrations).Methods("GET")
	router.HandleFunc("/calendar/sync/queue", h.QueueSync).Methods("POST")
}

func (h *Handler) getUserID(r *http.Request) int {
	return utils.GetUserIDFromContext(r.Context())
}

// @Summary GetGoogleAuthURL
// @Tags Calendar
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.AuthURLResponse
// @Router /api/calendar/google/auth-url [get]
func (h *Handler) GetGoogleAuthURL(w http.ResponseWriter, r *http.Request) {
	authURL, err := h.service.GetGoogleAuthURL(r.Context(), h.getUserID(r))
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Google OAuth yapılandırılmamış", err.Error())
		return
	}
	utils.WriteJson(w, dto.AuthURLResponse{AuthURL: authURL}, http.StatusOK, "Yetkilendirme URL'i oluşturuldu")
}

// @Summary HandleGoogleCallback
// @Tags Calendar
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.GoogleCallbackRequest true "Google Callback Request"
// @Success 200 {object} dto.IntegrationResponse
// @Router /api/calendar/google/callback [post]
func (h *Handler) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	var req dto.GoogleCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek", err.Error())
		return
	}

	if req.Code == "" {
		utils.ReturnError(w, "BAD_REQUEST", "Kod gerekli", "code parametresi eksik")
		return
	}

	integration, err := h.service.HandleGoogleCallback(r.Context(), h.getUserID(r), req.Code)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Google bağlantısı başarısız", err.Error())
		return
	}
	utils.WriteJson(w, integration, http.StatusOK, "Google Calendar bağlandı")
}

// @Summary DisconnectGoogle
// @Tags Calendar
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/calendar/google/disconnect [post]
func (h *Handler) DisconnectGoogle(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DisconnectGoogle(r.Context(), h.getUserID(r)); err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Bağlantı kesilemedi", err.Error())
		return
	}
	utils.WriteJson(w, nil, http.StatusOK, "Google Calendar bağlantısı kesildi")
}

// @Summary SyncGoogle
// @Tags Calendar
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/calendar/google/sync [post]
func (h *Handler) SyncGoogle(w http.ResponseWriter, r *http.Request) {
	if err := h.service.SyncGoogle(r.Context(), h.getUserID(r)); err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Senkronizasyon başarısız", err.Error())
		return
	}
	utils.WriteJson(w, nil, http.StatusOK, "Google Calendar senkronize edildi")
}

// @Summary GetSyncStatus
// @Tags Calendar
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.SyncStatusResponse
// @Router /api/calendar/status [get]
func (h *Handler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.GetSyncStatus(r.Context(), h.getUserID(r))
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Durum alınamadı", err.Error())
		return
	}
	utils.WriteJson(w, status, http.StatusOK, "Senkronizasyon durumu")
}

// @Summary GetIntegrations
// @Tags Calendar
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.IntegrationResponse
// @Router /api/calendar/integrations [get]
func (h *Handler) GetIntegrations(w http.ResponseWriter, r *http.Request) {
	integrations, err := h.service.GetIntegrations(r.Context(), h.getUserID(r))
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Entegrasyonlar alınamadı", err.Error())
		return
	}
	utils.WriteJson(w, integrations, http.StatusOK, "Takvim entegrasyonları")
}

// @Summary QueueSync
// @Tags Calendar
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.QueueSyncRequest true "Queue Sync Request"
// @Success 200 {object} map[string]string
// @Router /api/calendar/sync/queue [post]
func (h *Handler) QueueSync(w http.ResponseWriter, r *http.Request) {
	var req dto.QueueSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek", err.Error())
		return
	}
	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}
	if err := h.service.QueueSync(r.Context(), h.getUserID(r), req.EventID, req.Action); err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Senkronizasyon kuyruğa eklenemedi", err.Error())
		return
	}
	utils.WriteJson(w, nil, http.StatusOK, "Senkronizasyon kuyruğa eklendi")
}
