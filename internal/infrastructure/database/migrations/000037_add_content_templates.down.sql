-- Rollback Content Templates
-- Migration: 000037_add_content_templates.down.sql

-- Drop trigger
DROP TRIGGER IF EXISTS trigger_content_templates_updated_at ON content_templates;
DROP FUNCTION IF EXISTS update_content_templates_updated_at();

-- Remove columns from seo_pages
ALTER TABLE seo_pages
    DROP COLUMN IF EXISTS content_template_key,
    DROP COLUMN IF EXISTS template_variables,
    DROP COLUMN IF EXISTS rendered_content,
    DROP COLUMN IF EXISTS faq_items;

-- Drop indexes
DROP INDEX IF EXISTS idx_seo_pages_template;
DROP INDEX IF EXISTS idx_content_templates_key;

-- Drop table
DROP TABLE IF EXISTS content_templates;
