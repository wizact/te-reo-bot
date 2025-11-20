package repository

const schema = `
-- Words table: stores all Māori words and their metadata
CREATE TABLE IF NOT EXISTS words (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    day_index INTEGER UNIQUE,
    word TEXT NOT NULL,
    meaning TEXT NOT NULL,
    link TEXT,
    photo TEXT,
    photo_attribution TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT 1,
    
    CHECK (day_index IS NULL OR (day_index >= 1 AND day_index <= 366))
);

-- Index for fast lookups by day
CREATE INDEX IF NOT EXISTS idx_day_index ON words(day_index);

-- Index for active words
CREATE INDEX IF NOT EXISTS idx_active ON words(is_active);
`
