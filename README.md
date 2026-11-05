# Release Portal

Полноценный релизный портал на базе архитектуры Backstage (Spotify Developer Portal) для управления релизами, CI/CD пайплайнами и мониторингом качества.

## 🏗 Архитектура

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────┐
│   Frontend      │────▶│   Go Backend     │────▶│ PostgreSQL  │
│   React + TS    │     │   Gin Framework  │     │             │
│   Tailwind CSS  │     │   JWT Auth       │     └─────────────┘
└─────────────────┘     └──────────────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
   ┌────▼────┐          ┌─────▼─────┐         ┌─────▼─────┐
   │ GitLab  │          │   ArgoCD  │         │  Jira     │
   │ GitHub  │          │   K8s     │         │           │
   └─────────┘          └───────────┘         └───────────┘
        │                      │                      │
   ┌────▼────┐          ┌─────▼─────┐         ┌─────▼─────┐
   │ Grafana │          │  Graylog  │         │  Vault    │
   └─────────┘          └───────────┘         └───────────┘
```

## 🚀 Быстрый старт

### Требования
- Docker & Docker Compose
- Go 1.22+ (для локальной разработки backend)
- Node.js 22+ (для локальной разработки frontend)

### Запуск через Docker Compose

```bash
# 1. Скопируйте .env.example в .env
cp .env.example .env

# 2. Отредактируйте .env (JWT_SECRET и интеграции)
nano .env

# 3. Запустите все сервисы
docker-compose up -d

# 4. Проверьте статус
docker-compose ps

# 5. Откройте в браузере
open http://localhost:8080
```

### Первый вход

После запуска зарегистрируйте администратора:

```bash
curl -X POST http://localhost:8081/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@company.ru",
    "fullName": "Admin User",
    "password": "admin123",
    "role": "admin"
  }'
