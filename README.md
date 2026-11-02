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
- `GET /api/me` — Текущий пользователь

### Services
- `GET /api/services` — Список сервисов
- `GET /api/services/:id` — Детали сервиса
- `POST /api/services` — Создать сервис (admin)
- `PUT /api/services/:id` — Обновить (admin)
- `DELETE /api/services/:id` — Удалить (admin)

### Releases
- `GET /api/releases` — Список релизов
- `GET /api/releases/:id` — Детали релиза
- `POST /api/releases` — Создать релиз
- `PUT /api/releases/:id` — Обновить релиз
- `POST /api/releases/:id/deploy` — Деплой
- `POST /api/releases/:id/rollback` — Откат
- `POST /api/releases/:id/approve` — Одобрить

### Active Versions
- `POST /api/active-versions` — Задеплоить новую версию
- `GET /api/active-versions/service/:serviceId` — Получить активные версии
- `PUT /api/active-versions/:versionId/traffic` — Обновить процент трафика
- `POST /api/active-versions/:versionId/rollback` — Откатить версию
- `POST /api/active-versions/:versionId/promote` — Продвинуть версию до 100%

### Canary Deployments
- `POST /api/canary` — Создать canary deployment
- `GET /api/canary/:id` — Детали canary
- `GET /api/canary/release/:releaseId` — Canary по релизу
- `GET /api/canary/service/:serviceId/active` — Активный canary
- `PUT /api/canary/:id/percentage` — Обновить процент трафика
- `POST /api/canary/:id/promote` — Продвинуть canary
- `POST /api/canary/:id/rollback` — Откатить canary

### Integrations
- `GET /api/integrations` — Список интеграций
- `GET /api/integrations/:id` — Детали интеграции
- `POST /api/integrations` — Создать интеграцию (admin)
- `PUT /api/integrations/:id` — Обновить (admin)
- `DELETE /api/integrations/:id` — Удалить (admin)
- `GET /api/integrations/type/:type` — Интеграции по типу
- `POST /api/integrations/:id/test` — Тестировать соединение

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
- Хранение секретов интеграций
- KV v2 secrets engine
- Получение credentials для интеграций
- API endpoints:
  - `GetSecret(path)` — получить секрет
  - `WriteSecret(path, data)` — записать секрет
  - `DeleteSecret(path)` — удалить секрет
  - `ListSecrets(path)` — список секретов

## 🎨 Frontend

### Технологии
- React 18 + TypeScript
- Tailwind CSS (чёрно-бело-оранжевая тема)
- React Router (HashRouter)
- Recharts (графики)
- Lucide React (иконки)

### Страницы
- **Home** — дашборд с общей статистикой
- **Catalog** — каталог сервисов
- **Component** — детали сервиса (Overview, Releases, CI/CD, API, Observability, Active Versions)
- **Release** — детали релиза с пайплайнами
- **Create Release** — создание релиза с выбором Git-ветки
- **Active Versions** — управление активными версиями и трафиком
- **Deployments** — ArgoCD приложения
- **Kubernetes API** — прямое взаимодействие с K8s
- **Observability** — Grafana метрики + Graylog логи
- **Plugins** — Backstage плагины
- **Settings** — управление интеграциями (Git, Jira, ArgoCD, Kubernetes, Grafana, Graylog, Vault)
- **Vault Secrets** — управление секретами в Vault
- **Templates** — шаблоны создания сервисов
- **TechDocs** — документация
- **Checklists** — чек-листы для релизов
- **Time to Market** — метрики скорости доставки
- **Audit Log** — журнал действий

## 🗄 База данных

PostgreSQL 16 с таблицами:
- `users` — пользователи
- `services` — сервисы
- `releases` — релизы
- `active_versions` — активные версии
- `canary_deployments` — canary деплои
- `audit_log` — журнал аудита
- `integrations` — настройки интеграций
- `project_permissions` — проектные права

Миграции автоматически применяются при первом запуске PostgreSQL.

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

## 📊 Мониторинг

- Health check: `GET /health`
- Backend: `http://localhost:8081/health`
- Frontend: `http://localhost:8080/health`

## 📝 Лицензия

MIT
