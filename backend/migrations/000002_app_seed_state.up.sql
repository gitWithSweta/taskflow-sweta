CREATE TABLE IF NOT EXISTS app_seed_state (
    key TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
