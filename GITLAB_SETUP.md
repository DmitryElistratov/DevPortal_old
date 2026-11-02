# 🦊 Настройка GitLab интеграции

**Дата:** 2024  
**Статус:** ✅ GitLab добавлен в docker-compose

---

## 📋 Что было добавлено

GitLab CE (Community Edition) добавлен в `docker-compose.yml`:

```yaml
gitlab:
  image: gitlab/gitlab-ce:16.8.0-ce.0
  hostname: gitlab.local
  ports:
    - "8083:80"      # HTTP
    - "8084:443"     # HTTPS
    - "2222:22"      # SSH
  volumes:
    - gitlab_config:/etc/gitlab
    - gitlab_logs:/var/log/gitlab
    - gitlab_data:/var/opt/gitlab
```

---

## 🚀 Запуск GitLab

### 1. Запустите все сервисы

```bash
docker-compose up -d
```

### 2. Дождитесь запуска GitLab

GitLab требует **5-10 минут** для первого запуска. Проверьте статус:

```bash
docker-compose logs -f gitlab
```

**Ожидаемый вывод:**
```
gitlab_1  | ==> /var/log/gitlab/gitlab-rails/application.log <==
gitlab_1  | GitLab is ready!
```

Или проверьте health:
```bash
docker-compose ps
```

Должно показать: `Up (healthy)`

### 3. Откройте GitLab UI

```
http://localhost:8083
```

---

## 🔑 Получение root пароля

### Автоматический пароль (GitLab 15+)

GitLab автоматически генерирует root пароль. Получите его:

```bash
docker-compose exec gitlab cat /etc/gitlab/initial_root_password
```

**Вывод:**
```
# WARNING: This value will be reset after 24 hours.
# Please save it for future use.
Password: abc123XYZ456...
```

**Сохраните этот пароль!** Он будет удалён через 24 часа.

### Вход в GitLab

1. Откройте http://localhost:8083
2. Username: `root`
3. Password: (из файла выше)
4. Нажмите "Sign in"
5. Смените пароль при первом входе

---

## 🎫 Создание API токена

### 1. Войдите как root

### 2. Перейдите в настройки токенов

```
http://localhost:8083/-/profile/personal_access_tokens
```

Или через меню:
- Click on your avatar (top-right)
- Edit profile
- Access Tokens

### 3. Создайте новый токен

**Заполните форму:**
- **Token name:** `release-portal`
- **Expiration date:** (опционально, или оставьте пустым)
- **Scopes:** 
  - ✅ `api` (полный доступ к API)
  - ✅ `read_user` (чтение информации о пользователях)
  - ✅ `read_repository` (чтение репозиториев)
  - ✅ `write_repository` (запись в репозитории)

### 4. Нажмите "Create personal access token"

### 5. Скопируйте токен

**ВАЖНО:** Токен показывается только один раз! Скопируйте его сразу.

**Пример токена:**
```
glpat-xxxxxxxxxxxxxxxxxxxx
```

---

## 🔧 Настройка интеграции в Release Portal

### 1. Откройте Release Portal

```
http://localhost:8080
```

### 2. Войдите как администратор

### 3. Перейдите в Settings → Git

Левое меню → **Settings** → вкладка **Git**

### 4. Добавьте новую интеграцию

Нажмите **"Добавить"** и заполните форму:

- **Название:** `Main GitLab`
- **Тип:** `GitLab`
- **URL:** `http://gitlab:80` (внутренний Docker network)
- **Access Token:** `glpat-xxxxxxxxxxxxxxxxxxxx` (из шага выше)

### 5. Нажмите "Сохранить"

### 6. Проверьте соединение

Нажмите кнопку **"Проверить соединение"** (иконка пробирки).

**Ожидаемый результат:**
```
✅ Соединение успешно!
```

---

## 📦 Создание тестового проекта в GitLab

### 1. Создайте новый проект

В GitLab UI:
- Click "Create new project"
- Click "Create blank project"
- **Project name:** `payment-service`
- **Visibility:** `Private`
- Click "Create project"

### 2. Добавьте файлы

