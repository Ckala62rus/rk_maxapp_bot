# Разработка, изменения и сборка

## Требования
- Docker + Docker Compose
- (Опционально) локальные Go/Node, если хотите собирать без контейнеров

## Переменные окружения
1) Скопируйте `.env.example` в `.env`.
2) Заполните как минимум:
   - `BOT_TOKEN` — токен MAX бота
   - `POSTGRES_*` — доступ к Postgres
   - `DAX_MSSQL_*` — доступ к DAX/MSSQL

## Запуск dev‑стека
```bash
docker compose -f docker-compose.dev.yml up --build -d
```

Что будет запущено:
- Backend (hot‑reload через `air`)
- Frontend (Vite dev server)
- Nginx dev (порт 8080)
- Postgres + pgAdmin

## Внесение изменений

### Frontend
- Файлы фронта находятся в `frontend/`.
- Dev‑контейнер монтирует исходники, Vite подхватывает изменения автоматически.
- Обычно достаточно просто сохранить файл — перезапуск не нужен.

### Backend
- Файлы бэка находятся в `backend/`.
- Dev‑контейнер использует `air` (hot‑reload), изменения подхватываются автоматически.
- Если хот‑релод не сработал, можно перезапустить сервис:
```bash
docker compose -f docker-compose.dev.yml restart backend-dev-live
```

## Сборка и проверка

### Сборка фронта (для WebView/статического nginx)
```bash
docker compose -f docker-compose.dev.yml exec frontend npm run build
```

### Запуск статического фронта (web‑tunnel)
```bash
docker compose -f docker-compose.dev.yml up -d web-tunnel
```
После этого UI будет доступен на `http://localhost:8081`.

### Полная пересборка dev‑стека
```bash
docker compose -f docker-compose.dev.yml up --build -d
```

### Прод‑сборка
```bash
docker compose up --build -d
```

## Миграции БД

### Dev
```bash
docker compose -f docker-compose.dev.yml --profile tools run --rm migrate \
  -path /migrations \
  -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" up
```

### Prod
```bash
docker compose --profile tools run --rm migrate \
  -path /migrations \
  -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" up
```

## Проверка работоспособности
```bash
curl http://localhost:3000/healthz
curl http://localhost:3000/readyz
```

## Где смотреть логи
```bash
docker compose -f docker-compose.dev.yml logs -f backend-dev-live
docker compose -f docker-compose.dev.yml logs -f web-dev
```

## PgAdmin
1) Откройте `http://localhost:5050`.
2) Логин: `PGADMIN_DEFAULT_EMAIL`, пароль: `PGADMIN_DEFAULT_PASSWORD`.
3) Сервер Postgres внутри сети docker: `host=postgres`, `port=5432`.
