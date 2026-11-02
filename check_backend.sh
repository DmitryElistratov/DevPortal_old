#!/bin/bash

# Скрипт для проверки и пересборки backend

echo "🔍 Проверка backend..."

# Проверить что backend пересобран
echo "1. Проверка версии backend..."
docker-compose exec backend cat /app/cmd/api/main.go | grep -A 5 "integration :=" | head -10

echo ""
echo "2. Проверка handler Sync..."
docker-compose exec backend grep -n "func (h \*IntegrationHandler) Sync" /app/internal/handler/integration_handler.go

echo ""
echo "3. Проверка маршрутов в логах..."
docker-compose logs backend | grep -i "route" | tail -20

echo ""
echo "4. Пересборка backend..."
docker-compose down
docker-compose build --no-cache backend
docker-compose up -d

echo ""
echo "5. Ожидание запуска..."
sleep 30

echo ""
echo "6. Проверка логов..."
docker-compose logs backend | tail -30

echo ""
echo "7. Тест endpoint..."
curl -X POST http://localhost:8081/api/integrations/test-id/sync \
  -H "Authorization: Bearer test" \
  -H "Content-Type: application/json" \
  -v

echo ""
echo "✅ Проверка завершена"
