package http

import (
	"net/http"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/modules/subscription/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	service        service.Service
	webhookHandler *WebhookHandler
}

func NewHandler(service service.Service) *Handler {
	return &Handler{
		service:        service,
		webhookHandler: NewWebhookHandler(service),
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/subscriptions/me", h.GetMe).Methods("GET")
	router.HandleFunc("/subscriptions/sync", h.Sync).Methods("POST")
}

func (h *Handler) RegisterPublicRoutes(router *mux.Router) {
	router.HandleFunc("/webhooks/revenuecat", h.webhookHandler.HandleRevenueCat).Methods("POST")
	router.HandleFunc("/api/webhooks/revenuecat", h.webhookHandler.HandleRevenueCat).Methods("POST")
}

// @Summary Get current user's subscription status
// @Tags Subscription
// @Security BearerAuth
// @Produce json
// @Success 200 {object} service.PremiumStatus
// @Router /api/subscriptions/me [get]
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	status, err := h.service.GetPremiumStatus(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Abonelik durumu okunamadı", err.Error())
		return
	}
	utils.WriteJson(w, status, http.StatusOK, "Abonelik durumu getirildi")
}

// @Summary Sync/refresh current user's subscription status
// @Tags Subscription
// @Security BearerAuth
// @Produce json
// @Success 200 {object} service.PremiumStatus
// @Router /api/subscriptions/sync [post]
func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	status, err := h.service.GetPremiumStatus(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Abonelik durumu senkronize edilemedi", err.Error())
		return
	}
	utils.WriteJson(w, status, http.StatusOK, "Abonelik durumu güncellendi")
}
