-- Add SEO Relationships and Internal Linking Architecture
-- Migration: 000036_add_seo_relationships.up.sql

-- Add category and relationship fields to seo_pages
ALTER TABLE seo_pages
    ADD COLUMN category VARCHAR(100),
    ADD COLUMN parent_page_id INTEGER REFERENCES seo_pages(id) ON DELETE SET NULL,
    ADD COLUMN related_page_ids INTEGER[] DEFAULT '{}';

-- Create index for category queries (only active pages)
CREATE INDEX idx_seo_pages_category ON seo_pages(category) WHERE is_active = TRUE;

-- Create index for parent-child relationships
CREATE INDEX idx_seo_pages_parent ON seo_pages(parent_page_id) WHERE parent_page_id IS NOT NULL;

-- Update existing pages with categories based on their slugs
-- YKS Tools Category
UPDATE seo_pages
SET category = 'yks_tools'
WHERE slug IN (
    'yks-puan-hesaplama',
    'tyt-puan-hesaplama',
    'ayt-puan-hesaplama',
    'yks-sayisal-puan',
    'yks-sozel-puan',
    'yks-ea-puan',
    'yks-siralama-hesaplama',
    'obp-yks-puan',
    'obpsiz-yks-puan',
    'yks-net-puan'
);

-- GPA Tools Category
UPDATE seo_pages
SET category = 'gpa_tools'
WHERE slug IN (
    'gpa-hesaplama',
    'gano-hesaplama',
    'yano-hesaplama',
    'agno-hesaplama',
    '4luk-100luk-cevirme',
    '100luk-4luk-cevirme',
    'harf-notu-hesaplama'
);

-- Vize-Final Tools Category
UPDATE seo_pages
SET category = 'vize_final_tools'
WHERE slug IN (
    'vize-final-hesaplama',
    'dersi-gecme-notu',
    'finalden-kac-lazim',
    'bute-kalmamak',
    'vize-40-final-60',
    'vize-30-final-70',
    'can-egrisi-hesaplama'
);

-- University Specific Category
UPDATE seo_pages
SET category = 'university_specific'
WHERE slug IN (
    'itu-ortalama',
    'odtu-gpa',
    'aof-puan'
);

-- General Tools Category
UPDATE seo_pages
SET category = 'general_tools'
WHERE slug IN (
    'pomodoro-sayaci'
);

-- Hub Pages Category
UPDATE seo_pages
SET category = 'hub_pages'
WHERE type = 'hub';

-- Add comment
COMMENT ON COLUMN seo_pages.category IS 'Category for grouping related pages (e.g., yks_tools, gpa_tools, vize_final_tools)';
COMMENT ON COLUMN seo_pages.parent_page_id IS 'Parent page ID for hub-and-spoke model (nullable)';
COMMENT ON COLUMN seo_pages.related_page_ids IS 'Array of related page IDs for internal linking';
