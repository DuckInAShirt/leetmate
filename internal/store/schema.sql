-- LeetMate schema, version 1.

CREATE TABLE IF NOT EXISTS problems (
    slug         TEXT PRIMARY KEY,
    frontend_id  TEXT NOT NULL,
    title        TEXT NOT NULL,
    difficulty   TEXT,
    tags         TEXT,            -- JSON array of tag names
    is_paid_only INTEGER DEFAULT 0,
    updated_at   TEXT
);

CREATE TABLE IF NOT EXISTS attempts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    slug        TEXT NOT NULL,
    started_at  TEXT,
    finished_at TEXT,
    ac          INTEGER DEFAULT 0,
    runtime_ms  INTEGER DEFAULT 0,
    memory_kb   INTEGER DEFAULT 0,
    rating      INTEGER DEFAULT 0,
    gave_up     INTEGER DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_attempts_slug ON attempts(slug);

CREATE TABLE IF NOT EXISTS cards (
    slug           TEXT PRIMARY KEY,
    fsrs_state     BLOB,
    due_at         TEXT,
    reps           INTEGER DEFAULT 0,
    lapses         INTEGER DEFAULT 0,
    last_review_at TEXT,
    stability      REAL DEFAULT 0,
    difficulty     REAL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS conversations (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    slug       TEXT,
    tier       TEXT,
    role       TEXT,
    content    TEXT,
    created_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_conv_slug ON conversations(slug, id);

CREATE TABLE IF NOT EXISTS weakness_tags (
    slug  TEXT,
    tag   TEXT,
    count INTEGER DEFAULT 0,
    PRIMARY KEY (slug, tag)
);

CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY
);

-- Study-plan progress: one row per (plan, problem) the learner has finished.
-- Absence of a row means "todo".
CREATE TABLE IF NOT EXISTS studyplan_progress (
    plan_id     TEXT NOT NULL,
    frontend_id TEXT NOT NULL,
    updated_at  TEXT,
    PRIMARY KEY (plan_id, frontend_id)
);
