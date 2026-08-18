package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/common/validation"
	"github.com/M1ralai/notly-api/internal/modules/course/dto"
	"github.com/M1ralai/notly-api/internal/modules/course/service"
	subscriptionService "github.com/M1ralai/notly-api/internal/modules/subscription/service"
	"github.com/gorilla/mux"
)

const maxCourseResourceUploadSize = 32 << 20 // 32 MB

type Handler struct {
	service service.CourseService
}

func NewHandler(service service.CourseService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/courses", h.GetAll).Methods("GET")
	router.HandleFunc("/courses/active", h.GetActive).Methods("GET")
	router.HandleFunc("/courses", h.Create).Methods("POST")
	router.HandleFunc("/courses/components", h.CreateComponent).Methods("POST")
	router.HandleFunc("/courses/{id}/components", h.GetComponents).Methods("GET")
	router.HandleFunc("/courses/components/{id}", h.UpdateComponent).Methods("PUT", "PATCH")
	router.HandleFunc("/courses/components/{id}", h.DeleteComponent).Methods("DELETE")
	router.HandleFunc("/courses/schedules", h.CreateSchedule).Methods("POST")
	router.HandleFunc("/courses/{id}/schedules", h.GetSchedules).Methods("GET")
	router.HandleFunc("/courses/schedules/{id}", h.UpdateSchedule).Methods("PUT", "PATCH")
	router.HandleFunc("/courses/schedules/{id}", h.DeleteSchedule).Methods("DELETE")
	router.HandleFunc("/courses/{id}/resources", h.GetResources).Methods("GET")
	router.HandleFunc("/courses/{id}/resources", h.CreateResource).Methods("POST")
	router.HandleFunc("/courses/{id}/resources/upload", h.UploadResourceFile).Methods("POST")
	router.HandleFunc("/courses/resources/{id}", h.UpdateResource).Methods("PUT", "PATCH")
	router.HandleFunc("/courses/resources/{id}", h.DeleteResource).Methods("DELETE")
	router.HandleFunc("/courses/{id}", h.GetByID).Methods("GET")
	router.HandleFunc("/courses/{id}", h.Update).Methods("PUT", "PATCH")
	router.HandleFunc("/courses/{id}", h.Delete).Methods("DELETE")
}

func (h *Handler) getUserID(r *http.Request) int {
	return utils.GetUserIDFromContext(r.Context())
}

func handleCourseError(w http.ResponseWriter, err error, fallbackMessage string) {
	if errors.Is(err, subscriptionService.ErrPremiumRequired) {
		utils.ReturnError(w, "PREMIUM_REQUIRED", "Notly Pro required", err.Error())
		return
	}

	switch err.Error() {
	case "course not found", "resource not found":
		utils.ReturnError(w, "NOT_FOUND", "Kaynak bulunamadı", err.Error())
	case "unauthorized":
		utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
	case "file upload required":
		utils.ReturnError(w, "BAD_REQUEST", "Dosya kaynakları upload endpoint'i ile eklenmelidir", err.Error())
	case "storage unavailable":
		utils.ReturnError(w, "INTERNAL_ERROR", "Dosya depolama şu an hazır değil", err.Error())
	default:
		utils.ReturnError(w, "INTERNAL_ERROR", fallbackMessage, err.Error())
	}
}

// @Summary Create
// @Tags Course
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateCourseRequest true "Create Course"
// @Success 201 {object} dto.CourseResponse
// @Router /api/courses [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)
	course, err := h.service.Create(r.Context(), &req, userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Ders oluşturulamadı", err.Error())
		return
	}

	utils.WriteJson(w, course, http.StatusCreated, "Ders oluşturuldu")
}

// @Summary CreateComponent
// @Tags Course
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateComponentRequest true "Create Component"
// @Success 201 {object} dto.ComponentResponse
// @Router /api/courses/components [post]
func (h *Handler) CreateComponent(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)
	component, err := h.service.CreateComponent(r.Context(), &req, userID)
	if err != nil {
		if err.Error() == "course not found" {
			utils.ReturnError(w, "NOT_FOUND", "Ders bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Component oluşturulamadı", err.Error())
		return
	}

	utils.WriteJson(w, component, http.StatusCreated, "Component oluşturuldu")
}

