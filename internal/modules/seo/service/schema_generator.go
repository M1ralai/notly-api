package service

import (
	"fmt"
	"strings"

	"github.com/M1ralai/notly-api/internal/modules/seo/domain"
)

// SchemaGenerator generates JSON-LD schema markup for SEO pages
type SchemaGenerator struct{}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator() *SchemaGenerator {
	return &SchemaGenerator{}
}

// GenerateSchemaMarkup generates appropriate schema.org markup based on page type
func (g *SchemaGenerator) GenerateSchemaMarkup(page *domain.SEOPage) map[string]interface{} {
	switch page.Type {
	case "tool":
		return g.generateToolSchema(page)
	case "guide":
		return g.generateArticleSchema(page)
	case "hub":
		return g.generateCollectionPageSchema(page)
	default:
		return g.generateWebPageSchema(page)
	}
}

// generateToolSchema generates SoftwareApplication schema for calculator tools
func (g *SchemaGenerator) generateToolSchema(page *domain.SEOPage) map[string]interface{} {
	schema := map[string]interface{}{
		"@context":            "https://schema.org",
		"@type":               "SoftwareApplication",
		"name":                page.Title,
		"applicationCategory": "EducationalApplication",
		"operatingSystem":     "Web Browser",
		"description":         page.MetaDescription,
		"offers": map[string]interface{}{
			"@type":         "Offer",
			"price":         "0",
			"priceCurrency": "TRY",
			"availability":  "https://schema.org/InStock",
		},
		"aggregateRating": map[string]interface{}{
			"@type":       "AggregateRating",
			"ratingValue": "4.8",
			"ratingCount": "1200",
			"bestRating":  "5",
			"worstRating": "1",
		},
	}

	// Add FAQ schema if FAQ items exist
	if len(page.FAQItems) > 0 {
		faqSchema := g.generateFAQSchema(page.FAQItems)
		// Return both schemas as an array
		return map[string]interface{}{
			"@context": "https://schema.org",
			"@graph": []interface{}{
				schema,
				faqSchema,
			},
		}
	}

	return schema
}

// generateArticleSchema generates Article schema for guides
func (g *SchemaGenerator) generateArticleSchema(page *domain.SEOPage) map[string]interface{} {
	return map[string]interface{}{
		"@context":    "https://schema.org",
		"@type":       "Article",
		"headline":    page.Title,
		"description": page.MetaDescription,
		"author": map[string]interface{}{
			"@type": "Organization",
			"name":  "Notly",
			"url":   "https://notly.app",
		},
		"publisher": map[string]interface{}{
			"@type": "Organization",
			"name":  "Notly",
			"logo": map[string]interface{}{
				"@type": "ImageObject",
				"url":   "https://notly.app/logo.png",
			},
		},
		"datePublished": page.CreatedAt.Format("2006-01-02"),
		"dateModified":  page.UpdatedAt.Format("2006-01-02"),
	}
}

// generateCollectionPageSchema generates CollectionPage schema for hub pages
func (g *SchemaGenerator) generateCollectionPageSchema(page *domain.SEOPage) map[string]interface{} {
	return map[string]interface{}{
		"@context":    "https://schema.org",
		"@type":       "CollectionPage",
		"name":        page.Title,
		"description": page.MetaDescription,
		"url":         fmt.Sprintf("https://notly.app/araclar/%s", page.Slug),
	}
}

// generateWebPageSchema generates basic WebPage schema
func (g *SchemaGenerator) generateWebPageSchema(page *domain.SEOPage) map[string]interface{} {
	return map[string]interface{}{
		"@context":    "https://schema.org",
		"@type":       "WebPage",
		"name":        page.Title,
		"description": page.MetaDescription,
		"url":         fmt.Sprintf("https://notly.app/araclar/%s", page.Slug),
	}
}

// generateFAQSchema generates FAQPage schema from FAQ items
func (g *SchemaGenerator) generateFAQSchema(faqItems []domain.FAQItem) map[string]interface{} {
	mainEntity := make([]map[string]interface{}, 0, len(faqItems))

	for _, item := range faqItems {
		mainEntity = append(mainEntity, map[string]interface{}{
			"@type": "Question",
			"name":  item.Question,
			"acceptedAnswer": map[string]interface{}{
				"@type": "Answer",
				"text":  item.Answer,
			},
		})
	}

	return map[string]interface{}{
		"@context":   "https://schema.org",
		"@type":      "FAQPage",
		"mainEntity": mainEntity,
	}
}

// RenderContent renders a template with variables
func RenderContent(templateText string, variables map[string]interface{}) string {
	result := templateText
	for key, value := range variables {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}
