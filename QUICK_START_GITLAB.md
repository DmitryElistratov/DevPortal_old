# 🚀 Быстрый старт: GitLab интеграция

## 📋 Что нужно сделать

### Шаг 1: Запустите все сервисы

```bash
docker-compose up -d
```

**Ожидаемый вывод:**
```
✅ postgres     - Up (healthy)
✅ backend      - Up
✅ frontend     - Up
✅ keycloak     - Up
✅ vault        - Up
✅ gitlab       - Up (может занять 5-10 минут для первого запуска)
```

### Шаг 2: Дождитесь запуска GitLab

GitLab требует время для инициализации. Проверьте статус:

```bash
docker-compose logs -f gitlab
```

**Дождитесь сообщения:**
```
GitLab is ready!
```

Или проверьте health:
```bash
docker-compose ps
```

Должно показать: `Up (healthy)`

### Шаг 3: Получите root пароль GitLab

```bash
docker-compose exec gitlab cat /etc/gitlab/initial_root_password
```

**Скопируйте пароль** (он будет удалён через 24 часа):
```
Password: abc123XYZ456...
```

### Шаг 4: Войдите в GitLab

1. Откройте: http://localhost:8083
2. Username: `root`
3. Password: (из шага 3)
4. Смените пароль при первом входе

### Шаг 5: Создайте API токен

1. Перейдите: http://localhost:8083/-/profile/personal_access_tokens
2. Заполните форму:
   - **Token name:** `release-portal`
   - **Expiration date:** (оставьте пустым)
   - **Scopes:** ✅ `api`, ✅ `read_user`, ✅ `read_repository`, ✅ `write_repository`
3. Нажмите **"Create personal access token"**
4. **Скопируйте токен** (показывается только один раз!):
   ```
   glpat-xxxxxxxxxxxxxxxxxxxx
   ```

### Шаг 6: Настройте интеграцию в Release Portal

1. Откройте: http://localhost:8080
2. Войдите как администратор
3. Перейдите: **Settings** → **Git**
4. Нажмите **"Добавить"**
5. Заполните форму:
   - **Название:** `Main GitLab`
   - **Тип:** `GitLab`
   - **URL:** `http://gitlab:80`
   - **Access Token:** `glpat-xxxxxxxxxxxxxxxxxxxx` (из шага 5)
6. Нажмите **"Сохранить"**
7. Нажмите **"Проверить соединение"** (иконка пробирки)

**Ожидаемый результат:**
```
✅ Соединение успешно!
```

### Шаг 7: Создайте тестовый проект в GitLab

1. В GitLab UI нажмите **"Create new project"**
2. Выберите **"Create blank project"**
3. Заполните:
   - **Project name:** `payment-service`
   - **Visibility:** `Private`
4. Нажмите **"Create project"**
5. Скопируйте URL проекта (например: `http://localhost:8083/root/payment-service.git`)

### Шаг 8: Добавьте файлы и ветки

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

# Создайте теги
git tag v1.0.0
git push origin v1.0.0
```

### Шаг 9: Проверьте интеграцию

#### Проверка 1: Список веток

В Release Portal:
1. Откройте **Create Release**
2. Выберите сервис `payment-service`
3. В выпадающем списке "Git Branch" должны появиться:
   - ✅ `main`
   - ✅ `feature/new-payment`
   - ✅ `release/1.0.0`

#### Проверка 2: Canary deployment

В Release Portal:
1. Откройте Release со статусом `testing`
2. Нажмите **"Deploy as Canary"**
3. Выберите ArgoCD приложение
4. Установите 10% трафика
5. Нажмите **"Start Canary Deployment"**

#### Проверка 3: Логи backend

```bash
docker-compose logs -f backend | grep -E "(GitLab|Canary)"
```

**Ожидаемый вывод:**
```
[Integration] GitLab client initialized: http://gitlab:80
[Canary] Creating canary deployment in ArgoCD: app=payment-service, version=2.14.0, percentage=10%
[Canary] Successfully created canary deployment in ArgoCD
[Canary] Creating Git tag: canary-2.14.0-20240115-120000
[Canary] Successfully created Git tag
```

#### Проверка 4: GitLab UI

Откройте проект `payment-service` в GitLab:
1. Перейдите в **Repository → Tags**
2. Должен появиться новый тег:
   - ✅ `canary-2.14.0-20240115-120000`

---

## 🔍 Troubleshooting

### Проблема: GitLab не запускается

**Решение:**
```bash
# Увеличьте память Docker (минимум 4GB)
# Проверьте логи
docker-compose logs gitlab

# Перезапустите
docker-compose restart gitlab
```

### Проблема: Не могу получить root пароль

**Решение:**
```bash
# Сбросьте root пароль
docker-compose exec gitlab gitlab-rake "gitlab:password:reset[root]"
```

### Проблема: API токен не работает

**Решение:**
1. Проверьте что токен не истёк
2. Проверьте что у токена есть scope `api`
3. Создайте новый токен

### Проблема: Ветки не появляются в Release Portal

**Решение:**
1. Проверьте что интеграция настроена правильно
2. Проверьте что проект существует в GitLab
3. Проверьте логи backend:
   ```bash
   docker-compose logs backend | grep GitLab
   ```

---

## 📊 Полезные команды

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

# Резервная копия
docker-compose exec gitlab gitlab-backup create
```

---

## ✅ Итог

**GitLab интеграция готова!**

- ✅ GitLab CE запущен на http://localhost:8083
- ✅ API доступен на http://gitlab:80
- ✅ Интеграция с Release Portal настроена
- ✅ Можно создавать теги и управлять ветками
- ✅ Canary deployment работает с реальным GitLab

**Следующие шаги:**
- Создайте больше проектов в GitLab
- Настройте CI/CD pipelines
- Интегрируйте с ArgoCD для автоматического деплоя

---

**Готово к использованию!** 🎉

**Документация:**
- [GITLAB_SETUP.md](./GITLAB_SETUP.md) - полная документация
- [INTEGRATIONS_FIXED.md](./INTEGRATIONS_FIXED.md) - описание работы интеграций
