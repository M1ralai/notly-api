package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/common/validation"
	"github.com/M1ralai/notly-api/internal/modules/note/dto"
	"github.com/M1ralai/notly-api/internal/modules/note/service"
	subscriptionService "github.com/M1ralai/notly-api/internal/modules/subscription/service"
)

const maxUploadSize = 32 << 20 // 32 MB

// Handler holds the NoteService dependency.
type Handler struct {
	svc service.NoteService
}

// NewHandler constructs the HTTP handler.
func NewHandler(svc service.NoteService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes wires all note routes.
// publicRouter must be the *root* mux router (no auth middleware).
// protectedAPI must be the /api subrouter with AuthMiddleware applied.
//
// Call this function TWICE from server.go:
//
//	noteHandler.RegisterPublicRoutes(router)
//	noteHandler.RegisterRoutes(api)
func (h *Handler) RegisterRoutes(api *mux.Router) {
	api.HandleFunc("/notes", h.CreateNote).Methods("POST")
	api.HandleFunc("/notes", h.GetAllNotes).Methods("GET")
	api.HandleFunc("/notes/{id}", h.GetNoteByID).Methods("GET")
	api.HandleFunc("/notes/{id}", h.UpdateNote).Methods("PUT")
	api.HandleFunc("/notes/{id}", h.DeleteNote).Methods("DELETE")

	// Attachments
	api.HandleFunc("/notes/{id}/attachments", h.UploadAttachment).Methods("POST")
	api.HandleFunc("/notes/{id}/attachments/{attId}", h.DeleteAttachment).Methods("DELETE")

	// Sharing
	api.HandleFunc("/notes/{id}/share/public", h.SetPublic).Methods("PUT")

	// Collaborators
	api.HandleFunc("/notes/{id}/collaborators", h.AddCollaborator).Methods("POST")
	api.HandleFunc("/notes/{id}/collaborators", h.GetCollaborators).Methods("GET")
	api.HandleFunc("/notes/{id}/collaborators/{userId}", h.RemoveCollaborator).Methods("DELETE")
}

// RegisterPublicRoutes registers auth-free routes on the root router.
func (h *Handler) RegisterPublicRoutes(router *mux.Router) {
	router.HandleFunc("/api/shared/notes/{token}", h.GetSharedNote).Methods("GET")
	router.HandleFunc("/api/shared/notes/{token}/attachments/{attId}", h.DownloadSharedAttachment).Methods("GET")
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func decodeAndValidate(r *http.Request, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return err
	}
	return validation.Get().Struct(dst)
}

func pathParamInt(r *http.Request, key string) (int, error) {
	val := mux.Vars(r)[key]
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %s", key, val)
	}
	return n, nil
}

func handleServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, subscriptionService.ErrPremiumRequired) {
		utils.ReturnError(w, "PREMIUM_REQUIRED", "Notly Pro required", err.Error())
		return
	}

	switch err.Error() {
	case "unauthorized":
		utils.ReturnError(w, "FORBIDDEN", "Access denied", err.Error())
	case "not found":
		utils.ReturnError(w, "NOT_FOUND", "Resource not found", err.Error())
	case "storage unavailable":
		utils.ReturnError(w, "INTERNAL_ERROR", "File storage is not available", err.Error())
	default:
		utils.ReturnError(w, "INTERNAL_ERROR", "An error occurred", err.Error())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Core CRUD handlers
// ─────────────────────────────────────────────────────────────────────────────

// CreateNote godoc
// @Summary Create a note
// @Tags Notes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateNoteRequest true "Create Note Payload"
// @Success 201 {object} dto.NoteOwnerResponse
// @Router /api/notes [post]
func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	var req dto.CreateNoteRequest
	if err := decodeAndValidate(r, &req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Invalid request", err.Error())
		return
	}
	resp, err := h.svc.Create(r.Context(), &req, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, resp, http.StatusCreated, "Note created")
}

