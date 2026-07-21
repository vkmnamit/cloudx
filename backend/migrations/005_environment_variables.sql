

CREATE TABLE IF NOT EXISTS deployment_env_vars (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID         NOT NULL,
    key           VARCHAR(255) NOT NULL,
    value         TEXT         NOT NULL,
    is_secret     BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_deployment_env_deployment
        FOREIGN KEY(deployment_id)
        REFERENCES deployments(id)
        ON DELETE CASCADE,

    CONSTRAINT unique_deployment_env_key
        UNIQUE(deployment_id, key)
);

CREATE INDEX IF NOT EXISTS idx_deployment_env_vars_deployment_id
    ON deployment_env_vars(deployment_id);
