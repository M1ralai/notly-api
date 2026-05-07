package http

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/M1ralai/notly-api/internal/common/utils"
	notifService "github.com/M1ralai/notly-api/internal/modules/notification/service"
	"github.com/M1ralai/notly-api/internal/modules/sync/dto"
	"github.com/M1ralai/notly-api/internal/modules/sync/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	syncService service.SyncService
	broadcaster *notifService.Broadcaster
}

func NewHandler(syncService service.SyncService, broadcaster *notifService.Broadcaster) *Handler {
	return &Handler{
		syncService: syncService,
		broadcaster: broadcaster,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/sync/delta", h.GetDelta).Methods("GET")
	router.HandleFunc("/sync/signal", h.SyncSignal).Methods("POST")
}

// @Summary SyncSignal
// @Tags Sync
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.SyncSignalRequest true "Sync Signal"
// @Success 200 {object} map[string]string
// @Router /api/sync/signal [post]
func (h *Handler) SyncSignal(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	if userID == 0 {
		utils.ReturnError(w, "UNAUTHORIZED", "User not found in context", "Authentication required")
		return
	}

	// Try multiple cases for the connection ID header
	excludeCID := r.Header.Get("X-Connection-ID")
	if excludeCID == "" {
		excludeCID = r.Header.Get("X-Connection-Id")
	}
	if excludeCID == "" {
		excludeCID = r.Header.Get("x-connection-id")
	}

	var req dto.SyncSignalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid request body", err.Error())
		return
	}

	// Simple relay of the signal to other tabs
	h.broadcaster.SyncSignal(userID, excludeCID, req.Type, req.Payload)

	utils.WriteJson(w, nil, http.StatusOK, "Signal broadcasted")
}

// @Summary GetDelta
// @Tags Sync
// @Security BearerAuth
// @Produce json
// @Param since query string false "Since Time (RFC3339)"
// @Success 200 {object} dto.DeltaSyncResponse
// @Router /api/sync/delta [get]
func (h *Handler) GetDelta(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized access"})
		return
	}

	sinceStr := r.URL.Query().Get("since")
	var since time.Time

	if sinceStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			log.Printf("Sync GetDelta Invalid Time Format: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid 'since' format. Expected ISO8601 (RFC3339)."})
			return
		}
		since = parsedTime
	} else {
		// If since is not provided, we should return all records (essentially a full sync snapshot via delta)
		// We set 'since' to a very old date
		since = time.Time{}
	}

	// Calculate and get Delta
	delta, err := h.syncService.GetDelta(r.Context(), userID, since)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to compute sync delta"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(delta)
}
