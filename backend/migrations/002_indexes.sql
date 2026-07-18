-- ============================================
-- CloudX Indexes
-- ============================================

-- USERS
CREATE INDEX IF NOT EXISTS idx_users_email
ON users(email);

-- ORGANIZATIONS
CREATE INDEX IF NOT EXISTS idx_organizations_owner
ON organizations(owner_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_organizations_slug
ON organizations(slug);

-- MEMBERS
CREATE INDEX IF NOT EXISTS idx_members_org
ON organization_members(organization_id);

CREATE INDEX IF NOT EXISTS idx_members_user
ON organization_members(user_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_member_unique
ON organization_members(organization_id, user_id);

-- PROJECTS
CREATE INDEX IF NOT EXISTS idx_projects_org
ON projects(organization_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_slug
ON projects(organization_id, slug);

-- ENVIRONMENTS
CREATE INDEX IF NOT EXISTS idx_environment_project
ON environments(project_id);

-- SERVICES
CREATE INDEX IF NOT EXISTS idx_service_environment
ON services(environment_id);

CREATE INDEX IF NOT EXISTS idx_service_runtime
ON services(runtime);

-- DEPLOYMENTS
CREATE INDEX IF NOT EXISTS idx_deployment_service
ON deployments(service_id);

CREATE INDEX IF NOT EXISTS idx_deployment_status
ON deployments(status);

CREATE INDEX IF NOT EXISTS idx_deployment_created
ON deployments(created_at DESC);

-- BUILDS
CREATE INDEX IF NOT EXISTS idx_build_deployment
ON builds(deployment_id);

-- CONTAINERS
CREATE INDEX IF NOT EXISTS idx_container_deployment
ON containers(deployment_id);

CREATE INDEX IF NOT EXISTS idx_container_status
ON containers(status);

-- ENVIRONMENT VARIABLES
CREATE INDEX IF NOT EXISTS idx_env_service
ON environment_variables(service_id);

-- DOMAINS
CREATE UNIQUE INDEX IF NOT EXISTS idx_domain_name
ON domains(domain);

CREATE INDEX IF NOT EXISTS idx_domain_service
ON domains(service_id);

-- SSL
CREATE INDEX IF NOT EXISTS idx_ssl_domain
ON ssl_certificates(domain_id);

-- GITHUB INSTALLATIONS
CREATE INDEX IF NOT EXISTS idx_installation_org
ON github_installations(organization_id);

-- GITHUB REPOSITORIES
CREATE INDEX IF NOT EXISTS idx_repo_installation
ON github_repositories(installation_id);

-- ACTIVITY LOGS
CREATE INDEX IF NOT EXISTS idx_activity_org
ON activity_logs(organization_id);

CREATE INDEX IF NOT EXISTS idx_activity_user
ON activity_logs(user_id);

CREATE INDEX IF NOT EXISTS idx_activity_time
ON activity_logs(created_at DESC);

-- AUDIT LOGS
CREATE INDEX IF NOT EXISTS idx_audit_user
ON audit_logs(user_id);

CREATE INDEX IF NOT EXISTS idx_audit_resource
ON audit_logs(resource);

CREATE INDEX IF NOT EXISTS idx_audit_created
ON audit_logs(created_at DESC);

-- API KEYS
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_key_hash
ON api_keys(hashed_key);

-- METRICS
CREATE INDEX IF NOT EXISTS idx_metrics_container
ON metrics(container_id);

CREATE INDEX IF NOT EXISTS idx_metrics_recorded_at
ON metrics(recorded_at DESC);