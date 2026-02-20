CREATE TABLE IF NOT EXISTS analyses (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT       NOT NULL,
    idea_text   TEXT         NOT NULL,
    strategist  JSONB        NULL,
    financier   JSONB        NULL,
    auditor     JSONB        NULL,
    analyst     JSONB        NULL,
    moderator   JSONB        NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analyses_user_id ON analyses (user_id);
CREATE INDEX IF NOT EXISTS idx_analyses_created_at ON analyses (created_at DESC);

