# MAX Mini App — поиск партий

Проект: Vue 3 фронт + Go бэкенд + Postgres (пользователи, история поисков) + MSSQL (DAX) + observability стек (Grafana/Loki/Tempo/OTel/Pyroscope/Prometheus).  
Доступ к UI через Nginx, API через `/api`.

## Требования
- Docker + Docker Compose
- Loki docker‑driver (фиксированная версия)

### Установка Loki драйвера (пример для amd64)
```
docker plugin install grafana/loki-docker-driver:3.6.7-amd64 --alias loki --grant-all-permissions
```

## Переменные окружения
Скопируйте `.env.example` в `.env` и заполните:
- `BOT_TOKEN` — токен MAX бота
- `DAX_MSSQL_*` — доступ к DAX/MSSQL
- `POSTGRES_*` — доступ к Postgres
  
Для локальной проверки фронта можно задать `VITE_INIT_DATA` в `frontend/.env` (не коммитить).  
При проблемах с Vite proxy можно указать базовый URL API:
```
VITE_API_BASE=http://localhost:3000
```

## Подключение реального MAX‑бота (dev через туннель)
1) Поднимите dev‑стек и убедитесь, что UI открывается локально:
```
docker compose -f docker-compose.dev.yml up --build -d
```
2) Откройте туннель на **`http://localhost:8080`** (nginx dev).  
   Это важно — `8080` проксирует и фронт, и `/api` к бэку.
3) В кабинете MAX укажите **Web App URL** = ваш туннель (https).
4) Убедитесь, что в `.env` установлен **`BOT_TOKEN`** именно этого бота.
5) Откройте мини‑приложение из реального бота — `initData` придёт из MAX автоматически
   (`window.WebApp.InitData`), и валидация пройдёт на бэке.

Примечания:
- Если хотите тестировать **без MAX**, оставьте `ENABLE_INITDATA_MOCK=true` (dev endpoint).
- Для реального MAX можно поставить `ENABLE_INITDATA_MOCK=false`, чтобы отключить dev‑генерацию.
- Если туннель указывает на `5173` (Vite), используйте `VITE_API_BASE`
  и проверьте `server.allowedHosts` в `frontend/vite.config.js`.
- Если хочешь, могу ещё добавить раздел с пошаговым чек‑листом по MAX‑кабинету (скриншоты/поля).

### Важно для MAX WebView (если «ничего не грузит»)
Внутри MAX WebView dev‑сервер Vite с HMR может не работать.  
Рекомендуется использовать **статическую сборку** фронта:

1) Собрать фронт:
```
docker compose -f docker-compose.dev.yml exec frontend npm run build
```
2) Запустить статический nginx:
```
docker compose -f docker-compose.dev.yml up -d web-tunnel
```
3) Открыть туннель уже на `http://localhost:8081` и эту ссылку указать в MAX.

## Тест без MAX (initData вручную)
Есть два варианта:
1) **Через `.env` фронта**  
Создайте `frontend/.env`:
```
VITE_INIT_DATA=auth_date%3D...%26query_id%3D...%26user%3D%257B...%257D%26hash%3D...
```
Перезапустите фронт.

2) **Через ссылку**  
Откройте:
```
https://YOUR_TUNNEL_URL/?initData=auth_date%3D...%26query_id%3D...%26user%3D%257B...%257D%26hash%3D...
```
После первого открытия значение сохраняется в `localStorage` (`maxapp_init_data`), и можно заходить без параметра.

Важно:
- `initData` должен соответствовать `BOT_TOKEN`, иначе валидация не пройдёт.
- Строку `initData` нужно URL‑кодировать (как в документации MAX).

### Генерация initData на деве (без MAX)
В dev‑режиме доступен endpoint, который выдаёт подписанную строку:
```
http://localhost:3000/api/dev/init-data?user_id=1001&first_name=Ivan&last_name=Ivanov&username=ivan&language=ru&front=http://localhost:5173
```
Ответ:
```
{
  "success": true,
  "data": {
    "initData": "...",
    "url": "http://localhost:5173/?initData=..."
  }
}
```
Откройте `data.url` в браузере.  
Endpoint работает только если `ENABLE_INITDATA_MOCK=true`.
Если URL уже был сформирован без кодирования — фронт соберёт initData из `auth_date/hash/user` автоматически.
Также можно вставить `initData` вручную в поле **InitData (dev)** на странице ошибки.
Или нажать кнопку **Сгенерировать initData (dev)** прямо на странице ошибки.

## Таймауты и длинные запросы
Если запросы к DAX выполняются долго (например, `@`‑режим), нужен увеличенный таймаут
на всех уровнях. Где это править:

