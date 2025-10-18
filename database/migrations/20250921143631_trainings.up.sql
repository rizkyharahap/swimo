-- Trainings categories
CREATE TABLE training_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,               -- e.g. 'BREASTSTROKE'
    name VARCHAR(100) NOT NULL,                     -- e.g. 'Breaststroke'
    description TEXT,
    met NUMERIC(4,1) NOT NULL,                      -- MET baseline for calori calculation
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 2) Seed kategori + MET
INSERT INTO training_categories (code, name, description, met) VALUES
('FREESTYLE',         'Freestyle',           'Front crawl umum; pace moderat',                      8.3),
('BREASTSTROKE',      'Breaststroke',        'Gaya dada; relatif lebih berat',                     10.3),
('BACKSTROKE',        'Backstroke',          'Gaya punggung; intensitas menengah-tinggi',           9.5),
('BUTTERFLY',         'Butterfly',           'Gaya kupu-kupu; paling berat',                       13.8),
('INDIVIDUAL_MEDLEY', 'Individual Medley',   'Campuran 4 gaya; rata-rata intensitas tinggi',        9.8),
('KICK',              'Kick Set',            'Papan kaki; kerja kaki dominan',                      8.0),
('PULL',              'Pull Set',            'Pull buoy; kerja lengan dominan',                     7.5),
('DRILL',             'Drill Technique',     'Teknik/skill fokus',                                  6.0),
('WARM_UP',           'Warm Up',             'Pemanasan ringan',                                     5.0),
('COOL_DOWN',         'Cool Down',           'Pendinginan sangat ringan',                            4.0),
('OPEN_WATER',        'Open Water',          'Renang perairan terbuka; navigasi & gelombang',        9.8);

-- Trainings
CREATE TABLE trainings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES training_categories(id) ON DELETE RESTRICT,

    level VARCHAR(50) NOT NULL,             -- difficulty level, e.g. beginner/intermediate
    name VARCHAR(100) NOT NULL,             -- training name
    descriptions TEXT NOT NULL,             -- short description
    time_label VARCHAR(50) NOT NULL,        -- e.g. "10-15 min"
    calories_kcal INT NOT NULL,             -- estimated calories burned per session
    thumbnail_url TEXT NOT NULL,            -- thumbnail image URL
    video_url TEXT,                         -- video tutorial URL
    content_html TEXT NOT NULL,             -- rich HTML content

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_trainings_name UNIQUE (name)
);

-- Index for general query
CREATE INDEX IF NOT EXISTS idx_trainings_category ON trainings (category_id);
-- Trigram index for searching ILIKE faster (name & description)
CREATE INDEX IF NOT EXISTS idx_trainings_name_trgm ON trainings USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_trainings_desc_trgm ON trainings USING gin (descriptions gin_trgm_ops);


-- Training Sessions
CREATE TABLE training_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    training_id UUID REFERENCES trainings(id) ON DELETE SET NULL,

    distance_meters INT NOT NULL,   -- distance in meters
    duration_seconds INT NOT NULL,  -- duration in seconds
    pace NUMERIC(6,2) NOT NULL,     -- pace in minutes/100m
    calories_kcal INT NOT NULL,     -- calories burned per session

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index to quickly fetch the latest session for a user
CREATE INDEX IF NOT EXISTS idx_training_sessions_user_created_at
    ON training_sessions (user_id, created_at DESC);
