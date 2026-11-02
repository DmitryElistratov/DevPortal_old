#!/bin/bash

# Диагностический скрипт для проверки работы Release Portal

echo "🔍 Диагностика Release Portal"
echo "================================"
echo ""

# Проверка Docker контейнеров
echo "📦 Проверка Docker контейнеров..."
docker-compose ps
echo ""

# Проверка логов backend
echo "📋 Логи backend (последние 50 строк):"
docker-compose logs --tail=50 backend
echo ""

# Проверка здоровья backend
echo "🏥 Проверка здоровья backend..."
curl -s http://localhost:8081/health || echo "❌ Backend не отвечает"
echo ""

# Проверка подключения к БД
echo "🗄️ Проверка подключения к PostgreSQL..."
docker-compose exec postgres psql -U postgres -c "SELECT version();" 2>/dev/null || echo "❌ PostgreSQL не отвечает"
echo ""

# Проверка таблиц в БД
echo "📊 Проверка таблиц в БД..."
docker-compose exec postgres psql -U postgres -d release_portal -c "\dt" 2>/dev/null || echo "❌ Таблицы не найдены"
echo ""

# Проверка интеграций в БД
echo "🔌 Проверка интеграций в БД..."
docker-compose exec postgres psql -U postgres -d release_portal -c "SELECT id, type, name, enabled FROM integrations;" 2>/dev/null || echo "❌ Интеграции не найдены"
echo ""

# Проверка frontend
echo "🌐 Проверка frontend..."
curl -s http://localhost:8080/health || echo "❌ Frontend не отвечает"
echo ""

# Проверка API endpoints
echo "🔗 Проверка API endpoints..."
echo "Services:"
curl -s http://localhost:8081/api/services -H "Authorization: Bearer test" | head -c 200
echo ""
echo ""

echo "✅ Диагностика завершена"
echo ""
echo "💡 Если backend не отвечает, проверьте:"
echo "   1. docker-compose logs backend"
echo "   2. docker-compose restart backend"
echo "   3. docker-compose down && docker-compose up -d"
echo ""
echo "💡 Если frontend не отвечает, проверьте:"
echo "   1. docker-compose logs frontend"
echo "   2. docker-compose restart frontend"
echo ""
