CREATE TABLE IF NOT EXISTS github_app_installation_states (
    state_hash TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_github_app_states_expiry
    ON github_app_installation_states(expires_at);

CREATE INDEX IF NOT EXISTS idx_github_installation_org_id
    ON github_installations(organization_id, installation_id);
