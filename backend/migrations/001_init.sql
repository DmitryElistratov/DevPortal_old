-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'developer',
    auth_provider VARCHAR(50) DEFAULT 'local',
    external_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Services table
CREATE TABLE IF NOT EXISTS services (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    team VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'service',
    lifecycle VARCHAR(50) NOT NULL DEFAULT 'production',
    description TEXT,
    language VARCHAR(100),
    repository VARCHAR(500),
    owner VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Releases table
CREATE TABLE IF NOT EXISTS releases (
    id VARCHAR(255) PRIMARY KEY,
    service_id VARCHAR(255) NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    version VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'building',
    description TEXT,
    branch VARCHAR(255),
    git_commit VARCHAR(100),
    created_by VARCHAR(255) NOT NULL,
    approvals INTEGER NOT NULL DEFAULT 0,
    required_approvals INTEGER NOT NULL DEFAULT 2,
    is_primary BOOLEAN DEFAULT true,
    traffic_percentage INTEGER DEFAULT 100,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Active versions table
CREATE TABLE IF NOT EXISTS active_versions (
    id VARCHAR(255) PRIMARY KEY,
    service_id VARCHAR(255) NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    release_id VARCHAR(255) NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    version VARCHAR(100) NOT NULL,
    branch VARCHAR(255) NOT NULL,
    git_commit VARCHAR(100),
    environment VARCHAR(50) NOT NULL,
    traffic_percentage INTEGER NOT NULL DEFAULT 100,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    deployed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deployed_by VARCHAR(255) NOT NULL
);

-- Canary deployments table
CREATE TABLE IF NOT EXISTS canary_deployments (
    id VARCHAR(255) PRIMARY KEY,
    release_id VARCHAR(255) NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    service_id VARCHAR(255) NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    percentage INTEGER NOT NULL DEFAULT 10,
    stable_version VARCHAR(100) NOT NULL,
    canary_version VARCHAR(100) NOT NULL,
    stable_revision VARCHAR(100),
    canary_revision VARCHAR(100),
    argo_app VARCHAR(255) NOT NULL,
    git_tag VARCHAR(255),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit log table
CREATE TABLE IF NOT EXISTS audit_log (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    details TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Integrations table
CREATE TABLE IF NOT EXISTS integrations (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(500) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Project permissions table
CREATE TABLE IF NOT EXISTS project_permissions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_id VARCHAR(255) NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    granted_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, service_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_releases_service_id ON releases(service_id);
CREATE INDEX IF NOT EXISTS idx_releases_status ON releases(status);
CREATE INDEX IF NOT EXISTS idx_active_versions_service_id ON active_versions(service_id);
CREATE INDEX IF NOT EXISTS idx_active_versions_environment ON active_versions(environment);
CREATE INDEX IF NOT EXISTS idx_canary_deployments_release_id ON canary_deployments(release_id);
CREATE INDEX IF NOT EXISTS idx_canary_deployments_service_id ON canary_deployments(service_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_integrations_type ON integrations(type);
CREATE INDEX IF NOT EXISTS idx_project_permissions_user_id ON project_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_project_permissions_service_id ON project_permissions(service_id);
