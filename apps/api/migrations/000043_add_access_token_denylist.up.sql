CREATE TABLE IF NOT EXISTS access_token_denylist (
    jti VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_access_token_denylist_expires ON access_token_denylist(expires_at);
CREATE INDEX idx_access_token_denylist_user ON access_token_denylist(user_id);