1) **Nginx (dev/prod)**  
`deploy/nginx/nginx.dev.conf` и `deploy/nginx/nginx.conf`:
- `proxy_connect_timeout`
- `proxy_send_timeout`
- `proxy_read_timeout`

2) **Vite proxy (dev)**  
`frontend/vite.config.js`:
- `server.proxy["/api"].timeout`
- `server.proxy["/api"].proxyTimeout`

3) **HTTP‑сервер Go**  
`backend/config.yml` (dev) и `backend/docker/config.app.yml` (prod):
- `http_server.timeout`

После изменения таймаутов пересоберите:
```
docker compose -f docker-compose.dev.yml up -d --build backend-dev-live web-dev frontend
```

## Запуск (dev)
1. Скопировать `.env.example` → `.env` и заполнить.
2. Поднять dev‑стек:
```
docker compose -f docker-compose.dev.yml up --build -d
```
3. Применить миграции (см. раздел ниже).
4. Открыть UI и проверить доступ.

Dev UI: `http://localhost:8080`  
Backend: `http://localhost:3000`  
Grafana: `http://localhost:3001`  
Prometheus: `http://localhost:9090`  
PgAdmin: `http://localhost:5050`

## Запуск (prod)
1. Скопировать `.env.example` → `.env` и заполнить.
2. Поднять prod‑стек:
```
docker compose up --build -d
```
3. Применить миграции (см. раздел ниже).
4. Проверить health endpoints и логи.

По умолчанию UI доступен на портах **80** (HTTP) и **443** (HTTPS). Если на сервере эти порты заняты, в `.env` задайте `WEB_HTTP_PORT` и `WEB_HTTPS_PORT` (например `88` и `8443`) — тогда открывай `http://<хост>:88` и `https://<хост>:8443`.

Backend API (напрямую): `http://localhost:3000`

## Обновление кода (prod)
1. Обновить код в репозитории.
2. Пересобрать и перезапустить контейнеры:
```
docker compose up --build -d
```
3. Применить новые миграции (если есть).
4. Проверить статус:
```
docker compose ps
```
5. Проверить health endpoints:
```
curl http://localhost:3000/healthz
curl http://localhost:3000/readyz
```

## Миграции
### Dev (обновить миграции)
```
docker compose -f docker-compose.dev.yml --profile tools run --rm migrate \
  -path /migrations \
  -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" up
```

### Prod (обновить миграции)
```
docker compose --profile tools run --rm migrate \
  -path /migrations \
  -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" up
```

### Проверить текущую версию миграций
```
docker compose --profile tools run --rm migrate \
  -path /migrations \
  -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" version
```

## Проверка инициализации при запуске
### Логи сервиса
```
docker compose logs -f backend
docker compose -f docker-compose.dev.yml logs -f backend-dev-live
```
Ожидаемые строки:
- `telemetry initialized` (tracing работает)
- `profiling initialized` (pyroscope работает)
- `postgres connected` / `mssql connected`
- `http server started`
- `metrics endpoint ready`

### Метрики
- `http://localhost:3000/metrics` — endpoint метрик
- `http://localhost:9090` → targets → `backend` должен быть `UP`

### Трейсы
- `http://localhost:3001` → Explore → Tempo, Service `maxapp`

### Логи (Loki)
- `http://localhost:3001` → Explore → Loki, запрос `{app_name="maxapp"}`

## Частые ошибки и диагностика
### `Unexpected token '<' ... is not valid JSON`
Это означает, что вместо JSON пришла HTML‑страница ошибки (обычно `502/504` от прокси).
Причина — таймаут или недоступный бэк.
Что делать:
- Проверьте таймауты (см. раздел выше).
- Смотрите логи `web-dev` и `backend-dev-live`.

### `InitData пустая` при наличии Debug length
Ошибка появляется, если нажать **«Использовать initData»** с пустым полем.
При наличии `Debug: initData length > 0` просто повторно запустите валидацию
(или вставьте initData вручную).

### Где смотреть логи
Dev:
```
docker compose -f docker-compose.dev.yml logs -f backend-dev-live
docker compose -f docker-compose.dev.yml logs -f web-dev
```
Prod:
```
docker compose logs -f backend
docker compose logs -f web
```

## PgAdmin
1. Открыть `http://localhost:5050`.
2. Логин: `PGADMIN_DEFAULT_EMAIL`, пароль: `PGADMIN_DEFAULT_PASSWORD`.
3. Добавить сервер:
   - Host name/address: `postgres`
   - Port: `5432`
   - Username/Password: из `.env`

## Основные API
- `POST /api/auth/validate` — валидация initData
- `POST /api/users/profile` — указать имя/фамилию
- `GET /api/warehouse/batches?code=...` — поиск партий
- `GET /api/admin/users` — список пользователей (admin)
- `PATCH /api/admin/users/{id}` — апрув/блок/назначение admin (admin)
