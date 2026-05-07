-- Add Content Templates for Programmatic Content Generation
-- Migration: 000037_add_content_templates.up.sql

-- Create content_templates table
CREATE TABLE IF NOT EXISTS content_templates (
    id SERIAL PRIMARY KEY,
    template_key VARCHAR(100) UNIQUE NOT NULL,
    template_text TEXT NOT NULL,
    variables TEXT[] NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on template_key for fast lookups
CREATE UNIQUE INDEX idx_content_templates_key ON content_templates(template_key);

-- Add content template fields to seo_pages
ALTER TABLE seo_pages
    ADD COLUMN content_template_key VARCHAR(100) REFERENCES content_templates(template_key) ON DELETE SET NULL,
    ADD COLUMN template_variables JSONB DEFAULT '{}'::jsonb,
    ADD COLUMN rendered_content TEXT,
    ADD COLUMN faq_items JSONB DEFAULT '[]'::jsonb;

-- Create index for template lookups
CREATE INDEX idx_seo_pages_template ON seo_pages(content_template_key) WHERE content_template_key IS NOT NULL;

-- Insert default content templates
INSERT INTO content_templates (template_key, template_text, variables) VALUES
('university_calculator',
 '{university} öğrencileri için hazırladığımız {system} sistemine uygun not hesaplama aracı. {university} yönetmeliğine göre geçme notu {passing_grade} puandır. Vize sınavının ağırlığı %{vize_weight}, final sınavının ağırlığı %{final_weight} olarak hesaplanır. Bu araç ile notlarınızı kolayca hesaplayabilir ve akademik başarınızı takip edebilirsiniz.',
 ARRAY['university', 'system', 'passing_grade', 'vize_weight', 'final_weight']),

('yks_calculator_intro',
 '{year} YKS sınavına hazırlanan öğrenciler için geliştirilmiş ücretsiz puan hesaplama aracı. {exam_type} netlerinizi girerek tahmini puanınızı ve sıralamanızı öğrenebilirsiniz. ÖSYM katsayıları ile güncel hesaplama yapılmaktadır.',
 ARRAY['year', 'exam_type']),

('gpa_calculator_intro',
 'Üniversite öğrencileri için {system} not sistemine uygun GPA hesaplama aracı. Ders notlarınızı ve kredilerinizi girerek {calculation_type} hesaplayabilirsiniz. Tüm Türkiye üniversitelerinde geçerli standart hesaplama yöntemi kullanılmaktadır.',
 ARRAY['system', 'calculation_type']),

('vize_final_intro',
 'Vize ve final sınavlarınız için geliştirilmiş hesaplama aracı. Vize notunuzu girerek finalden kaç almanız gerektiğini öğrenebilir, {weight_system} ağırlık sistemine göre hesaplama yapabilirsiniz. Dersi geçmek için gereken minimum final notunu anında hesaplayın.',
 ARRAY['weight_system']);

-- Add updated_at trigger for content_templates
CREATE OR REPLACE FUNCTION update_content_templates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_content_templates_updated_at
BEFORE UPDATE ON content_templates
FOR EACH ROW
EXECUTE FUNCTION update_content_templates_updated_at();

-- Add comments
COMMENT ON TABLE content_templates IS 'Reusable content templates for programmatic content generation';
COMMENT ON COLUMN seo_pages.content_template_key IS 'Reference to content template for dynamic content generation';
COMMENT ON COLUMN seo_pages.template_variables IS 'JSONB object containing variables to render in the template';
COMMENT ON COLUMN seo_pages.rendered_content IS 'Pre-rendered content from template (cached)';
COMMENT ON COLUMN seo_pages.faq_items IS 'JSONB array of FAQ items for rich snippets [{question, answer}]';
