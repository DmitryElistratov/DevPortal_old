# 🔧 Инструкция по пересборке Backend

## Проблема

POST запрос к `/api/integrations/:id/sync` возвращает 404, хотя OPTIONS запрос проходит (204).

## Причина

Backend не пересобран после изменений в коде. Docker использует старый образ backend.

## Решение

### Шаг 1: Остановить контейнеры

```bash
docker-compose down
```

### Шаг 2: Удалить старые образы

```bash
docker-compose rm -f
docker rmi $(docker images | grep backend | awk '{print $3}')
```

### Шаг 3: Пересобрать backend без кэша

```bash
docker-compose build --no-cache backend
```

**Ожидаемый вывод:**
```
[+] Building 40.0s (15/15) FINISHED
 => [internal] load build definition from Dockerfile
 => [internal] load metadata for docker.io/library/golang:1.22-bookworm
 => [builder 1/7] FROM docker.io/library/golang:1.22-bookworm
 => [internal] load build context
 => [builder 2/7] WORKDIR /app
 => [builder 3/7] COPY go.mod ./
 => [builder 4/7] RUN go mod download
 => [builder 5/7] COPY . .
 => [builder 6/7] RUN go mod tidy
 => [builder 7/7] RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go
 => ✅ Компиляция успешна!
```

### Шаг 4: Запустить контейнеры

```bash
docker-compose up -d
```

### Шаг 5: Проверить логи

```bash
docker-compose logs -f backend
```

**Ожидаемый вывод:**
```
Starting server on port 8080
[Integration] ArgoCD client status: false
[Integration] GitLab client status: true
[Integration] Jira client status: false
[Integration] Kubernetes client status: false
[Integration] Grafana client status: false
[Integration] Graylog client status: false
[Integration] Vault client status: false
Starting sync scheduler...
Sync scheduler started (interval: 1 minute)
```

### Шаг 6: Проверить endpoint

```bash
# Тестирование подключения
curl -X POST http://localhost:8081/api/integrations/git-1793644899542/test \
  -H "Authorization: Bearer <token>"

# Синхронизация
curl -X POST http://localhost:8081/api/integrations/git-1793644899542/sync \
  -H "Authorization: Bearer <token>"
```

**Ожидаемый ответ:**
```json
{
  "success": true,
  "message": "Sync completed successfully",
  "type": "gitlab",
  "name": "Main GitLab",
  "synced_at": "2024-01-15T10:30:00Z",
  "projects_count": 42
}
```

## Проверка маршрутов

Если проблема сохраняется, проверьте регистрацию маршрутов:

```bash
docker-compose logs backend | grep -i "route"
```

**Ожидаемые маршруты:**
```
POST /api/integrations/:id/test
POST /api/integrations/:id/sync
GET  /api/integrations/:id
PUT  /api/integrations/:id
DELETE /api/integrations/:id
```

## Проверка handler

Если маршруты зарегистрированы, но endpoint всё ещё возвращает 404, проверьте handler:

```bash
docker-compose exec backend cat /app/internal/handler/integration_handler.go | grep -A 5 "func (h \*IntegrationHandler) Sync"
```

**Ожидаемый вывод:**
```go
func (h *IntegrationHandler) Sync(c *gin.Context) {
    id := c.Param("id")
    integration, err := h.repo.GetByID(id)
    // ...
}
```

## Проверка main.go

Проверьте регистрацию маршрутов в main.go:

```bash
docker-compose exec backend cat /app/cmd/api/main.go | grep -A 10 "integrations :="
```

**Ожидаемый вывод:**
```go
integrations := api.Group("/integrations")
{
    integrations.GET("", integrationHandler.List)
    integrations.POST("", middleware.RequireRole("admin"), integrationHandler.Create)
    integrations.GET("/type/:type", integrationHandler.GetByType)
    integrations.POST("/:id/test", integrationHandler.Test)
    integrations.POST("/:id/sync", integrationHandler.Sync)
    integrations.GET("/:id", integrationHandler.Get)
    integrations.PUT("/:id", middleware.RequireRole("admin"), integrationHandler.Update)
    integrations.DELETE("/:id", middleware.RequireRole("admin"), integrationHandler.Delete)
}
```

## Troubleshooting

### Проблема: Backend не запускается

**Решение:**
```bash
docker-compose logs backend
```

Проверьте ошибки компиляции и исправьте их.

### Проблема: Маршруты не регистрируются

**Решение:**
```bash
docker-compose exec backend cat /app/cmd/api/main.go | grep "integrations.POST"
```

Убедитесь, что маршруты зарегистрированы.

### Проблема: Handler не найден

**Решение:**
```bash
docker-compose exec backend cat /app/internal/handler/integration_handler.go | grep "func (h \*IntegrationHandler) Sync"
```

Убедитесь, что handler существует.

### Проблема: Endpoint возвращает 404

**Решение:**
1. Пересоберите backend: `docker-compose build --no-cache backend`
2. Перезапустите контейнеры: `docker-compose restart backend`
3. Проверьте логи: `docker-compose logs -f backend`

## Итог

После пересборки backend все endpoints должны работать корректно:
- ✅ POST `/api/integrations/:id/test` - тестирование подключения
- ✅ POST `/api/integrations/:id/sync` - синхронизация данных
- ✅ GET `/api/integrations/:id` - получение интеграции
- ✅ PUT `/api/integrations/:id` - обновление интеграции
- ✅ DELETE `/api/integrations/:id` - удаление интеграции

**Backend пересобран и готов к использованию!** 🚀