```

## 📁 Структура проекта

```
release-portal/
├── backend/                    # Go backend
│   ├── cmd/api/main.go        # Точка входа
│   ├── internal/
│   │   ├── config/            # Конфигурация
│   │   ├── handler/           # HTTP handlers
│   │   ├── service/           # Бизнес-логика
│   │   ├── repository/        # Работа с БД
│   │   ├── integration/       # Клиенты внешних систем
│   │   └── middleware/        # Auth, CORS, RBAC
│   ├── migrations/            # SQL миграции
│   ├── Dockerfile
│   └── go.mod
├── src/                        # React frontend
│   ├── components/
│   ├── contexts/
│   ├── pages/
│   ├── lib/api.ts             # API клиент
│   └── App.tsx
├── docker-compose.yml
├── Dockerfile                 # Frontend Dockerfile
├── nginx.conf
└── .env.example
```

## 🔧 API Endpoints

### Authentication
- `POST /api/auth/login` — Вход
- `POST /api/auth/register` — Регистрация
- `POST /api/auth/oauth` — Вход через OAuth-провайдера (обмен профиля внешнего IdP на локальный JWT)
- `GET /api/auth-providers` — Публичный список включённых OIDC-провайдеров (для кнопок входа; секреты не раскрываются)
- `GET /api/me` — Текущий пользователь
- `GET /api/users` — Список пользователей (admin)

### Services
- `GET /api/services` — Список сервисов
- `GET /api/services/:id` — Детали сервиса
- `POST /api/services` — Создать сервис (admin)
- `PUT /api/services/:id` — Обновить (admin)
- `DELETE /api/services/:id` — Пометить на удаление (soft-delete, требует подтверждения вторым человеком)

### Releases
- `GET /api/releases` — Список релизов
- `GET /api/releases/:id` — Детали релиза (включая security/tests статусы)
- `POST /api/releases` — Создать релиз (с указанием Git-ветки)
- `PUT /api/releases/:id` — Обновить релиз
- `POST /api/releases/:id/deploy` — Деплой
- `POST /api/releases/:id/rollback` — Откат
- `POST /api/releases/:id/approve` — Одобрить
- `POST /api/releases/:id/security-scan` — Подтянуть отчёт безопасности из GitLab CI
- `POST /api/releases/:id/tests` — Подтянуть результаты тестов из GitLab CI (JUnit/coverage)

### Active Versions
- `POST /api/active-versions` — Задеплоить новую версию
- `GET /api/active-versions/service/:serviceId` — Получить активные версии
- `PUT /api/active-versions/:versionId/traffic` — Обновить процент трафика
- `POST /api/active-versions/:versionId/rollback` — Откатить версию
- `POST /api/active-versions/:versionId/promote` — Продвинуть версию до 100%

### Canary Deployments
- `POST /api/canary` — Создать canary deployment
- `GET /api/canary` — Все canary-деплои (активные + история)
- `GET /api/canary/:id` — Детали canary
- `GET /api/canary/release/:releaseId` — Canary по релизу
- `GET /api/canary/service/:serviceId/active` — Активный canary
- `GET /api/canary/:id/analyze` — Живой статус из ArgoCD (sync/health/traffic split)
- `PUT /api/canary/:id/percentage` — Обновить процент трафика
- `POST /api/canary/:id/promote` — Продвинуть canary
- `POST /api/canary/:id/rollback` — Откатить canary

### Git / CI/CD
- `GET /api/ci/branches?serviceId=<id>` — Реальные ветки Git-репозитория сервиса (GitLab)
- `GET /api/ci/tags?serviceId=<id>` — Теги репозитория сервиса
- Триггер пайплайнов выбранной ветки, статусы MR и pipeline — через handlers `ci_handler.go`

### ArgoCD / Deployments (строго GitOps)
- `GET /api/argocd/applications` — Список приложений подключённого ArgoCD (live sync/health)
- `GET /api/argocd/applications/:name` — Детали приложения
- `POST /api/argocd/applications/:name/sync` — Синхронизация приложения
- `GET /api/deployments/:id` — Результат GitOps-деплоя (коммит манифестов → ArgoCD sync → health-check)
- Деплой выполняется коммитом в gitops-репозиторий и триггером ArgoCD; прямых write-операций в кластер нет

### Vault Secrets
- `GET /api/vault/secrets?path=<p>` — Список секретов с данными (KV v1/v2, автодетект mount)
- `GET /api/vault/secrets/*path` — Читать секрет (любая глубина вложенности)
- `POST /api/vault/secrets/*path` — Создать секрет
- `PUT /api/vault/secrets/*path` — Обновить секрет
- `DELETE /api/vault/secrets/*path` — Удалить секрет (soft-delete workflow для каталога портала)
- `GET /api/vault/secrets/*path/versions` — История версий (KV v2)
- `POST /api/vault/secrets/*path/restore` — Восстановить версию

### TestOps / QA
- `GET/POST/PUT/DELETE /api/testruns` — Тест-раны (CRUD)
- `GET/POST/PUT/DELETE /api/checklists` — Чек-листы релизов (CRUD, soft-delete)

### Integrations
- `GET /api/integrations` — Список интеграций
- `GET /api/integrations/:id` — Детали интеграции
- `POST /api/integrations` — Создать интеграцию (admin)
- `PUT /api/integrations/:id` — Обновить (admin; частичное обновление не затирает сохранённые токены)
- `DELETE /api/integrations/:id` — Деактивировать/пометить на удаление (soft-delete, требует подтверждения)
- `GET /api/integrations/type/:type` — Интеграции по типу
- `POST /api/integrations/:id/test` — Тестировать соединение
- `POST /api/integrations/:id/sync` — Синхронизация (GitLab/GitHub/Jira/ArgoCD/K8s сохраняют сущности в каталог `services`, GitLab дополнительно импортирует пользователей и права проектов)

### Pending Deletions (4-eyes контроль)
- `GET /api/pending-deletions` — Очередь запросов на удаление
- `POST /api/pending-deletions/:id/approve|reject` — Подтверждение/отклонение вторым администратором

### Audit
- `GET /api/audit` — Журнал действий

## 🔐 Аутентификация и авторизация

### Локальная аутентификация
- JWT токены с HMAC-SHA256
- Регистрация и вход по username/password
- Хеширование паролей bcrypt

### Внешние провайдеры (OAuth2/OIDC/LDAP)

#### Keycloak (OIDC)
```bash
KEYCLOAK_URL=http://localhost:8082
KEYCLOAK_REALM=release-portal
KEYCLOAK_CLIENT_ID=release-portal
KEYCLOAK_CLIENT_SECRET=your-secret
```

#### LDAP
```bash
LDAP_URL=ldap://ldap.company.ru:389
LDAP_BASE_DN=dc=company,dc=ru
LDAP_BIND_DN=cn=admin,dc=company,dc=ru
LDAP_BIND_PASS=your-password
```

#### GitHub OAuth
```bash
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret
GITHUB_REDIRECT_URI=http://localhost:8080/auth/callback/github
```

#### GitLab OAuth
```bash
GITLAB_CLIENT_ID=your-client-id
GITLAB_CLIENT_SECRET=your-client-secret
GITLAB_REDIRECT_URI=http://localhost:8080/auth/callback/gitlab
```

### Маппинг групп на роли
Автоматический маппинг групп из IdP на роли приложения:
- `admins`, `operations` → `admin`
- `developers`, `engineering` → `developer`
- `qa`, `testing` → `qa`
- Все остальные → `viewer`

### Проектные права (RBAC на уровне проектов)
- Назначение ролей пользователям для конкретных сервисов
- Роли: `admin`, `developer`, `viewer`

### Роли
- `admin` — полный доступ ко всем сервисам
- `developer` — доступ к назначенным проектам
- `qa` — доступ к тестированию
- `viewer` — только просмотр

## 🔌 Интеграции

### GitLab / GitHub
- Список проектов и репозиториев
- Merge Requests
- CI/CD пайплайны
- Триггер пайплайнов

### Jira
- Получение задач (issues)
- Поиск по JQL
- Транзишн статусов
- Синхронизация с релизами

### ArgoCD
- Список приложений
- Статус синхронизации и здоровья
- Ручной sync приложений
- Мониторинг деплоев
- Canary deployment (traffic split через JSON-patch Rollouts)

> ⚠️ **ВАЖНО — MOCK ArgoCD для тестов!**
> В `docker-compose.yml` добавлен сервис `argocd` (`scripts/argocd-mock/server.js`) —
> это заглушка, эмулирующая REST API ArgoCD для локальной разработки без K8s-кластера.
> **При окончательной (production) сборке проекта этот мок необходимо удалить:**
> 1. Убрать сервис `argocd:` из `docker-compose.yml` (и каталог `scripts/argocd-mock`).
> 2. Развернуть реальный ArgoCD (в K8s-кластере) и в портале (Settings → Integrations)
>    указать URL настоящего ArgoCD API server, затем выполнить Test Connection + Sync.
> 3. Аналогично проверить, что все остальные интеграции (GitLab, Jira, Vault, Grafana,
>    Graylog, Kubernetes) переключены на реальные инстансы, а не локальные dev-контейнеры.

### Kubernetes
- Pods, Deployments, Services
- Логи подов
- События кластера
- Статус ресурсов

### Grafana
- Список дашбордов
- Datasources
- Prometheus запросы
- Embed дашбордов

### Graylog
- Поиск логов
- Streams
- Фильтрация по сервисам

### HashiCorp Vault
- Хранение секретов интеграций и пользовательских секретов портала
- KV v1 **и** v2: автоопределение mount/версии движка через `sys/mounts` (read-only LIST-пробы, тестовые ключи в хранилище не создаются), самовосстановление после перезапуска dev-Vault
- Древовидный браузер в UI: раскрытие папок по клику, breadcrumbs + «Наверх», поиск по пути
- Полный CRUD секретов произвольной глубины (`secret/a/b/c`), история версий и восстановление (KV v2)
- Криптографический генератор паролей на клиенте (WebCrypto `getRandomValues`, unbiased rejection sampling, индикатор энтропии)
- API endpoints:
  - `GetSecret(path)` — получить секрет
  - `WriteSecret(path, data)` — записать секрет
  - `DeleteSecret(path)` — удалить секрет
  - `ListSecrets(path)` — список секретов (LIST-глагол Vault)
  - `ListVersions(path)` / `RestoreVersion(path, version)` — версии KV v2

### TestOps / QA
- Тест-раны (`test_runs`): запуск, статусы passed/failed/skipped, покрытие
- Релизные чек-листы (`checklists`) с soft-delete
- Интеграция с CI/CD GitLab: пайплайны, MR, триггер по выбранной ветке
- Оценка безопасности релиза: SAST/DAST/Dependency-Scanning отчёты GitLab → блок Security Assessment на странице релиза
- Результаты тестов: JUnit XML + coverage из логов pipeline → блок Test Results

## 🎨 Frontend

### Технологии
- React 18 + TypeScript
- Tailwind CSS (чёрно-бело-оранжевая тема)
- React Router (HashRouter)
- Lucide React (иконки)
- Vite 6 (сборка)

### Страницы
- **Home** — дашборд с общей статистикой (в т.ч. компактные бейджи security/tests по релизам)
- **Catalog** — каталог сервисов (синхронизируется из GitLab/GitHub/Jira/ArgoCD/K8s)
- **Component** — детали сервиса (Overview, Releases, CI/CD, API, Observability, Active Versions, деплой выбранной ветки)
- **Releases / Release** — список и детали релиза: статус, версии, автор, коммит, блоки **Security Assessment** и **Test Results** (из GitLab CI), действия Deploy / Approve / Rollback / Canary
- **Create Release** — создание релиза с выбором Git-ветки (реальный список веток подтягивается из GitLab)
- **Active Versions** — управление активными версиями и распределением трафика
- **Canary Deployments** — полный цикл canary через ArgoCD Rollouts: выбор приложения из списка ArgoCD, поэтапное повышение трафика 5→100%, Promote/Rollback, live-анализ health/sync
- **Deployments** — GitOps-деплои через ArgoCD (sync/health, кнопка Sync, фильтры; результат деплоя — модалка стадий GitOps-пайплайна)
- **Tests / TestOps** — тест-раны и интеграция с CI/CD пайплайнами GitLab
- **Plugins** — обзор плагинов/интеграций портала
- **Kubernetes API** — вспомогательное чтение состояния кластера (логи, события; write-операции вынесены в ArgoCD)
- **Observability** — Grafana метрики + Graylog логи
- **Settings** — управление интеграциями (Git, Jira, ArgoCD, Kubernetes, Grafana, Graylog, Vault), **auth-интеграции** (Keycloak/GitLab/GitHub/LDAP) и список пользователей из БД
- **Vault Secrets** — древовидный браузер секретов: ленивое раскрытие папок по клику, breadcrumbs + «Назад», поиск по пути, CRUD с историей версий KV v2, криптографический генератор паролей (WebCrypto)
- **Login** — локальный вход + кнопки OAuth («Войти через GitLab/GitHub/Keycloak») для настроенных провайдеров
- **Templates** — шаблоны создания сервисов
- **TechDocs** — документация
- **Checklists** — QA-чек-листы для релизов
- **Time to Market** — метрики скорости доставки
- **Audit Log** — журнал действий
- **Pending Deletions** — очередь запросов на удаление (подтверждение вторым администратором)

## 🛡 Безопасность данных

- **Soft-delete**: портал никогда не удаляет записи физически — объекты помечаются `deleted_at` / статусом «на удаление».
- **4-eyes**: любое удаление (сервис, релиз, интеграция, запись каталога) создаёт запись в `pending_deletions` и требует подтверждения **вторым** администратором.
- Интеграции с внешними системами при синхронизации только деактивируют objects, удалённые в источнике, но не удаляют их сами.
- Секреты интеграций (токены, client secret, bind password) маскируются в UI и не отдаются наружу (`/api/auth-providers` отдаёт только name/provider/url/clientId/realm).
- Пароли — bcrypt; сессии — JWT HMAC-SHA256; RBAC: admin / developer / qa / viewer + проектные права (`project_permissions`, импортируются из GitLab: Owner→admin, Maintainer→maintainer, Developer→developer).

## 🗄 База данных

PostgreSQL 16 с таблицами:
- `users` — пользователи (в т.ч. импортированные из GitLab)
- `services` — сервисы каталога (синхронизируются из GitLab/GitHub/Jira/ArgoCD/K8s)
- `releases` — релизы (колонки `branch`, `security_*`, `tests_*`)
- `active_versions` — активные версии
- `canary_deployments` — canary деплои
- `test_runs` — тест-раны TestOps
- `checklists` — QA-чек-листы
- `audit_log` — журнал аудита
- `integrations` — настройки интеграций (config хранит credentials, merge при частичном обновлении)
- `project_permissions` — проектные права
- `pending_deletions` — очередь подтверждений удалений (4-eyes)

Миграции применяются автоматически при старте бэкенда (`ApplyMigrations`, идемпотентно через `IF NOT EXISTS`) — старые тома Postgres добирают недостающие таблицы/колонки без пересоздания.

## 🛠 Локальная разработка

### Backend
```bash
cd backend
go mod tidy
go run cmd/api/main.go
```

### Frontend
```bash
npm install
npm run dev
```

### Сборка
```bash
# Frontend
npm run build

# Backend
cd backend && go build -o main cmd/api/main.go
```

## ☸️ Деплой в Kubernetes (production release)

Для продакшена используется **только** стек: frontend + backend + PostgreSQL. Никаких моков и dev-заглушек (ArgoCD mock, Keycloak, GitLab CE из `docker-compose.yml` — исключительно для локальной разработки, в релиз не попадают).

### Сборка поименованного релиза в tar-архив
```bash
# версия передаётся аргументом или через VERSION; по умолчанию v1.0.0-<gitsha>
REGISTRY=registry.example.com/team PUSH=true ./release/build-release.sh v1.1.0
# без реестра (air-gapped / minikube/kind): просто
./release/build-release.sh v1.1.0
```
Результат: `release/release-portal-v<VERSION>.tar.gz`, внутри:
- `images/backend.tar`, `images/frontend.tar` — готовые production-образы (`docker load`);
- `k8s/` — манифесты (namespace, postgres StatefulSet+PVC+Secret, backend, frontend+Ingress, kustomization), имена образов уже подставлены сборщиком;
- `install.sh` — загрузка образов + `kubectl apply -k` + ожидание rollout;
- `RELEASE_INFO.txt` — версия, git commit, дата сборки, список образов.

### Установка в кластер
```bash
tar -xzf release/release-portal-v1.1.0.tar.gz && cd release-portal-v1.1.0
# 1. отредактируйте секреты: k8s/postgres.yaml (POSTGRES_PASSWORD/DATABASE_URL),
#    k8s/backend.yaml (JWT_SECRET), хост в Ingress (k8s/frontend.yaml)
./install.sh          # docker load + kubectl apply -k k8s/
```
Без своего образа кластера: `minikube image load images/*.tar` или `kind load image-archive`. Миграции БД применяются автоматически при старте бэкенда (идемпотентно). Фронтенд ходит в API через относительный путь `/api` (nginx проксирует на сервис `portal-backend` в кластере, CORS не нужен); внешние интеграции (GitLab/ArgoCD/Vault/Jira) настраиваются через UI Settings на реальные инстансы.

Прямой деплой манифестов из репозитория (без архива): `kubectl apply -k k8s/` (перед этим замените плейсхолдеры тегов образов `v1.0.0` и секретов).

## 📊 Мониторинг

- Health check: `GET /health`
- Backend: `http://localhost:8081/health` (в кластере — `portal-backend:8080/health`)
- Frontend: `http://localhost:8080/health`

## ✅ Чек-лист перед Pull Request / слиянием в main

1. **Сборка и тесты** (всё должно пройти без ошибок):
   ```bash
   # Backend
   cd backend && go build ./... && go vet ./... && go test ./...
   # Frontend
   npm ci && npx tsc --noEmit && npm run build
   ```
2. **Удалить мок ArgoCD** (см. предупреждение в разделе «Интеграции → ArgoCD» выше): убрать сервис `argocd:` из `docker-compose.yml`, удалить каталог `scripts/argocd-mock/`, переключить интеграцию в Settings на реальный ArgoCD API server.
3. **Проверить, что в diff нет артефактов**: `node_modules/`, `dist/`, бинарник `backend/api`, `.env` — всё в `.gitignore`, коммитить нельзя.
4. **Секреты**: убедиться, что в коде/конфигах/миграциях нет реальных токенов, паролей и JWT_SECRET (только placeholder'ы вида `your-secret`).
5. **Миграции**: новые таблицы/колонки добавлялись идемпотентно (`IF NOT EXISTS`) — проверить, что старт бэкенда на старой БД проходит без `[migrations] statement failed`.
6. **Smoke-тест после `docker compose up -d`**: логин → Settings (Test Connection + Sync по каждой интеграции) → Services (каталог не пуст) → Create Release (ветки подтянулись) → Deployments (ArgoCD sync) → Canary (поднятие трафика/promote) → Vault (дерево секретов, CRUD) → Audit Log.

## 🔀 Как оформить Pull Request

```bash
# 1. Ветка от актуального main
git checkout main && git pull
git checkout -b feature/<название-задачи>

# 2. Изменения + коммиты в конвенциональном формате
git add <files>
git commit -m "feat(scope): краткое описание"   # feat|fix|refactor|chore|docs|test

# 3. Push и создание MR/PR
git push -u origin feature/<название-задачи>
# GitLab: create merge request; GitHub: gh pr create
```

В описании PR укажите: что изменено, какие проверки выполнены (пункты 1–6 чек-листа), затронутые миграции и требования к конфигурации окружения (новые env-переменные, настройки OAuth redirect URI и т.п.).

## 📝 Лицензия

MIT
