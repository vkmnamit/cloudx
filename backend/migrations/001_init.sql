-- ============================================
-- CloudX Initial Schema
-- Part 1
-- ============================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- ENUMS
-- ============================================

CREATE TYPE organization_plan AS ENUM (
    'FREE',
    'PRO',
    'ENTERPRISE'
);

CREATE TYPE member_role AS ENUM (
    'OWNER',
    'ADMIN',
    'DEVELOPER',
    'VIEWER'
);

CREATE TYPE project_visibility AS ENUM (
    'PRIVATE',
    'PUBLIC'
);

CREATE TYPE deployment_status AS ENUM (
    'QUEUED',
    'BUILDING',
    'DEPLOYING',
    'RUNNING',
    'FAILED',
    'CANCELLED'
);

CREATE TYPE container_status AS ENUM (
    'CREATING',
    'RUNNING',
    'STOPPED',
    'FAILED'
);

-- ============================================
-- USERS
-- ============================================

CREATE TABLE users (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name VARCHAR(100) NOT NULL,

    email VARCHAR(255) UNIQUE NOT NULL,

    password_hash TEXT NOT NULL,

    avatar_url TEXT,

    email_verified BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW()

);

-- ============================================
-- ORGANIZATIONS
-- ============================================

CREATE TABLE organizations (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    owner_id UUID NOT NULL,

    name VARCHAR(120) NOT NULL,

    slug VARCHAR(120) UNIQUE NOT NULL,

    plan organization_plan NOT NULL DEFAULT 'FREE',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_org_owner
        FOREIGN KEY(owner_id)
        REFERENCES users(id)
        ON DELETE CASCADE

);

-- ============================================
-- ORGANIZATION MEMBERS
-- ============================================