// GetAllNotes godoc
// @Summary List all notes for authenticated user
// @Tags Notes
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.NoteOwnerResponse
// @Router /api/notes [get]
func (h *Handler) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	resp, err := h.svc.GetAll(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, resp, http.StatusOK, "Notes fetched")
}

// GetNoteByID godoc
// @Summary Get a note by ID
// @Tags Notes
// @Security BearerAuth
// @Produce json
// @Param id path int true "Note ID"
// @Success 200 {object} dto.NoteOwnerResponse
// @Router /api/notes/{id} [get]
func (h *Handler) GetNoteByID(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	id, err := pathParamInt(r, "id")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid note ID", err.Error())
		return
	}
	resp, err := h.svc.GetByID(r.Context(), id, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, resp, http.StatusOK, "Note fetched")
}

// UpdateNote godoc
// @Summary Update a note
// @Tags Notes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Note ID"
// @Param request body dto.UpdateNoteRequest true "Update Note Payload"
// @Success 200 {object} dto.NoteOwnerResponse
// @Router /api/notes/{id} [put]
func (h *Handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	id, err := pathParamInt(r, "id")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid note ID", err.Error())
		return
	}
	var req dto.UpdateNoteRequest
	if err := decodeAndValidate(r, &req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Invalid request", err.Error())
		return
	}
	resp, err := h.svc.Update(r.Context(), id, &req, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, resp, http.StatusOK, "Note updated")
}

// DeleteNote godoc
// @Summary Delete a note
// @Tags Notes
// @Security BearerAuth
// @Produce json
// @Param id path int true "Note ID"
// @Success 200 {object} map[string]string
// @Router /api/notes/{id} [delete]
func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	id, err := pathParamInt(r, "id")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid note ID", err.Error())
		return
	}
	if err := h.svc.Delete(r.Context(), id, userID); err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, nil, http.StatusOK, "Note deleted")
}

// ─────────────────────────────────────────────────────────────────────────────
// Attachment handlers (Kısım 03)
// ─────────────────────────────────────────────────────────────────────────────

// UploadAttachment godoc
// @Summary Upload a media file (image/PDF) to a note
// @Tags Notes
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Note ID"
// @Param file formData file true "File to upload"
// @Success 201 {object} dto.AttachmentResponse
// @Router /api/notes/{id}/attachments [post]
func (h *Handler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	noteID, err := pathParamInt(r, "id")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid note ID", err.Error())
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "File too large or malformed", err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Missing file field", err.Error())
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	resp, err := h.svc.UploadAttachment(r.Context(), noteID, userID, file, header.Filename, contentType, header.Size)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, resp, http.StatusCreated, "Attachment uploaded")
}

// DeleteAttachment godoc
// @Summary Delete an attachment from a note
// @Tags Notes
// @Security BearerAuth
// @Produce json
// @Param id path int true "Note ID"
// @Param attId path int true "Attachment ID"
// @Success 200 {object} map[string]string
// @Router /api/notes/{id}/attachments/{attId} [delete]
func (h *Handler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	attID, err := pathParamInt(r, "attId")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid attachment ID", err.Error())
		return
	}
	if err := h.svc.DeleteAttachment(r.Context(), attID, userID); err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, nil, http.StatusOK, "Attachment deleted")
}

// ─────────────────────────────────────────────────────────────────────────────
// Sharing handlers (Kısım 04)
// ─────────────────────────────────────────────────────────────────────────────

// SetPublic godoc
// @Summary Toggle public sharing for a note
// @Tags Notes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Note ID"
// @Param request body dto.SetPublicRequest true "Public flag"
// @Success 200 {object} dto.ShareTokenResponse
// @Router /api/notes/{id}/share/public [put]
func (h *Handler) SetPublic(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	noteID, err := pathParamInt(r, "id")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid note ID", err.Error())
		return
	}
	var req dto.SetPublicRequest
	if err := decodeAndValidate(r, &req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Invalid request", err.Error())
		return
	}
	resp, err := h.svc.SetPublic(r.Context(), noteID, userID, req.IsPublic)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, resp, http.StatusOK, "Sharing updated")
}

