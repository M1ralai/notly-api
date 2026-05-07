package http

import (
	"net/http"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/modules/dashboard/dto"
	"github.com/M1ralai/notly-api/internal/modules/dashboard/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	service *service.DashboardService
}

func NewHandler(service *service.DashboardService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Register route
	router.HandleFunc("/dashboard", h.GetDashboardData).Methods("GET")
}

// @Summary Get Dashboard Data
// @Tags Dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.DashboardResponse
// @Router /api/dashboard [get]
func (h *Handler) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := utils.GetUserIDFromContext(ctx)

	resp, err := h.service.GetDashboardData(ctx, userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Failed to fetch dashboard data", err.Error())
		return
	}

	var _ *dto.DashboardResponse = resp

	utils.WriteJson(w, resp, http.StatusOK, "Dashboard data fetched successfully")
}
