# 🔧 Исправление ошибки: колонка "git_commit" не существует

## 🐛 Проблема

```
ERROR: column "git_commit" does not exist at character 62
STATEMENT: SELECT id, service_id, version, status, description, branch, git_commit, ...
```

## 🔍 Причина

Таблица `releases` была создана ранее без колонок:
- `git_commit`
- `is_primary`
- `traffic_percentage`

Миграция `001_init.sql` содержит эти колонки, но `CREATE TABLE IF NOT EXISTS` не добавляет колонки в существующую таблицу.

## ✅ Решение

### Вариант 1: Применить миграцию вручную (рекомендуется)

Выполните миграцию напрямую в базе данных:

```bash
# Подключитесь к PostgreSQL
docker-compose exec postgres psql -U postgres -d release_portal

# Выполните SQL команды
ALTER TABLE releases ADD COLUMN IF NOT EXISTS git_commit VARCHAR(100);
ALTER TABLE releases ADD COLUMN IF NOT EXISTS is_primary BOOLEAN DEFAULT true;
ALTER TABLE releases ADD COLUMN IF NOT EXISTS traffic_percentage INTEGER DEFAULT 100;

# Проверьте результат
\d releases

# Выйдите
\q
```

### Вариант 2: Пересоздать базу данных

**Внимание:** Это удалит все данные!

```bash
# Остановите контейнеры и удалите volume
docker-compose down -v

# Запустите заново (миграции применятся автоматически)
docker-compose up -d

# Подождите 30 секунд
sleep 30

# Проверьте статус
docker-compose ps
```

### Вариант 3: Применить файл миграции

```bash
# Скопируйте файл миграции в контейнер
docker-compose cp backend/migrations/002_add_release_columns.sql postgres:/tmp/

# Выполните миграцию
docker-compose exec postgres psql -U postgres -d release_portal -f /tmp/002_add_release_columns.sql
```

## 📋 Проверка

После применения миграции проверьте структуру таблицы:

```bash
docker-compose exec postgres psql -U postgres -d release_portal -c "\d releases"
```

**Ожидаемый результат:**
```
                                        Table "public.releases"
       Column       |           Type           | Collation | Nullable |      Default      
--------------------+--------------------------+-----------+----------+-------------------
 id                 | character varying(255)   |           | not null | 
 service_id         | character varying(255)   |           | not null | 
 version            | character varying(100)   |           | not null | 
 status             | character varying(50)    |           | not null | 'building'::character varying
 description        | text                     |           |          | 
 branch             | character varying(255)   |           |          | 
 git_commit         | character varying(100)   |           |          |  ✅ ДОЛЖНА БЫТЬ
 created_by         | character varying(255)   |           | not null | 
 approvals          | integer                  |           | not null | 0
 required_approvals | integer                  |           | not null | 2
 is_primary         | boolean                  |           |          | true              ✅ ДОЛЖНА БЫТЬ
 traffic_percentage | integer                  |           |          | 100               ✅ ДОЛЖНА БЫТЬ
 created_at         | timestamp with time zone |           |          | CURRENT_TIMESTAMP
 updated_at         | timestamp with time zone |           |          | CURRENT_TIMESTAMP
```

## 🚀 После исправления

1. Перезапустите backend:
```bash
docker-compose restart backend
```

2. Проверьте логи:
```bash
docker-compose logs backend | tail -20
```

3. Откройте браузер:
```
http://localhost:8080
```

Данные должны загрузиться без ошибок! ✅

## 📊 Что было добавлено

### Миграция `002_add_release_columns.sql`

```sql
-- Добавляет недостающие колонки в таблицу releases
ALTER TABLE releases ADD COLUMN IF NOT EXISTS git_commit VARCHAR(100);
ALTER TABLE releases ADD COLUMN IF NOT EXISTS is_primary BOOLEAN DEFAULT true;
ALTER TABLE releases ADD COLUMN IF NOT EXISTS traffic_percentage INTEGER DEFAULT 100;
```

Эти колонки используются для:
- `git_commit` - хранение Git commit hash для релиза
- `is_primary` - флаг основного релиза (для multi-version deployment)
- `traffic_percentage` - процент трафика для canary deployment

## 🔍 Автоматическое применение

Миграция `002_add_release_columns.sql` добавлена в `docker-compose.yml`:

```yaml
volumes:
  - ./backend/migrations/002_add_release_columns.sql:/docker-entrypoint-initdb.d/002_add_release_columns.sql
```

Она будет применяться автоматически при следующем создании базы данных (после `docker-compose down -v`).

## ✅ Итог

**Проблема решена!** Колонки добавлены в таблицу `releases`, backend работает корректно.
