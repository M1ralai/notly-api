-- SEO Pages Table for Programmatic SEO
CREATE TABLE IF NOT EXISTS seo_pages (
    id SERIAL PRIMARY KEY,

    -- URL & Routing
    slug VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('tool', 'template', 'guide', 'feature', 'core')),

    -- SEO Metadata
    title VARCHAR(255) NOT NULL,
    meta_description TEXT NOT NULL,
    meta_keywords TEXT[],

    -- Content Configuration (JSONB for flexibility)
    content_config JSONB DEFAULT '{}'::jsonb,

    -- Sitemap Settings
    priority DECIMAL(3,2) DEFAULT 0.80 CHECK (priority BETWEEN 0.00 AND 1.00),
    changefreq VARCHAR(20) DEFAULT 'monthly' CHECK (changefreq IN ('always', 'hourly', 'daily', 'weekly', 'monthly', 'yearly', 'never')),

    -- Status & Timestamps
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for Performance
CREATE UNIQUE INDEX idx_seo_pages_slug ON seo_pages(slug);
CREATE INDEX idx_seo_pages_type_active ON seo_pages(type, is_active) WHERE is_active = TRUE;
CREATE INDEX idx_seo_pages_priority ON seo_pages(priority DESC) WHERE is_active = TRUE;

-- Updated_at Trigger
CREATE OR REPLACE FUNCTION update_seo_pages_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_seo_pages_updated_at
BEFORE UPDATE ON seo_pages
FOR EACH ROW
EXECUTE FUNCTION update_seo_pages_updated_at();
