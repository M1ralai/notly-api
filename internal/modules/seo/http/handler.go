package http

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/M1ralai/notly-api/internal/modules/seo/domain"
	"github.com/M1ralai/notly-api/internal/modules/seo/service"
	"github.com/gorilla/mux"
)

const siteURL = "https://notly.app"

// Handler handles HTTP requests for SEO module
type Handler struct {
	service service.Service
}

// NewHandler creates a new SEO HTTP handler
func NewHandler(svc service.Service) *Handler {
	return &Handler{service: svc}
}

// GetAllPages handles GET /api/seo/pages - Returns all active SEO pages
// @Summary GetAllPages
// @Tags Seo
// @Security BearerAuth
// @Produce json
// @Success 200 {array} domain.SEOPage
// @Router /api/seo/pages [get]
func (h *Handler) GetAllPages(w http.ResponseWriter, r *http.Request) {
	pages, err := h.service.GetAllActivePages(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch SEO pages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pages)
}

// GetPageBySlug handles GET /api/seo/pages/{slug} - Returns single page by slug
// @Summary GetPageBySlug
// @Tags Seo
// @Security BearerAuth
// @Produce json
// @Param slug path string true "Slug"
// @Success 200 {object} domain.SEOPage
// @Router /api/seo/pages/{slug} [get]
func (h *Handler) GetPageBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	page, err := h.service.GetPageBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "Failed to fetch SEO page", http.StatusInternalServerError)
		return
	}

	if page == nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(page)
}

// GenerateSitemap handles GET /sitemap.xml - Generates dynamic sitemap from database
// @Summary GenerateSitemap
// @Tags Seo
// @Security BearerAuth
// @Produce xml
// @Success 200 {string} string "XML Sitemap"
// @Router /api/sitemap.xml [get]
func (h *Handler) GenerateSitemap(w http.ResponseWriter, r *http.Request) {
	pages, err := h.service.GetAllActivePages(r.Context())
	if err != nil {
		http.Error(w, "Failed to generate sitemap", http.StatusInternalServerError)
		return
	}

	// Build sitemap URLs
	urls := make([]domain.SitemapURL, 0, len(pages)+1)

	// Add dynamic SEO pages from database
	for _, page := range pages {
		var pathPrefix string
		switch page.Type {
		case "tool":
			pathPrefix = "/araclar"
		case "template":
			pathPrefix = "/sablonlar"
		case "guide":
			pathPrefix = "/rehber"
		case "feature":
			pathPrefix = "/ozellikler"
		case "core":
			// Core pages use their slug directly (e.g., /tasks, /habits)
			if page.Slug == "home" {
				pathPrefix = ""
			} else {
				pathPrefix = ""
			}
		default:
			pathPrefix = ""
		}

		// Build URL
		var loc string
		if page.Slug == "home" {
			loc = siteURL + "/"
		} else if page.Type == "core" {
			loc = siteURL + "/" + page.Slug
		} else {
			loc = siteURL + pathPrefix + "/" + page.Slug
		}

		urls = append(urls, domain.SitemapURL{
			Loc:        loc,
			Lastmod:    page.UpdatedAt.Format("2006-01-02"),
			Changefreq: page.Changefreq,
			Priority:   page.Priority,
		})
	}

	// Generate XML
	sitemap := domain.Urlset{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	// Set headers for XML and caching
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")

	// Write XML
	w.Write([]byte(xml.Header))
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	encoder.Encode(sitemap)
}

// GetRelatedPages handles GET /api/seo/pages/{id}/related - Returns related pages for internal linking
// @Summary GetRelatedPages
// @Tags Seo
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID"
// @Param limit query int false "Limit"
// @Success 200 {array} domain.SEOPage
// @Router /api/seo/pages/{id}/related [get]
func (h *Handler) GetRelatedPages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["id"]

	// Parse page ID
	var id int
	if _, err := fmt.Sscanf(pageID, "%d", &id); err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	// Get limit from query params (default 6)
	limit := 6
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if _, err := fmt.Sscanf(limitParam, "%d", &limit); err == nil && limit > 0 && limit <= 20 {
			// Use parsed limit
		} else {
			limit = 6 // Reset to default if invalid
		}
	}

	pages, err := h.service.GetRelatedPages(r.Context(), id, limit)
	if err != nil {
		http.Error(w, "Failed to fetch related pages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pages)
}

// ServeEmbedWidget handles GET /embed/{slug} - Serves embeddable calculator widget
// @Summary ServeEmbedWidget
// @Tags Seo
// @Security BearerAuth
// @Produce html
// @Param slug path string true "Slug"
// @Success 200 {string} string "HTML Embed"
// @Router /api/embed/{slug} [get]
func (h *Handler) ServeEmbedWidget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	page, err := h.service.GetPageBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "Failed to fetch widget", http.StatusInternalServerError)
		return
	}

	if page == nil {
		http.Error(w, "Widget not found", http.StatusNotFound)
		return
	}

	// Generate minimal HTML embed with calculator + backlink
	embedHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            padding: 20px;
            background: #f9fafb;
        }
        .calculator {
            background: white;
            padding: 24px;
            border-radius: 12px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            max-width: 600px;
            margin: 0 auto;
        }
        .notly-credit {
            margin-top: 20px;
            text-align: center;
            padding-top: 16px;
            border-top: 1px solid #e5e7eb;
        }
        .notly-credit a {
            color: #3b82f6;
            text-decoration: none;
            font-weight: 500;
        }
        .notly-credit a:hover {
            text-decoration: underline;
        }
        h1 {
            font-size: 24px;
            margin-bottom: 16px;
            color: #111827;
        }
        p {
            color: #6b7280;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <div class="calculator">
        <h1>%s</h1>
        <p>%s</p>
        <div id="calculator-content">
            <!-- Calculator UI would be rendered here -->
            <p style="text-align: center; padding: 40px; color: #9ca3af;">
                Hesaplayıcı yükleniyor...
            </p>
        </div>
    </div>
    <div class="notly-credit">
        <a href="https://notly.app/araclar/%s" target="_blank" rel="noopener">
            ⚡ Powered by Notly - Ücretsiz Öğrenci Planlayıcı
        </a>
    </div>
</body>
</html>`, page.Title, page.Title, page.MetaDescription, page.Slug)

	// Set headers to allow embedding
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Frame-Options", "ALLOWALL")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.Write([]byte(embedHTML))
}

// RegisterRoutes registers all SEO module routes
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Public API routes (no auth required for SEO endpoints)
	router.HandleFunc("/api/seo/pages", h.GetAllPages).Methods("GET")
	router.HandleFunc("/api/seo/pages/{slug}", h.GetPageBySlug).Methods("GET")
	router.HandleFunc("/api/seo/pages/{id}/related", h.GetRelatedPages).Methods("GET")

	// Embed widget endpoint
	router.HandleFunc("/embed/{slug}", h.ServeEmbedWidget).Methods("GET")

	// Sitemap endpoint (outside /api prefix, public)
	router.HandleFunc("/sitemap.xml", h.GenerateSitemap).Methods("GET")
}