```bash
# Клонируйте проект
git clone http://localhost:8083/root/payment-service.git
cd payment-service

# Создайте файлы
echo "# Payment Service" > README.md
mkdir -p src
cat > src/main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Payment Service v1.0.0")
}
EOF

# Создайте ветки
git checkout -b feature/new-payment
git add .
git commit -m "Initial commit"
git push origin feature/new-payment

git checkout -b release/1.0.0
git push origin release/1.0.0

git checkout main
git push origin main
```

### 3. Создайте теги

```bash
git tag v1.0.0
git push origin v1.0.0

git tag v1.1.0
git push origin v1.1.0
```

---

## ✅ Проверка интеграции

### 1. Проверьте список веток

В Release Portal:
- Откройте **Create Release**
- Выберите сервис `payment-service`
- В выпадающем списке "Git Branch" должны появиться:
  - ✅ `main`
  - ✅ `feature/new-payment`
  - ✅ `release/1.0.0`

### 2. Проверьте список тегов

В GitLab UI:
- Откройте проект `payment-service`
- Перейдите в **Repository → Tags**
- Должны появиться:
  - ✅ `v1.0.0`
  - ✅ `v1.1.0`

### 3. Создайте canary deployment

В Release Portal:
1. Откройте Release со статусом `testing`
2. Нажмите **"Deploy as Canary"**
3. Выберите ArgoCD приложение
4. Установите 10% трафика
5. Нажмите **"Start Canary Deployment"**

### 4. Проверьте логи backend

```bash
docker-compose logs -f backend | grep GitLab
```

**Ожидаемый вывод:**
```
[Canary] Creating Git tag: canary-2.14.0-20240115-120000
[Integration] GitLab client initialized: http://gitlab:80
[Canary] Successfully created Git tag
```

### 5. Проверьте GitLab UI

Откройте проект `payment-service` в GitLab:
- Перейдите в **Repository → Tags**
- Должен появиться новый тег:
  - ✅ `canary-2.14.0-20240115-120000`

---

## 🔍 Troubleshooting

### Проблема: GitLab не запускается

**Симптом:**
```
docker-compose ps
gitlab_1  | (exited with code 1)
```

**Решение:**
```bash
# Увеличьте память Docker
# GitLab требует минимум 4GB RAM

# Проверьте логи
docker-compose logs gitlab

# Перезапустите
docker-compose restart gitlab
```

### Проблема: Не могу получить root пароль

**Симптом:**
```bash
docker-compose exec gitlab cat /etc/gitlab/initial_root_password
cat: /etc/gitlab/initial_root_password: No such file or directory
```

**Решение:**
```bash
# Сбросьте root пароль
docker-compose exec gitlab gitlab-rake "gitlab:password:reset[root]"

# Введите новый пароль
# Подтвердите пароль
```

### Проблема: API токен не работает

**Симптом:**
```
❌ Соединение не удалось: 401 Unauthorized
```

**Решение:**
1. Проверьте что токен не истёк
2. Проверьте что у токена есть scope `api`
3. Создайте новый токен

### Проблема: Ветки не появляются в Release Portal

**Симптом:**
В выпадающем списке "Git Branch" нет веток из GitLab.

**Решение:**
1. Проверьте что интеграция настроена правильно
2. Проверьте что проект существует в GitLab
3. Проверьте логи backend:
   ```bash
   docker-compose logs backend | grep GitLab
   ```
4. Проверьте что у токена есть доступ к проекту

---

## 📊 Архитектура интеграции

```
┌─────────────────┐
│  Release Portal │
│   (Frontend)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Backend      │
│   (Go + Gin)    │
└────────┬────────┘
         │
         ├──────────────┬──────────────┐
         │              │              │
         ▼              ▼              ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  GitLab API │  │  ArgoCD API │  │  Jira API   │
│             │  │             │  │             │
│ - branches  │  │ - apps      │  │ - issues    │
│ - tags      │  │ - canary    │  │ - tasks     │
│ - projects  │  │ - traffic   │  │ - projects  │
└─────────────┘  └─────────────┘  └─────────────┘
```

