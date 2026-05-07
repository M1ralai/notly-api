package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// SEOPage represents a programmatic SEO page stored in the database
type SEOPage struct {
	ID              int           `json:"id" db:"id"`
	Slug            string        `json:"slug" db:"slug"`
	Type            string        `json:"type" db:"type"`
	Title           string        `json:"title" db:"title"`
	MetaDescription string        `json:"meta_description" db:"meta_description"`
	MetaKeywords    []string      `json:"meta_keywords" db:"meta_keywords"`
	ContentConfig   ContentConfig `json:"content_config" db:"content_config"`
	Priority        float64       `json:"priority" db:"priority"`
	Changefreq      string        `json:"changefreq" db:"changefreq"`
	IsActive        bool          `json:"is_active" db:"is_active"`

	// Internal Linking & Relationships
	Category       *string `json:"category,omitempty" db:"category"`
	ParentPageID   *int    `json:"parent_page_id,omitempty" db:"parent_page_id"`
	RelatedPageIDs []int   `json:"related_page_ids,omitempty" db:"related_page_ids"`

	// Content Templates & Dynamic Content
	ContentTemplateKey *string      `json:"content_template_key,omitempty" db:"content_template_key"`
	TemplateVariables  TemplateVars `json:"template_variables,omitempty" db:"template_variables"`
	RenderedContent    *string      `json:"rendered_content,omitempty" db:"rendered_content"`
	FAQItems           FAQItems     `json:"faq_items,omitempty" db:"faq_items"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ContentConfig is a JSONB field that stores flexible configuration
type ContentConfig map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (c ContentConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for JSONB
func (c *ContentConfig) Scan(value interface{}) error {
	if value == nil {
		*c = make(ContentConfig)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte{}, c)
	}

	return json.Unmarshal(bytes, c)
}

// TemplateVars represents template variables stored as JSONB
type TemplateVars map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (t TemplateVars) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan implements the sql.Scanner interface for JSONB
func (t *TemplateVars) Scan(value interface{}) error {
	if value == nil {
		*t = make(TemplateVars)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte{}, t)
	}

	return json.Unmarshal(bytes, t)
}

// FAQItem represents a single FAQ entry for schema markup
type FAQItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// FAQItems is a slice of FAQItem with JSONB support
type FAQItems []FAQItem

// Value implements the driver.Valuer interface for JSONB
func (f FAQItems) Value() (driver.Value, error) {
	return json.Marshal(f)
}

// Scan implements the sql.Scanner interface for JSONB
func (f *FAQItems) Scan(value interface{}) error {
	if value == nil {
		*f = []FAQItem{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte("[]"), f)
	}

	return json.Unmarshal(bytes, f)
}

// SitemapURL represents a single URL entry in the sitemap
type SitemapURL struct {
	Loc        string  `xml:"loc"`
	Lastmod    string  `xml:"lastmod"`
	Changefreq string  `xml:"changefreq"`
	Priority   float64 `xml:"priority"`
}

// Urlset represents the root element of the sitemap
type Urlset struct {
	XMLName string       `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}
