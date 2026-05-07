-- Rollback SEO Relationships
-- Migration: 000036_add_seo_relationships.down.sql

-- Drop indexes
DROP INDEX IF EXISTS idx_seo_pages_category;
DROP INDEX IF EXISTS idx_seo_pages_parent;

-- Remove columns
ALTER TABLE seo_pages
    DROP COLUMN IF EXISTS category,
    DROP COLUMN IF EXISTS parent_page_id,
    DROP COLUMN IF EXISTS related_page_ids;
