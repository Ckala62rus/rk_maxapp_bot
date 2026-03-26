# Общее устройство системы

## Компоненты
- **Frontend**: Vue 3 (Vite), файл приложения: `frontend/src/App.vue`.
- **Backend**: Go (HTTP API), основная бизнес‑логика в `backend/internal`.
- **БД пользователей**: Postgres (профили, флаги доступа, история).
- **DAX/MSSQL**: поиск партий (складские данные).
- **Reverse proxy**: Nginx (проксирует UI и `/api`).
- **Observability**: Grafana/Loki/Tempo/OTel/Pyroscope/Prometheus (по желанию).

## Авторизация (initData от MAX WebApp)
Ниже описан реальный поток, который реализован в коде (см. `backend/internal/service/auth_service.go` и `backend/internal/app/middleware/auth.go`).

1) **Откуда берется initData**
   - В реальном MAX WebApp `initData` приходит от SDK в браузере (`window.WebApp.InitData`).
   - Во фронте строка берется из SDK или вручную (dev) и используется дальше.

2) **Как строка попадает на бэкенд**
   - Для первичной валидации фронт вызывает `POST /api/auth/validate` и отправляет `{ initData: "..." }`.
   - Для остальных защищенных запросов фронт добавляет заголовок `X-Max-InitData` с той же строкой.

3) **Что делает бэкенд с initData**
   - Строка initData **URL‑декодируется** и парсится как query‑параметры (`key=value`).
   - Из параметров извлекается `hash`.
   - Строится **data‑check string**: сортировка ключей (кроме `hash`) и склейка `key=value` через `\n`.
   - Вычисляется `secret = HMAC_SHA256("WebAppData", BOT_TOKEN)`.
   - Вычисляется `expected = HMAC_SHA256(secret, data-check string)` и сравнивается с `hash`.
   - Если подпись валидна, парсится `user` (JSON), берется `id`.

4) **Создание/поиск пользователя**
   - По `max_user_id` ищется пользователь в Postgres.
   - Если пользователя нет — создается новый с флагами:
     - `isAdmin = false`
     - `isBlocked = false`
     - `isApproved = false`

5) **Результат**
   - В `POST /api/auth/validate` возвращаются флаги (`isAdmin`, `isApproved`, `isBlocked`)
     и информация профиля, чтобы UI знал, что показывать.
   - В middleware (`X-Max-InitData`) этот же механизм используется для защиты всех API.

## Админ‑раздел
- `GET /api/admin/users?q=...` — поиск пользователей по **ФИО профиля** (first_name/last_name).
- `PATCH /api/admin/users/{id}` — изменение флагов (`isApproved`, `isBlocked`, `isAdmin`).
- В UI есть кнопки: **Одобрить**, **Блок**, **Разблок**, **Сделать админом**, **Снять админа**.

## Поиск партий
- `GET /api/warehouse/batches?code=...` — запрос в DAX/MSSQL.
- Доступен только подтвержденным и не заблокированным пользователям.