// @Summary GetComponents
// @Tags Course
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {array} dto.ComponentResponse
// @Router /api/courses/{id}/components [get]
func (h *Handler) GetComponents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	components, err := h.service.GetComponents(r.Context(), courseID, userID)
	if err != nil {
		if err.Error() == "course not found" {
			utils.ReturnError(w, "NOT_FOUND", "Ders bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Components getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, components, http.StatusOK, "Components getirildi")
}

// @Summary UpdateComponent
// @Tags Course
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body dto.UpdateComponentRequest true "Update Component"
// @Success 200 {object} dto.ComponentResponse
// @Router /api/courses/components/{id} [put]
func (h *Handler) UpdateComponent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	var req dto.UpdateComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)
	component, err := h.service.UpdateComponent(r.Context(), id, &req, userID)
	if err != nil {
		if err.Error() == "component not found" || err.Error() == "course not found" {
			utils.ReturnError(w, "NOT_FOUND", "Component bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Component güncellenemedi", err.Error())
		return
	}

	utils.WriteJson(w, component, http.StatusOK, "Component güncellendi")
}

// @Summary DeleteComponent
// @Tags Course
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} map[string]string
// @Router /api/courses/components/{id} [delete]
func (h *Handler) DeleteComponent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	if err := h.service.DeleteComponent(r.Context(), id, userID); err != nil {
		if err.Error() == "component not found" || err.Error() == "course not found" {
			utils.ReturnError(w, "NOT_FOUND", "Component bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Component silinemedi", err.Error())
		return
	}

	utils.WriteJson(w, nil, http.StatusOK, "Component silindi")
}

// @Summary CreateSchedule
// @Tags Course
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateScheduleRequest true "Create Schedule"
// @Success 201 {object} dto.ScheduleResponse
// @Router /api/courses/schedules [post]
func (h *Handler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)
	schedule, err := h.service.CreateSchedule(r.Context(), &req, userID)
	if err != nil {
		if err.Error() == "course not found" {
			utils.ReturnError(w, "NOT_FOUND", "Ders bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Schedule oluşturulamadı", err.Error())
		return
	}

	utils.WriteJson(w, schedule, http.StatusCreated, "Schedule oluşturuldu")
}

// @Summary GetSchedules
// @Tags Course
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {array} dto.ScheduleResponse
// @Router /api/courses/{id}/schedules [get]
func (h *Handler) GetSchedules(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	schedules, err := h.service.GetSchedules(r.Context(), courseID, userID)
	if err != nil {
		if err.Error() == "course not found" {
			utils.ReturnError(w, "NOT_FOUND", "Ders bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Schedules getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, schedules, http.StatusOK, "Schedules getirildi")
}

// @Summary UpdateSchedule
// @Tags Course
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body dto.UpdateScheduleRequest true "Update Schedule"
// @Success 200 {object} dto.ScheduleResponse
// @Router /api/courses/schedules/{id} [put]
func (h *Handler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	var req dto.UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)
	schedule, err := h.service.UpdateSchedule(r.Context(), id, &req, userID)
	if err != nil {
		if err.Error() == "schedule not found" || err.Error() == "course not found" {
			utils.ReturnError(w, "NOT_FOUND", "Schedule bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Schedule güncellenemedi", err.Error())
		return
	}

	utils.WriteJson(w, schedule, http.StatusOK, "Schedule güncellendi")
}

// @Summary DeleteSchedule
// @Tags Course
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} map[string]string
// @Router /api/courses/schedules/{id} [delete]
func (h *Handler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	if err := h.service.DeleteSchedule(r.Context(), id, userID); err != nil {
		if err.Error() == "schedule not found" || err.Error() == "course not found" {
			utils.ReturnError(w, "NOT_FOUND", "Schedule bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Schedule silinemedi", err.Error())
		return
	}

	utils.WriteJson(w, nil, http.StatusOK, "Schedule silindi")
}

// @Summary CreateResource
// @Tags Course
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Course ID"
// @Param request body dto.CreateResourceRequest true "Create Resource"
// @Success 201 {object} dto.ResourceResponse
// @Router /api/courses/{id}/resources [post]
func (h *Handler) CreateResource(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ders ID", err.Error())
		return
	}

	var req dto.CreateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}
	req.CourseID = courseID

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	resource, err := h.service.CreateResource(r.Context(), &req, h.getUserID(r))
	if err != nil {
		handleCourseError(w, err, "Kaynak oluşturulamadı")
		return
	}

	utils.WriteJson(w, resource, http.StatusCreated, "Kaynak oluşturuldu")
}

// @Summary UploadResourceFile
// @Tags Course
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Course ID"
// @Param file formData file true "File to upload"
// @Success 201 {object} dto.ResourceResponse
// @Router /api/courses/{id}/resources/upload [post]
func (h *Handler) UploadResourceFile(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ders ID", err.Error())
		return
	}

	if err := r.ParseMultipartForm(maxCourseResourceUploadSize); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Dosya çok büyük veya form hatalı", err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "file alanı eksik", err.Error())
		return
	}
	defer file.Close()

	metadata := dto.CreateResourceRequest{
		CourseID:    courseID,
		Type:        "file",
		Title:       strings.TrimSpace(r.FormValue("title")),
		Description: strings.TrimSpace(r.FormValue("description")),
		Tags:        parseResourceTags(r.FormValue("tags")),
		IsPrimary:   r.FormValue("is_primary") == "true",
	}
	if componentID, ok := parseOptionalInt(r.FormValue("component_id")); ok {
		metadata.ComponentID = &componentID
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	resource, err := h.service.UploadResourceFile(
		r.Context(),
		courseID,
		h.getUserID(r),
		file,
		header.Filename,
		contentType,
		header.Size,
		&metadata,
	)
	if err != nil {
		handleCourseError(w, err, "Dosya yüklenemedi")
		return
	}

	utils.WriteJson(w, resource, http.StatusCreated, "Dosya yüklendi")
}

