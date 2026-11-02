# 🔧 Финальное исправление маршрутов - Backend пересборка

## Проблема

POST запрос к `/api/integrations/:id/sync` возвращает 404, хотя OPTIONS запрос проходит (204).

## Причина

В Gin маршруты с параметрами должны быть зарегистрированы в правильном порядке. Более специфичные маршруты (с дополнительными путями) должны быть зарегистрированы **ПОСЛЕ** общих маршрутов с параметрами.

## Решение

### Изменён порядок маршрутов в `backend/cmd/api/main.go`:

**Было:**
```go
integrations.POST("/:id/test", integrationHandler.Test)
integrations.POST("/:id/sync", integrationHandler.Sync)
integrations.GET("/:id", integrationHandler.Get)
```

**Стало:**
```go
integrations.GET("/:id", integrationHandler.Get)
integrations.PUT("/:id", middleware.RequireRole("admin"), integrationHandler.Update)
integrations.DELETE("/:id", middleware.RequireRole("admin"), integrationHandler.Delete)
integrations.POST("/:id/test", integrationHandler.Test)
integrations.POST("/:id/sync", integrationHandler.Sync)
```

**Почему это работает:**
- Gin обрабатывает маршруты в порядке регистрации
- GET, PUT, DELETE маршруты регистрируются первыми
- POST маршруты с дополнительными путями (`/test`, `/sync`) регистрируются последними
- Это предотвращает конфликты маршрутов

## Инструкция по пересборке

### Шаг 1: Остановить контейнеры

```bash
docker-compose down
```

### Шаг 2: Удалить старые образы

```bash
docker-compose rm -f
docker rmi $(docker images | grep backend | awk '{print $3}') 2>/dev/null || true
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

### Шаг 5: Подождать 30 секунд

```bash
sleep 30
```

### Шаг 6: Проверить логи

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

### Шаг 7: Проверить endpoint

```bash
# Тестирование подключения
curl -X POST http://localhost:8081/api/integrations/git-1793644899542/test \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"

# Синхронизация
curl -X POST http://localhost:8081/api/integrations/git-1793644899542/sync \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
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

## Проверка в браузере

1. Откройте http://localhost:8080
2. Войдите как admin
3. Перейдите в **Settings** → **Git**
4. Найдите интеграцию GitLab
5. Нажмите кнопку 🔄 **"Синхронизировать"**
6. Должно появиться: "✅ Синхронизация завершена!"

## Проверка в DevTools

1. Откройте DevTools (F12)
2. Перейдите на вкладку **Network**
3. Нажмите кнопку синхронизации
4. Проверьте запрос:
   - **URL:** `/api/integrations/git-1793644899542/sync`
   - **Method:** `POST`
   - **Status:** `200 OK` (не 404!)
   - **Response:** JSON с `success: true`

## Troubleshooting

### Проблема: Endpoint всё ещё возвращает 404

**Решение:**
```bash
# Проверить что backend пересобран
docker-compose logs backend | grep "Starting server"

# Если нет, пересобрать
docker-compose build --no-cache backend
docker-compose up -d
```

### Проблема: Backend не запускается

**Решение:**
```bash
# Проверить логи
docker-compose logs backend

# Перезапустить
docker-compose restart backend
```

### Проблема: Маршруты не регистрируются

**Решение:**
```bash
# Проверить код
docker-compose exec backend cat /app/cmd/api/main.go | grep -A 10 "integrations :="

# Убедиться что маршруты зарегистрированы в правильном порядке
```

## Проверка порядка маршрутов

```bash
docker-compose exec backend cat /app/cmd/api/main.go | grep -A 15 "integrations :="
```

**Ожидаемый вывод:**
```go
integrations := api.Group("/integrations")
{
    // List and create
    integrations.GET("", integrationHandler.List)
    integrations.POST("", middleware.RequireRole("admin"), integrationHandler.Create)
    
    // Get by type (static route with parameter)
    integrations.GET("/type/:type", integrationHandler.GetByType)
    
    // All operations on specific integration by ID
    // Group them together to avoid routing conflicts
    integrations.GET("/:id", integrationHandler.Get)
    integrations.PUT("/:id", middleware.RequireRole("admin"), integrationHandler.Update)
    integrations.DELETE("/:id", middleware.RequireRole("admin"), integrationHandler.Delete)
    integrations.POST("/:id/test", integrationHandler.Test)
    integrations.POST("/:id/sync", integrationHandler.Sync)
}
```

## Итог

После пересборки backend все endpoints должны работать корректно:
- ✅ GET `/api/integrations` - список интеграций
- ✅ POST `/api/integrations` - создание интеграции
- ✅ GET `/api/integrations/:id` - получение интеграции
- ✅ PUT `/api/integrations/:id` - обновление интеграции
- ✅ DELETE `/api/integrations/:id` - удаление интеграции
- ✅ POST `/api/integrations/:id/test` - тестирование подключения
- ✅ POST `/api/integrations/:id/sync` - синхронизация данных

**Backend пересобран и готов к использованию!** 🚀
