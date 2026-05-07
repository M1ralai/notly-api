package http

import (
	"encoding/json"
	"net/http"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/common/validation"
	"github.com/M1ralai/notly-api/internal/modules/auth/dto"
)

// @Summary Verify Email
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.VerifyEmailRequest true "Verify Email Request"
// @Success 200 {object} dto.AuthResponse
// @Router /api/auth/verify [post]
func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	response, err := h.service.VerifyEmail(r.Context(), &req)
	if err != nil {
		if err.Error() == "invalid or expired verification code" {
			utils.ReturnError(w, "BAD_REQUEST", "Doğrulama kodu geçersiz veya süresi dolmuş", err.Error())
			return
		}
		if err.Error() == "email already verified" {
			utils.ReturnError(w, "BAD_REQUEST", "E-posta zaten doğrulanmış", err.Error())
			return
		}
		if err.Error() == "user not found" {
			utils.ReturnError(w, "NOT_FOUND", "Kullanıcı bulunamadı", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Doğrulama başarısız", err.Error())
		return
	}

	utils.WriteJson(w, response, http.StatusOK, "E-posta başarıyla doğrulandı")
}

// @Summary Resend Verification Code
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ResendCodeRequest true "Resend Code Request"
// @Success 200 {object} map[string]string
// @Router /api/auth/resend-code [post]
func (h *Handler) ResendCode(w http.ResponseWriter, r *http.Request) {
	var req dto.ResendCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	err := h.service.ResendCode(r.Context(), &req)
	if err != nil {
		if err.Error() == "email already verified" {
			utils.ReturnError(w, "BAD_REQUEST", "E-posta zaten doğrulanmış", err.Error())
			return
		}
		if err.Error() == "user not found" {
			utils.ReturnError(w, "NOT_FOUND", "Kullanıcı bulunamadı", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Kod gönderilemedi", err.Error())
		return
	}

	utils.WriteJson(w, map[string]string{"message": "Doğrulama kodu tekrar gönderildi"}, http.StatusOK, "Kod başarıyla gönderildi")
}