// GetSharedNote godoc
// @Summary View a publicly shared note (no auth required)
// @Tags Notes (Public)
// @Produce json
// @Param token path string true "Share token"
// @Success 200 {object} dto.SharedNoteMinimalResponse
// @Router /api/shared/notes/{token} [get]
func (h *Handler) GetSharedNote(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	if token == "" {
		utils.ReturnError(w, "BAD_REQUEST", "Missing share token", "token is empty")
		return
	}
	resp, err := h.svc.GetByShareToken(r.Context(), token)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, resp, http.StatusOK, "Shared note fetched")
}

// DownloadSharedAttachment godoc
// @Summary Download an attachment from a publicly shared note
// @Tags Notes (Public)
// @Produce octet-stream
// @Param token path string true "Share token"
// @Param attId path int true "Attachment ID"
// @Success 200 {file} file
// @Router /api/shared/notes/{token}/attachments/{attId} [get]
func (h *Handler) DownloadSharedAttachment(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	if token == "" {
		utils.ReturnError(w, "BAD_REQUEST", "Missing share token", "token is empty")
		return
	}

	attID, err := pathParamInt(r, "attId")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid attachment ID", err.Error())
		return
	}

	download, err := h.svc.DownloadSharedAttachment(r.Context(), token, attID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	defer download.Body.Close()

	fileName := download.FileName
	if fileName == "" {
		fileName = "attachment"
	}

	w.Header().Set("Content-Type", download.ContentType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": fileName}))
	w.Header().Set("Cache-Control", "no-store")
	if download.Size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(download.Size, 10))
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, download.Body)
}

// ─────────────────────────────────────────────────────────────────────────────
// Collaborator handlers (Kısım 04)
// ─────────────────────────────────────────────────────────────────────────────

// AddCollaborator godoc
// @Summary Add a collaborator to a note
// @Tags Notes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Note ID"
// @Param request body dto.AddCollaboratorRequest true "Collaborator payload"
// @Success 201 {object} dto.CollaboratorResponse
// @Router /api/notes/{id}/collaborators [post]
func (h *Handler) AddCollaborator(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	noteID, err := pathParamInt(r, "id")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid note ID", err.Error())
		return
	}
	var req dto.AddCollaboratorRequest
	if err := decodeAndValidate(r, &req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Invalid request", err.Error())
		return
	}
	resp, err := h.svc.AddCollaborator(r.Context(), noteID, userID, &req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, resp, http.StatusCreated, "Collaborator added")
}

// GetCollaborators godoc
// @Summary List collaborators for a note
// @Tags Notes
// @Security BearerAuth
// @Produce json
// @Param id path int true "Note ID"
// @Success 200 {array} dto.CollaboratorResponse
// @Router /api/notes/{id}/collaborators [get]
func (h *Handler) GetCollaborators(w http.ResponseWriter, r *http.Request) {
	userID := utils.GetUserIDFromContext(r.Context())
	noteID, err := pathParamInt(r, "id")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid note ID", err.Error())
		return
	}
	resp, err := h.svc.GetCollaborators(r.Context(), noteID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, resp, http.StatusOK, "Collaborators fetched")
}

// RemoveCollaborator godoc
// @Summary Remove a collaborator from a note
// @Tags Notes
// @Security BearerAuth
// @Produce json
// @Param id path int true "Note ID"
// @Param userId path int true "Collaborator User ID"
// @Success 200 {object} map[string]string
// @Router /api/notes/{id}/collaborators/{userId} [delete]
func (h *Handler) RemoveCollaborator(w http.ResponseWriter, r *http.Request) {
	ownerID := utils.GetUserIDFromContext(r.Context())
	noteID, err := pathParamInt(r, "id")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid note ID", err.Error())
		return
	}
	collaboratorID, err := pathParamInt(r, "userId")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Invalid user ID", err.Error())
		return
	}
	if err := h.svc.RemoveCollaborator(r.Context(), noteID, ownerID, collaboratorID); err != nil {
		handleServiceError(w, err)
		return
	}
	utils.WriteJson(w, nil, http.StatusOK, "Collaborator removed")
}
