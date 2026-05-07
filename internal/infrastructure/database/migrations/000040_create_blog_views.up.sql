CREATE TABLE IF NOT EXISTS blog_views (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(255) NOT NULL,
    title TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer TEXT,
    reading_time_seconds INTEGER DEFAULT 0,
    viewed_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_blog_views_slug ON blog_views(slug);
CREATE INDEX idx_blog_views_viewed_at ON blog_views(viewed_at);
CREATE INDEX idx_blog_views_ip_slug ON blog_views(ip_address, slug);
