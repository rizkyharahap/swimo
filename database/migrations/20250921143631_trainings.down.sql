-- Drop indexes
DROP INDEX IF EXISTS idx_training_sessions_user_created_at;
DROP INDEX IF EXISTS idx_trainings_desc_trgm;
DROP INDEX IF EXISTS idx_trainings_name_trgm;
DROP INDEX IF EXISTS idx_trainings_category;

-- Drop tables
DROP TABLE IF EXISTS training_sessions;
DROP TABLE IF EXISTS trainings;
DROP TABLE IF EXISTS training_categories;
