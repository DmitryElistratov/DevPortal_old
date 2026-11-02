#!/bin/bash

echo "🔄 Пересборка backend..."

# Остановить
docker-compose down

# Очистить кэш
docker system prune -f

# Пересобрать
docker-compose build --no-cache backend

# Запустить
docker-compose up -d

# Подождать
echo "⏳ Ожидание запуска..."
sleep 30

# Проверить
echo "🔍 Проверка..."
docker-compose logs backend | tail -20

echo "✅ Готово!"