// @Summary GetResources
// @Tags Course
// @Security BearerAuth
// @Produce json
// @Param id path int true "Course ID"
// @Success 200 {array} dto.ResourceResponse
// @Router /api/courses/{id}/resources [get]
func (h *Handler) GetResources(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ders ID", err.Error())
		return
	}

	resources, err := h.service.GetResources(r.Context(), courseID, h.getUserID(r))
	if err != nil {
		handleCourseError(w, err, "Kaynaklar getirilemedi")
		return
	}

	utils.WriteJson(w, resources, http.StatusOK, "Kaynaklar getirildi")
}

// @Summary UpdateResource
// @Tags Course
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Resource ID"
// @Param request body dto.UpdateResourceRequest true "Update Resource"
// @Success 200 {object} dto.ResourceResponse
// @Router /api/courses/resources/{id} [put]
func (h *Handler) UpdateResource(w http.ResponseWriter, r *http.Request) {
	resourceID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz kaynak ID", err.Error())
		return
	}

	var req dto.UpdateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}
	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	resource, err := h.service.UpdateResource(r.Context(), resourceID, &req, h.getUserID(r))
	if err != nil {
		handleCourseError(w, err, "Kaynak güncellenemedi")
		return
	}

	utils.WriteJson(w, resource, http.StatusOK, "Kaynak güncellendi")
}

// @Summary DeleteResource
// @Tags Course
// @Security BearerAuth
// @Produce json
// @Param id path int true "Resource ID"
// @Success 200 {object} map[string]string
// @Router /api/courses/resources/{id} [delete]
func (h *Handler) DeleteResource(w http.ResponseWriter, r *http.Request) {
	resourceID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz kaynak ID", err.Error())
		return
	}

	if err := h.service.DeleteResource(r.Context(), resourceID, h.getUserID(r)); err != nil {
		handleCourseError(w, err, "Kaynak silinemedi")
		return
	}

	utils.WriteJson(w, nil, http.StatusOK, "Kaynak silindi")
}

func parseResourceTags(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func parseOptionalInt(raw string) (int, bool) {
	if strings.TrimSpace(raw) == "" {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	return value, err == nil
}

// @Summary GetByID
// @Tags Course
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} dto.CourseResponse
// @Router /api/courses/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	course, err := h.service.GetByID(r.Context(), id, userID)
	if err != nil {
		if err.Error() == "course not found" {
			utils.ReturnError(w, "NOT_FOUND", "Ders bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Ders getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, course, http.StatusOK, "Ders getirildi")
}

// @Summary GetAll
// @Tags Course
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.CourseResponse
// @Router /api/courses [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	courses, err := h.service.GetAll(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Dersler getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, courses, http.StatusOK, "Dersler getirildi")
}

// @Summary GetActive
// @Tags Course
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.CourseResponse
// @Router /api/courses/active [get]
func (h *Handler) GetActive(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	courses, err := h.service.GetActive(r.Context(), userID)
	if err != nil {
		utils.ReturnError(w, "INTERNAL_ERROR", "Aktif dersler getirilemedi", err.Error())
		return
	}

	utils.WriteJson(w, courses, http.StatusOK, "Aktif dersler getirildi")
}

// @Summary Update
// @Tags Course
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param request body dto.UpdateCourseRequest true "Update Course"
// @Success 200 {object} dto.CourseResponse
// @Router /api/courses/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	var req dto.UpdateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz istek formatı", err.Error())
		return
	}

	if err := validation.Get().Struct(req); err != nil {
		utils.ReturnError(w, "VALIDATION_ERROR", "Doğrulama hatası", validation.FormatErr(err))
		return
	}

	userID := h.getUserID(r)
	course, err := h.service.Update(r.Context(), id, &req, userID)
	if err != nil {
		if err.Error() == "course not found" {
			utils.ReturnError(w, "NOT_FOUND", "Ders bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Ders güncellenemedi", err.Error())
		return
	}

	utils.WriteJson(w, course, http.StatusOK, "Ders güncellendi")
}

// @Summary Delete
// @Tags Course
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} map[string]string
// @Router /api/courses/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Geçersiz ID", err.Error())
		return
	}

	userID := h.getUserID(r)
	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		if err.Error() == "course not found" {
			utils.ReturnError(w, "NOT_FOUND", "Ders bulunamadı", err.Error())
			return
		}
		if err.Error() == "unauthorized" {
			utils.ReturnError(w, "FORBIDDEN", "Bu işlem için yetkiniz yok", err.Error())
			return
		}
		utils.ReturnError(w, "INTERNAL_ERROR", "Ders silinemedi", err.Error())
		return
	}

	utils.WriteJson(w, nil, http.StatusOK, "Ders silindi")
}
