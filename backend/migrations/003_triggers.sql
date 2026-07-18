-- ============================================
-- CloudX Triggers
-- ============================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;


-- USERS
CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- ORGANIZATIONS
CREATE TRIGGER trg_organizations_updated_at
BEFORE UPDATE ON organizations
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- PROJECTS
CREATE TRIGGER trg_projects_updated_at
BEFORE UPDATE ON projects
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- ENVIRONMENTS
CREATE TRIGGER trg_environments_updated_at
BEFORE UPDATE ON environments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- SERVICES
CREATE TRIGGER trg_services_updated_at
BEFORE UPDATE ON services
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- DEPLOYMENTS
CREATE TRIGGER trg_deployments_updated_at
BEFORE UPDATE ON deployments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- BUILDS
CREATE TRIGGER trg_builds_updated_at
BEFORE UPDATE ON builds
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- CONTAINERS
CREATE TRIGGER trg_containers_updated_at
BEFORE UPDATE ON containers
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- DOMAINS
CREATE TRIGGER trg_domains_updated_at
BEFORE UPDATE ON domains
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- SSL CERTIFICATES
CREATE TRIGGER trg_ssl_updated_at
BEFORE UPDATE ON ssl_certificates
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- GITHUB INSTALLATIONS
CREATE TRIGGER trg_github_installations_updated_at
BEFORE UPDATE ON github_installations
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- GITHUB REPOSITORIES
CREATE TRIGGER trg_github_repositories_updated_at
BEFORE UPDATE ON github_repositories
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


-- API KEYS
CREATE TRIGGER trg_api_keys_updated_at
BEFORE UPDATE ON api_keys
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();