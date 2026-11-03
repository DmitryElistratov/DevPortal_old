-- TestOps / QA functionality (README: "qa — доступ к тестированию", CI/CD пайплайны,
-- Merge Requests, Checklists — чек-листы для релизов)

CREATE TABLE IF NOT EXISTS test_runs (
    id              TEXT PRIMARY KEY,
    service_id      TEXT NOT NULL DEFAULT '',
    release_id      TEXT NOT NULL DEFAULT '',
    environment     TEXT NOT NULL DEFAULT 'staging',
    suite           TEXT NOT NULL DEFAULT 'smoke',
    status          TEXT NOT NULL DEFAULT 'pending',   -- pending|running|passed|failed|aborted
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
);

CREATE INDEX IF NOT EXISTS idx_test_runs_release ON test_runs(release_id);
CREATE INDEX IF NOT EXISTS idx_test_runs_service ON test_runs(service_id);

CREATE TABLE IF NOT EXISTS checklists (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    steps       JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    created_by  TEXT NOT NULL DEFAULT '',
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS checklist_executions (
    id           TEXT PRIMARY KEY,
    checklist_id TEXT NOT NULL,
    release_id   TEXT NOT NULL DEFAULT '',
    steps        JSONB NOT NULL DEFAULT '[]'::jsonb,  -- [{id,title,checked,checkedBy,checkedAt}]
    completed    BOOLEAN NOT NULL DEFAULT FALSE,
    created_by   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_checklist_exec_release ON checklist_executions(release_id);

-- Seed default release checklist (releasenый чек-лист по умолчанию)
INSERT INTO checklists (id, name, description, steps, is_default, created_by)
VALUES ('cl-default-release', 'Стандартный релизный чек-лист',
        'Базовые шаги проверки качества перед деплоем в production',
        '[{"title":"Code review завершён, все MR смёржены"},{"title":"CI/CD пайплайн зелёный на ветке релиза"},{"title":"Смоук-тесты на staging пройдены"},{"title":"Тест-кеи регрессии выполнены (QA)"},{"title":"Метрики Grafana в норме после canary 5%"},{"title":"Логи Graylog без новых ошибок"},{"title":"Секреты и конфиги проверены (Vault)"},{"title":"Откат сценария протестирован"}]'::jsonb,
        TRUE, 'system')
ON CONFLICT (id) DO NOTHING;
