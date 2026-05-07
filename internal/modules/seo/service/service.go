package service

import (
	"context"
	"regexp"

	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
	"github.com/M1ralai/notly-api/internal/modules/seo/domain"
	"github.com/M1ralai/notly-api/internal/modules/seo/repository"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

// Service defines the business logic for SEO operations
type Service interface {
	GetAllActivePages(ctx context.Context) ([]domain.SEOPage, error)
	GetPageBySlug(ctx context.Context, slug string) (*domain.SEOPage, error)
	GetPagesByType(ctx context.Context, pageType string) ([]domain.SEOPage, error)
	GetRelatedPages(ctx context.Context, pageID int, limit int) ([]domain.SEOPage, error)
	GetPagesByCategory(ctx context.Context, category string) ([]domain.SEOPage, error)
	ValidateSlug(slug string) bool
}

type seoService struct {
	repo   repository.Repository
	logger *logger.ZapLogger
}

// NewSEOService creates a new SEO service
func NewSEOService(repo repository.Repository, logger *logger.ZapLogger) Service {
	return &seoService{
		repo:   repo,
		logger: logger,
	}
}

// GetAllActivePages retrieves all active SEO pages
func (s *seoService) GetAllActivePages(ctx context.Context) ([]domain.SEOPage, error) {
	pages, err := s.repo.GetAllActivePages(ctx)
	if err != nil {
		s.logger.Error("Failed to get all active pages", err, nil)
		return nil, err
	}
	return pages, nil
}

// GetPageBySlug retrieves a single page by slug with validation
func (s *seoService) GetPageBySlug(ctx context.Context, slug string) (*domain.SEOPage, error) {
	if !s.ValidateSlug(slug) {
		// Invalid slug - just return nil, no need to log
		return nil, nil
	}

	page, err := s.repo.GetPageBySlug(ctx, slug)
	if err != nil {
		s.logger.Error("Failed to get page by slug", err, map[string]interface{}{"slug": slug})
		return nil, err
	}
	return page, nil
}

// GetPagesByType retrieves all pages of a specific type
func (s *seoService) GetPagesByType(ctx context.Context, pageType string) ([]domain.SEOPage, error) {
	pages, err := s.repo.GetPagesByType(ctx, pageType)
	if err != nil {
		s.logger.Error("Failed to get pages by type", err, map[string]interface{}{"type": pageType})
		return nil, err
	}
	return pages, nil
}

// GetRelatedPages retrieves related pages for internal linking
func (s *seoService) GetRelatedPages(ctx context.Context, pageID int, limit int) ([]domain.SEOPage, error) {
	pages, err := s.repo.GetRelatedPages(ctx, pageID, limit)
	if err != nil {
		s.logger.Error("Failed to get related pages", err, map[string]interface{}{"page_id": pageID, "limit": limit})
		return nil, err
	}
	return pages, nil
}

// GetPagesByCategory retrieves all pages in a specific category
func (s *seoService) GetPagesByCategory(ctx context.Context, category string) ([]domain.SEOPage, error) {
	pages, err := s.repo.GetPagesByCategory(ctx, category)
	if err != nil {
		s.logger.Error("Failed to get pages by category", err, map[string]interface{}{"category": category})
		return nil, err
	}
	return pages, nil
}

// ValidateSlug validates that a slug is URL-safe
func (s *seoService) ValidateSlug(slug string) bool {
	return slugRegex.MatchString(slug)
}
