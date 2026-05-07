DROP TRIGGER IF EXISTS trigger_seo_pages_updated_at ON seo_pages;
DROP FUNCTION IF EXISTS update_seo_pages_updated_at();
DROP TABLE IF EXISTS seo_pages CASCADE;
