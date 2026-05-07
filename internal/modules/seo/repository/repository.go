package repository

import (
	"context"

	"github.com/M1ralai/notly-api/internal/modules/seo/domain"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Repository defines the SEO page data access layer
type Repository interface {
	GetAllActivePages(ctx context.Context) ([]domain.SEOPage, error)
	GetPageBySlug(ctx context.Context, slug string) (*domain.SEOPage, error)
	GetPagesByType(ctx context.Context, pageType string) ([]domain.SEOPage, error)
	GetRelatedPages(ctx context.Context, pageID int, limit int) ([]domain.SEOPage, error)
	GetPagesByCategory(ctx context.Context, category string) ([]domain.SEOPage, error)
}

type postgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL repository for SEO pages
func NewPostgresRepository(db *sqlx.DB) Repository {
	return &postgresRepository{db: db}
}

const selectColumns = `
	id, slug, type, title, meta_description, meta_keywords,
	content_config, priority, changefreq, is_active,
	category, parent_page_id, related_page_ids,
	content_template_key, template_variables, rendered_content, faq_items,
	created_at, updated_at
`

func (r *postgresRepository) scanRow(row *sqlx.Rows) (domain.SEOPage, error) {
	var page domain.SEOPage
	err := row.Scan(
		&page.ID,
		&page.Slug,
		&page.Type,
		&page.Title,
		&page.MetaDescription,
		pq.Array(&page.MetaKeywords),
		&page.ContentConfig,
		&page.Priority,
		&page.Changefreq,
		&page.IsActive,
		&page.Category,
		&page.ParentPageID,
		pq.Array(&page.RelatedPageIDs),
		&page.ContentTemplateKey,
		&page.TemplateVariables,
		&page.RenderedContent,
		&page.FAQItems,
		&page.CreatedAt,
		&page.UpdatedAt,
	)
	return page, err
}

// GetAllActivePages retrieves all active SEO pages ordered by priority
func (r *postgresRepository) GetAllActivePages(ctx context.Context) ([]domain.SEOPage, error) {
	query := `SELECT ` + selectColumns + ` FROM seo_pages WHERE is_active = TRUE ORDER BY priority DESC, slug ASC`

	var pages []domain.SEOPage
	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		page, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}

// GetPageBySlug retrieves a single SEO page by its slug
func (r *postgresRepository) GetPageBySlug(ctx context.Context, slug string) (*domain.SEOPage, error) {
	query := `SELECT ` + selectColumns + ` FROM seo_pages WHERE slug = $1 AND is_active = TRUE LIMIT 1`

	var page domain.SEOPage
	// QueryRowxContext is slightly different for scanning so we do it manually to match scanRow structure or use Queryx for consistency + Next()
	rows, err := r.db.QueryxContext(ctx, query, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		page, err = r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		return &page, nil
	}

	return nil, nil // Not found
}

// GetPagesByType retrieves all active SEO pages of a specific type
func (r *postgresRepository) GetPagesByType(ctx context.Context, pageType string) ([]domain.SEOPage, error) {
	query := `SELECT ` + selectColumns + ` FROM seo_pages WHERE type = $1 AND is_active = TRUE ORDER BY priority DESC, slug ASC`

	var pages []domain.SEOPage
	rows, err := r.db.QueryxContext(ctx, query, pageType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		page, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}

// GetRelatedPages retrieves related pages based on category or explicit relationships
func (r *postgresRepository) GetRelatedPages(ctx context.Context, pageID int, limit int) ([]domain.SEOPage, error) {
	query := `
		WITH current_page AS (
			SELECT category, related_page_ids
			FROM seo_pages
			WHERE id = $1 AND is_active = TRUE
		)
		SELECT DISTINCT
			sp.id, sp.slug, sp.type, sp.title, sp.meta_description, sp.meta_keywords,
			sp.content_config, sp.priority, sp.changefreq, sp.is_active,
			sp.category, sp.parent_page_id, sp.related_page_ids,
			sp.content_template_key, sp.template_variables, sp.rendered_content, sp.faq_items,
			sp.created_at, sp.updated_at
		FROM seo_pages sp
		CROSS JOIN current_page cp
		WHERE sp.is_active = TRUE
		  AND sp.id != $1
		  AND (
		      (sp.category = cp.category AND cp.category IS NOT NULL)
		      OR
		      sp.id = ANY(cp.related_page_ids)
		  )
		ORDER BY sp.priority DESC, sp.slug ASC
		LIMIT $2
	`

	var pages []domain.SEOPage
	rows, err := r.db.QueryxContext(ctx, query, pageID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		page, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}

// GetPagesByCategory retrieves all active pages in a specific category
func (r *postgresRepository) GetPagesByCategory(ctx context.Context, category string) ([]domain.SEOPage, error) {
	query := `SELECT ` + selectColumns + ` FROM seo_pages WHERE category = $1 AND is_active = TRUE ORDER BY priority DESC, slug ASC`

	var pages []domain.SEOPage
	rows, err := r.db.QueryxContext(ctx, query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		page, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}
