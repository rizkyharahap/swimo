-- General extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto; -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pg_trgm;  -- fuzzy search
CREATE EXTENSION IF NOT EXISTS citext;   -- email case-insensitive