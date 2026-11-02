# 🔧 Решение проблемы "Failed to load data"

## 🐛 Проблема

На главном экране появляется ошибка:
```
Failed to load data. Please check if backend is running.
```

## 🔍 Причины

1. **Backend не запущен** - контейнер backend упал или не стартовал
2. **Backend не может подключиться к БД** - PostgreSQL не готов
3. **Frontend обращается к неправильному URL** - `VITE_API_URL` не настроен
4. **Backend падает при инициализации** - ошибка в коде

## ✅ Решение

### Шаг 1: Проверить статус контейнеров

```bash
docker-compose ps
```

**Ожидаемый результат:**
```
NAME                    STATUS
release-portal-postgres-1   Up (healthy)
release-portal-backend-1    Up
release-portal-frontend-1   Up
```

Если backend не запущен, переходите к шагу 2.

---

### Шаг 2: Проверить логи backend

```bash
docker-compose logs backend
```

**Ищите ошибки:**
- `Failed to connect to database` - проблема с PostgreSQL
- `panic:` - ошибка в коде
- `Integration client initialized` - клиенты инициализированы

**Пример успешного запуска:**
```
Starting server on port 8080
[Integration] ArgoCD client status: false
[Integration] GitLab client status: false
[Integration] Jira client status: false
[Integration] Kubernetes client status: false
[Integration] Grafana client status: false
[Integration] Graylog client status: false
[Integration] Vault client status: false
```

---

### Шаг 3: Перезапустить backend

```bash
docker-compose restart backend
```

Или полностью пересобрать:

```bash
docker-compose down
docker-compose build --no-cache backend
docker-compose up -d
```

---

### Шаг 4: Проверить здоровье backend

```bash
curl http://localhost:8081/health
```

**Ожидаемый ответ:**
```json
{"status":"healthy"}
```

---

### Шаг 5: Проверить подключение к БД

```bash
docker-compose exec postgres psql -U postgres -c "SELECT version();"
```

**Ожидаемый результат:**
```
                                                 version                                                 
---------------------------------------------------------------------------------------------------------
 PostgreSQL 16.x on x86_64-pc-linux-musl, compiled by gcc (Alpine 13.2.1_git20231014) 13.2.1 0, 64-bit
```

---

### Шаг 6: Проверить таблицы в БД

```bash
docker-compose exec postgres psql -U postgres -d release_portal -c "\dt"
```

**Ожидаемый результат:**
```
              List of relations
 Schema |        Name         | Type  |  Owner   
--------+---------------------+-------+----------
 public | active_versions     | table | postgres
 public | audit_log           | table | postgres
 public | canary_deployments  | table | postgres
 public | integrations        | table | postgres
 public | project_permissions | table | postgres
 public | releases            | table | postgres
 public | services            | table | postgres
 public | users               | table | postgres
```

---

### Шаг 7: Пересобрать frontend с правильным URL

**Проблема:** Frontend собран с `VITE_API_URL=http://localhost:8081`, но внутри Docker это не работает.

**Решение:**

1. Остановить контейнеры:
```bash
docker-compose down
```

2. Пересобрать frontend:
```bash
docker-compose build --no-cache frontend
```

3. Запустить контейнеры:
```bash
docker-compose up -d
```

---

### Шаг 8: Проверить frontend

```bash
curl http://localhost:8080/health
```

**Ожидаемый ответ:**
```
healthy
```

---

### Шаг 9: Открыть браузер

Откройте http://localhost:8080

**Если ошибка осталась:**
1. Откройте DevTools (F12)
2. Перейдите на вкладку Console
3. Найдите ошибку
4. Проверьте Network tab - какие запросы failing

---

## 🔧 Дополнительные проверки

### Проверить что backend слушает правильный порт

```bash
docker-compose exec backend netstat -tlnp | grep 8080
```

**Ожидаемый результат:**
```
tcp        0      0 :::8080                 :::*                    LISTEN      1/main
```

### Проверить что frontend может обратиться к backend

```bash
docker-compose exec frontend wget -qO- http://backend:8080/health
```

**Ожидаемый результат:**
```
{"status":"healthy"}
```

### Проверить логи frontend

```bash
docker-compose logs frontend
```

**Ищите ошибки:**
- `502 Bad Gateway` - nginx не может подключиться к backend
- `Connection refused` - backend не слушает порт
- `404 Not Found` - неправильный URL

---

## 🚨 Если backend падает

### Ошибка 1: "Failed to connect to database"

**Причина:** PostgreSQL не готов или неправильный URL

**Решение:**
```bash
# Проверить что PostgreSQL запущен
docker-compose ps postgres

# Перезапустить PostgreSQL
docker-compose restart postgres

# Подождать 10 секунд
sleep 10

# Перезапустить backend
docker-compose restart backend
```

### Ошибка 2: "panic: runtime error"

**Причина:** Ошибка в коде

**Решение:**
```bash
# Посмотреть полный стек ошибки
docker-compose logs backend | grep -A 20 "panic:"

# Проверить что все файлы на месте
ls -la backend/internal/service/
ls -la backend/internal/handler/
ls -la backend/internal/integration/
```

### Ошибка 3: "table does not exist"

**Причина:** Миграции не применены

**Решение:**
```bash
# Удалить volume с БД
docker-compose down -v

# Запустить заново
docker-compose up -d

# Подождать 30 секунд для применения миграций
sleep 30
```

---

## 📊 Диагностика с помощью скрипта

Запустите диагностический скрипт:

```bash
chmod +x diagnose.sh
./diagnose.sh
```

Скрипт проверит:
- ✅ Статус контейнеров
- ✅ Логи backend
- ✅ Здоровье backend
- ✅ Подключение к БД
- ✅ Таблицы в БД
- ✅ Интеграции в БД
- ✅ Здоровье frontend
- ✅ API endpoints

---

## ✅ Ожидаемое поведение

После успешного запуска:

1. **Backend логи:**
```
Starting server on port 8080
[Integration] ArgoCD client status: false
[Integration] GitLab client status: false
...
```

2. **Health check:**
```bash
curl http://localhost:8081/health
# {"status":"healthy"}
```

3. **Frontend:**
- Открывается http://localhost:8080
- Загружается главная страница
- Данные загружаются из backend

---

## 🎯 Быстрое решение

Если ничего не помогает, выполните полную переустановку:

```bash
# Остановить все контейнеры
docker-compose down -v

# Удалить все образы
docker-compose rm -f

# Пересобрать все образы
docker-compose build --no-cache

# Запустить заново
docker-compose up -d

# Подождать 30 секунд
sleep 30

# Проверить статус
docker-compose ps
```

---

## 📞 Если проблема не решена

1. Проверьте логи:
```bash
docker-compose logs backend > backend.log
docker-compose logs frontend > frontend.log
docker-compose logs postgres > postgres.log
```

2. Проверьте конфигурацию:
```bash
cat .env
cat docker-compose.yml
cat backend/Dockerfile
```

3. Создайте issue с логами

---

**Проблема должна быть решена!** ✅
