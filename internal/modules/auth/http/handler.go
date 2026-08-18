package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/common/validation"
	"github.com/M1ralai/notly-api/internal/modules/auth/dto"
	"github.com/M1ralai/notly-api/internal/modules/auth/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	service service.AuthService
}

func NewHandler(service service.AuthService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/auth/login", h.Login).Methods("POST")
	router.HandleFunc("/auth/register", h.Register).Methods("POST")
	router.HandleFunc("/auth/verify", h.VerifyEmail).Methods("POST")
	router.HandleFunc("/auth/resend-code", h.ResendCode).Methods("POST")
	router.HandleFunc("/auth/refresh", h.RefreshToken).Methods("POST")
	router.HandleFunc("/auth/logout", h.Logout).Methods("POST")
}

// @Summary Login User
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Request"
// @Success 200 {object} dto.AuthResponse
// @Router /api/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	response, err := h.service.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrBotVerificationFailed) {
			utils.ReturnError(w, "FORBIDDEN", "Bot doğrulaması başarısız", err.Error())
			return
		}
		if err.Error() == "invalid email or password" {
			utils.ReturnError(w, "UNAUTHORIZED", "Geçersiz e-posta veya şifre", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Giriş yapılamadı", err.Error())
		return
	}

	utils.WriteJson(w, response, http.StatusOK, "Giriş başarılı")
}

// @Summary Register
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register Request"
// @Success 201 {object} map[string]string
// @Router /api/auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	response, err := h.service.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrBotVerificationFailed) {
			utils.ReturnError(w, "FORBIDDEN", "Bot doğrulaması başarısız", err.Error())
			return
		}
		if err.Error() == "email already exists" || err.Error() == "email already registered" {
			utils.ReturnError(w, "BAD_REQUEST", "Bu e-posta adresi zaten kullanımda", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Kayıt oluşturulamadı", err.Error())
		return
	}

	utils.WriteJson(w, response, http.StatusCreated, "Kayıt başarılı")
}

// @Summary Refresh Token
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} dto.AuthResponse
// @Router /api/auth/refresh [post]
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	response, err := h.service.RefreshToken(r.Context(), &req)
	if err != nil {
		if err.Error() == "invalid refresh token" || err.Error() == "refresh token expired" || err.Error() == "user not found" {
			utils.ReturnError(w, "UNAUTHORIZED", "Oturum süresi dolmuş, lütfen tekrar giriş yapın", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Token yenilenemedi", err.Error())
		return
	}

	utils.WriteJson(w, response, http.StatusOK, "Token başarıyla yenilendi")
}

// @Summary Logout
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.LogoutRequest true "Logout Request"
// @Success 200 {object} map[string]string
// @Router /api/auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	if err := h.service.Logout(r.Context(), &req); err != nil {
		// Log error but return success to client
		utils.ReturnError(w, "INTERNAL_ERROR", "Çıkış yapılırken bir hata oluştu", err.Error())
		return
	}

	utils.WriteJson(w, map[string]string{"message": "Başarıyla çıkış yapıldı"}, http.StatusOK, "Başarıyla çıkış yapıldı")
}
