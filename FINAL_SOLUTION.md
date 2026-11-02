# ✅ Финальное решение - Пересборка Backend

## Проблема
POST `/api/integrations/:id/sync` возвращает 404

## Решение
Пересобрать backend без кэша:

```bash
docker-compose down
docker-compose build --no-cache backend
docker-compose up -d
```

## Проверка

### 1. Проверить что backend пересобран
```bash
docker-compose logs backend | grep "Starting server"
```

### 2. Проверить маршруты
```bash
docker-compose exec backend cat /app/cmd/api/main.go | grep -A 10 "integration :="
```

Должно показать:
```go
integration := integrations.Group("/:id")
{
    integration.GET("", integrationHandler.Get)
    integration.PUT("", middleware.RequireRole("admin"), integrationHandler.Update)
    integration.DELETE("", middleware.RequireRole("admin"), integrationHandler.Delete)
    integration.POST("/test", integrationHandler.Test)
    integration.POST("/sync", integrationHandler.Sync)
}
```

### 3. Тест endpoint
```bash
curl -X POST http://localhost:8081/api/integrations/git-1793644899542/sync \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
```

Ожидаемый ответ:
```json
{
  "success": true,
  "message": "Sync completed successfully",
  "type": "gitlab",
  "projects_count": 42
}
```

## Если не работает

### Проверить handler
```bash
docker-compose exec backend grep -n "func (h \*IntegrationHandler) Sync" /app/internal/handler/integration_handler.go
```

Должно показать строку с определением функции.

### Проверить компиляцию
```bash
docker-compose logs backend | grep -i "error"
```

Не должно быть ошибок компиляции.

### Пересобрать заново
```bash
docker-compose down
docker system prune -a
docker-compose build --no-cache backend
docker-compose up -d
```

## Статус
✅ Код исправлен
✅ Frontend собран
⏳ Требуется пересборка backend
