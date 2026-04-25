-- Migration 001: initial schema
-- Run this once when setting up the database.
-- In production, use golang-migrate or goose to run migrations automatically.

CREATE TABLE IF NOT EXISTS urls (
    -- short_code is the primary key: the 6-char random string like "a3f9c2"
    short_code  TEXT PRIMARY KEY,

    -- long_url is what we redirect to
    long_url    TEXT NOT NULL,

    -- clicks is incremented every time someone follows the short link
    clicks      INTEGER NOT NULL DEFAULT 0,

    -- created_at for analytics (when was this link created?)
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index on created_at so we can list "recently created URLs" efficiently.
-- Without this index, ORDER BY created_at scans the entire table.
CREATE INDEX IF NOT EXISTS idx_urls_created_at ON urls(created_at DESC);
