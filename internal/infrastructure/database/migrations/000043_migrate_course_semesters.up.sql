-- Add semester_id column
ALTER TABLE courses ADD COLUMN semester_id INTEGER REFERENCES semesters(id) ON DELETE SET NULL;

-- Migrate existing semester strings to semester table
INSERT INTO semesters (user_id, name, start_date, end_date, is_current)
SELECT DISTINCT
  user_id,
  semester,
  CURRENT_DATE,  -- Default start date
  CURRENT_DATE + INTERVAL '4 months',  -- Default end date (typical semester length)
  FALSE
FROM courses
WHERE semester IS NOT NULL AND semester != '';

-- Link courses to their semesters
UPDATE courses c
SET semester_id = s.id
FROM semesters s
WHERE c.user_id = s.user_id
  AND c.semester = s.name
  AND c.semester IS NOT NULL;

CREATE INDEX idx_courses_semester_id ON courses(semester_id);