CREATE TABLE organization_members (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    organization_id UUID NOT NULL,

    user_id UUID NOT NULL,

    role member_role NOT NULL DEFAULT 'DEVELOPER',

    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_member_org
        FOREIGN KEY(organization_id)
        REFERENCES organizations(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_member_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT unique_member
        UNIQUE(organization_id,user_id)

);

-- ============================================
-- PROJECTS
-- ============================================

CREATE TABLE projects (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    organization_id UUID NOT NULL,

    name VARCHAR(120) NOT NULL,

    slug VARCHAR(120) NOT NULL,

    description TEXT,

    visibility project_visibility
        DEFAULT 'PRIVATE',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_project_org
        FOREIGN KEY(organization_id)
        REFERENCES organizations(id)
        ON DELETE CASCADE

);

-- ============================================
-- ENVIRONMENTS
-- ============================================

CREATE TABLE environments (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    project_id UUID NOT NULL,

    name VARCHAR(50) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_environment_project
        FOREIGN KEY(project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,

    CONSTRAINT unique_environment
        UNIQUE(project_id,name)

);

-- ============================================
-- SERVICES
-- ============================================

CREATE TABLE services (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    environment_id UUID NOT NULL,

    name VARCHAR(120) NOT NULL,

    runtime VARCHAR(50) NOT NULL,

    repository_url TEXT,

    branch VARCHAR(120) DEFAULT 'main',

    dockerfile_path TEXT DEFAULT './Dockerfile',

    build_context TEXT DEFAULT '.',

    port INTEGER,

    health_check TEXT,

    cpu_limit VARCHAR(20),

    memory_limit VARCHAR(20),

    replicas INTEGER DEFAULT 1,

    auto_deploy BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_service_environment
        FOREIGN KEY(environment_id)
        REFERENCES environments(id)
        ON DELETE CASCADE

);

-- ============================================
-- DEPLOYMENTS
-- ============================================

CREATE TABLE deployments (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    service_id UUID NOT NULL,

    commit_sha VARCHAR(64),

    commit_message TEXT,

    deployed_by UUID,

    status deployment_status
        DEFAULT 'QUEUED',

    image_tag VARCHAR(255),

    started_at TIMESTAMP,

    finished_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_deployment_service
        FOREIGN KEY(service_id)
        REFERENCES services(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_deployment_user
        FOREIGN KEY(deployed_by)
        REFERENCES users(id)
        ON DELETE SET NULL

);

-- ============================================
-- BUILDS
-- ============================================

CREATE TABLE builds (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    deployment_id UUID NOT NULL,

    image_name VARCHAR(255),

    image_tag VARCHAR(255),

    duration INTEGER,

    logs TEXT,

    status deployment_status
        DEFAULT 'BUILDING',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_build_deployment
        FOREIGN KEY(deployment_id)
        REFERENCES deployments(id)
        ON DELETE CASCADE

);

-- ============================================
-- CONTAINERS
-- ============================================

CREATE TABLE containers (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    deployment_id UUID NOT NULL,

    docker_container_id VARCHAR(255) UNIQUE,

    hostname VARCHAR(255),

    ip_address INET,

    cpu_limit VARCHAR(20),

    memory_limit VARCHAR(20),

    status container_status
        DEFAULT 'CREATING',

    started_at TIMESTAMP,

    stopped_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_container_deployment
        FOREIGN KEY(deployment_id)
        REFERENCES deployments(id)
        ON DELETE CASCADE

);
-- ============================================
-- ENVIRONMENT VARIABLES
-- ============================================

CREATE TABLE environment_variables (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    service_id UUID NOT NULL,

    variable_key VARCHAR(255) NOT NULL,

    variable_value TEXT NOT NULL,

    is_secret BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_env_service
        FOREIGN KEY(service_id)
        REFERENCES services(id)
        ON DELETE CASCADE,

    CONSTRAINT unique_env_key
        UNIQUE(service_id, variable_key)

);

-- ============================================
-- DOMAINS
-- ============================================

CREATE TABLE domains (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    service_id UUID NOT NULL,

    domain VARCHAR(255) NOT NULL,

    is_primary BOOLEAN DEFAULT FALSE,

    verified BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_domain_service
        FOREIGN KEY(service_id)
        REFERENCES services(id)
        ON DELETE CASCADE,

    CONSTRAINT unique_domain
        UNIQUE(domain)

);

-- ============================================
-- SSL CERTIFICATES
-- ============================================

CREATE TABLE ssl_certificates (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    domain_id UUID NOT NULL,

    provider VARCHAR(100),

    certificate TEXT,

    private_key TEXT,

    expires_at TIMESTAMP,

    auto_renew BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_ssl_domain
        FOREIGN KEY(domain_id)
        REFERENCES domains(id)
        ON DELETE CASCADE

);

-- ============================================
-- GITHUB INSTALLATIONS
-- ============================================

CREATE TABLE github_installations (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    organization_id UUID NOT NULL,

    installation_id BIGINT NOT NULL,

    account_name VARCHAR(255),

    access_token TEXT,

    refresh_token TEXT,

    expires_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_installation_org
        FOREIGN KEY(organization_id)
        REFERENCES organizations(id)
        ON DELETE CASCADE,

    CONSTRAINT unique_installation
        UNIQUE(installation_id)

);

-- ============================================
-- GITHUB REPOSITORIES
-- ============================================

CREATE TABLE github_repositories (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    installation_id UUID NOT NULL,

    github_repository_id BIGINT NOT NULL,

    repository_name VARCHAR(255) NOT NULL,

    full_name VARCHAR(255) NOT NULL,

    default_branch VARCHAR(100),

    private BOOLEAN DEFAULT TRUE,

    clone_url TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_repo_installation
        FOREIGN KEY(installation_id)
        REFERENCES github_installations(id)
        ON DELETE CASCADE,

    CONSTRAINT unique_repository
        UNIQUE(github_repository_id)

);
-- ============================================
-- ACTIVITY LOGS
-- ============================================

CREATE TABLE activity_logs (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    organization_id UUID NOT NULL,

    user_id UUID,

    action VARCHAR(255) NOT NULL,

    resource_type VARCHAR(100),

    resource_id UUID,

    metadata JSONB,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_activity_org
        FOREIGN KEY(organization_id)
        REFERENCES organizations(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_activity_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE SET NULL

);

-- ============================================
-- AUDIT LOGS
-- ============================================

CREATE TABLE audit_logs (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID,

    resource VARCHAR(255) NOT NULL,

    action VARCHAR(255) NOT NULL,

    ip_address INET,

    user_agent TEXT,

    metadata JSONB,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_audit_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE SET NULL

);

-- ============================================
-- API KEYS
-- ============================================

CREATE TABLE api_keys (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    organization_id UUID NOT NULL,

    name VARCHAR(120) NOT NULL,

    hashed_key TEXT NOT NULL,

    last_used TIMESTAMP,

    expires_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_api_org
        FOREIGN KEY(organization_id)
        REFERENCES organizations(id)
        ON DELETE CASCADE

);

-- ============================================
-- METRICS
-- ============================================

CREATE TABLE metrics (

    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    container_id UUID NOT NULL,

    cpu_usage DECIMAL(5,2),

    memory_usage BIGINT,

    disk_usage BIGINT,

    network_in BIGINT,

    network_out BIGINT,

    recorded_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_metric_container
        FOREIGN KEY(container_id)
        REFERENCES containers(id)
        ON DELETE CASCADE

);