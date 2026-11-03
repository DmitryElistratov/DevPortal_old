package repository

import (
	"database/sql"
	"log"
)

// ApplyMigrations executes idempotent DDL statements at backend startup.
// docker-entrypoint-initdb.d only runs when the Postgres volume is created for
// the first time, so new feature tables must also be applied here with
// IF NOT EXISTS semantics.
func ApplyMigrations(db *sql.DB) {
	stmts := []string{
		// --- Core schema (mirrors 001_init.sql; docker-entrypoint-initdb.d only
		// runs on a fresh volume, so existing databases get these here). ---
		`CREATE TABLE IF NOT EXISTS users (
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
		)`,
		`CREATE TABLE IF NOT EXISTS services (
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
		)`,
		`CREATE TABLE IF NOT EXISTS releases (
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
		)`,
		`CREATE TABLE IF NOT EXISTS active_versions (
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
		)`,
		`CREATE TABLE IF NOT EXISTS canary_deployments (
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
		)`,
		`CREATE INDEX IF NOT EXISTS idx_canary_deployments_release_id ON canary_deployments(release_id)`,
		`CREATE INDEX IF NOT EXISTS idx_canary_deployments_service_id ON canary_deployments(service_id)`,
		`CREATE TABLE IF NOT EXISTS audit_log (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			user_name VARCHAR(255) NOT NULL,
			action VARCHAR(100) NOT NULL,
			resource VARCHAR(100) NOT NULL,
			resource_id VARCHAR(255),
			details TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS integrations (
			id VARCHAR(255) PRIMARY KEY,
			type VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			url VARCHAR(500) NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			config JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS project_permissions (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			service_id VARCHAR(255) NOT NULL REFERENCES services(id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL,
			granted_by VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id, service_id)
		)`,
		// Idempotent ALTERs for columns added after the original init migration.
		`ALTER TABLE releases ADD COLUMN IF NOT EXISTS git_commit VARCHAR(100)`,
		`ALTER TABLE releases ADD COLUMN IF NOT EXISTS is_primary BOOLEAN DEFAULT true`,
		`ALTER TABLE releases ADD COLUMN IF NOT EXISTS traffic_percentage INTEGER DEFAULT 100`,
		// --- TestOps feature tables ---
		`CREATE TABLE IF NOT EXISTS test_runs (
			id              TEXT PRIMARY KEY,
			service_id      TEXT NOT NULL DEFAULT '',
			release_id      TEXT NOT NULL DEFAULT '',
			environment     TEXT NOT NULL DEFAULT 'staging',
			suite           TEXT NOT NULL DEFAULT 'smoke',
			status          TEXT NOT NULL DEFAULT 'pending',
			total           INT  NOT NULL DEFAULT 0,
			passed          INT  NOT NULL DEFAULT 0,
			failed          INT  NOT NULL DEFAULT 0,
			skipped         INT  NOT NULL DEFAULT 0,
			duration_ms     BIGINT NOT NULL DEFAULT 0,
			report_url      TEXT NOT NULL DEFAULT '',
			pipeline_ref    TEXT NOT NULL DEFAULT '',
			triggered_by    TEXT NOT NULL DEFAULT '',
			details         TEXT NOT NULL DEFAULT '',
			deleted_at      TIMESTAMPTZ,
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_test_runs_release ON test_runs(release_id)`,
		`CREATE INDEX IF NOT EXISTS idx_test_runs_service ON test_runs(service_id)`,
		`CREATE TABLE IF NOT EXISTS checklists (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			steps       JSONB NOT NULL DEFAULT '[]'::jsonb,
			is_default  BOOLEAN NOT NULL DEFAULT FALSE,
			created_by  TEXT NOT NULL DEFAULT '',
			deleted_at  TIMESTAMPTZ,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS checklist_executions (
			id           TEXT PRIMARY KEY,
			checklist_id TEXT NOT NULL,
			release_id   TEXT NOT NULL DEFAULT '',
			steps        JSONB NOT NULL DEFAULT '[]'::jsonb,
			completed    BOOLEAN NOT NULL DEFAULT FALSE,
			created_by   TEXT NOT NULL DEFAULT '',
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_checklist_exec_release ON checklist_executions(release_id)`,
		`INSERT INTO checklists (id, name, description, steps, is_default, created_by)
		 VALUES ('cl-default-release', 'Стандартный релизный чек-лист',
		         'Базовые шаги проверки качества перед деплоем в production',
		         '[{"title":"Code review завершён, все MR смёржены"},{"title":"CI/CD пайплайн зелёный на ветке релиза"},{"title":"Смоук-тесты на staging пройдены"},{"title":"Тест-кеи регрессии выполнены (QA)"},{"title":"Метрики Grafana в норме после canary 5%"},{"title":"Логи Graylog без новых ошибок"},{"title":"Секреты и конфиги проверены (Vault)"},{"title":"Откат сценария протестирован"}]'::jsonb,
		         TRUE, 'system')
		 ON CONFLICT (id) DO NOTHING`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			log.Printf("[migrations] statement failed: %v", err)
		}
	}
}