---

## 🎯 Примеры использования

### Пример 1: Получение списка веток

**Backend код:**
```go
// backend/internal/handler/git_handler.go
func (h *GitHandler) ListBranches(c *gin.Context) {
    projectID := c.Query("projectId")
    
    branches, err := h.gitlabClient.ListBranches(c.Request.Context(), projectID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, branches)
}
```

**GitLab API вызов:**
```
GET http://gitlab:80/api/v4/projects/{projectId}/repository/branches
Authorization: Bearer glpat-xxxxxxxxxxxxxxxxxxxx
```

**Ответ:**
```json
[
  {
    "name": "main",
    "commit": {
      "id": "abc123def456"
    }
  },
  {
    "name": "feature/new-payment",
    "commit": {
      "id": "def456ghi789"
    }
  }
]
```

---

### Пример 2: Создание тега

**Backend код:**
```go
// backend/internal/service/canary_service.go
func (s *CanaryService) Create(ctx context.Context, canary *model.CanaryDeployment, ...) error {
    // ...
    
    // РЕАЛЬНЫЙ вызов GitLab API для создания тега
    if s.gitlabClient != nil && canary.GitTag != "" {
        log.Printf("[Canary] Creating Git tag: %s", canary.GitTag)
        
        if err := s.gitlabClient.CreateTag(ctx, canary.ServiceID, canary.GitTag, canary.CanaryRevision); err != nil {
            log.Printf("[Canary] WARNING: Failed to create Git tag: %v", err)
        } else {
            log.Printf("[Canary] Successfully created Git tag")
        }
    }
    
    // ...
}
```

**GitLab API вызов:**
```
POST http://gitlab:80/api/v4/projects/{projectId}/repository/tags
Authorization: Bearer glpat-xxxxxxxxxxxxxxxxxxxx
Content-Type: application/json

{
  "tag_name": "canary-2.14.0-20240115-120000",
  "ref": "feature/new-payment"
}
```

**Ответ:**
```json
{
  "name": "canary-2.14.0-20240115-120000",
  "message": "",
  "commit": {
    "id": "def456ghi789"
  }
}
```

---

## 📝 Полезные команды

### Управление GitLab

```bash
# Запустить GitLab
docker-compose up -d gitlab

# Остановить GitLab
docker-compose stop gitlab

# Перезапустить GitLab
docker-compose restart gitlab

# Посмотреть логи
docker-compose logs -f gitlab

# Получить root пароль
docker-compose exec gitlab cat /etc/gitlab/initial_root_password

# Сбросить root пароль
docker-compose exec gitlab gitlab-rake "gitlab:password:reset[root]"

# Войти в контейнер
docker-compose exec gitlab bash
```

### Управление данными

```bash
# Резервная копия
docker-compose exec gitlab gitlab-backup create

# Восстановление из backup
docker-compose exec gitlab gitlab-backup restore BACKUP=timestamp

# Очистка данных
docker-compose down -v
docker volume rm release-portal_gitlab_config
docker volume rm release-portal_gitlab_logs
docker volume rm release-portal_gitlab_data
```

---

## ✅ Итог

**GitLab успешно добавлен в docker-compose!**

- ✅ GitLab CE 16.8.0 запущен
- ✅ Доступен на http://localhost:8083
- ✅ API доступен на http://gitlab:80
- ✅ Интеграция с Release Portal работает
- ✅ Можно создавать теги и управлять ветками

**Следующие шаги:**
1. Запустите `docker-compose up -d`
2. Дождитесь запуска GitLab (5-10 минут)
3. Получите root пароль
4. Создайте API токен
5. Настройте интеграцию в Release Portal
6. Создайте тестовый проект
7. Проверьте интеграцию

**Документация:**
- [GitLab API Documentation](https://docs.gitlab.com/ee/api/)
- [GitLab Docker Installation](https://docs.gitlab.com/ee/install/docker.html)
- [INTEGRATIONS_FIXED.md](./INTEGRATIONS_FIXED.md) - описание работы интеграций

---

**GitLab готов к использованию!** 🎉
