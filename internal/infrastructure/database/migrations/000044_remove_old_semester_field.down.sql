ALTER TABLE courses ADD COLUMN semester VARCHAR(50);
CREATE INDEX idx_courses_semester ON courses(semester);
