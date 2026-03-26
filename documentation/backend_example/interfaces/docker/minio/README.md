# MinIO + Nginx (reverse proxy)

> Примечание: основной стек теперь в корневых `docker-compose.yml` и `docker-compose.dev.yml`.
> Этот файл оставлен как справка по конфигурации MinIO/Nginx.

Локальный стенд для тестирования `minio-go` и `presigned URL` без правок `hosts`.

## Что поднимается

- `minio` - S3-совместимое хранилище
- `nginx` - reverse proxy перед MinIO
  - `http://localhost:9000` -> MinIO S3 API
  - `http://localhost:9001` -> MinIO Console

## Запуск

Из папки `docker/minio`:

```bash
docker compose up -d
```

Проверить:

```bash
docker compose ps
```

Открыть MinIO Console:

- URL: `http://localhost:9001`
- Login: `minioadmin` (или `MINIO_ROOT_USER`)
- Password: `minioadmin` (или `MINIO_ROOT_PASSWORD`)

## Почему это решает проблему с hosts

Ваше приложение может генерировать presigned URL с хостом `localhost:9000`, а браузер будет ходить в `nginx`, который прозрачно проксирует запрос в MinIO с сохранением `Host` и query-параметров. За счет этого подпись не ломается, и внутренние имена контейнеров (`minio:9000`) не светятся наружу.

## Полезные переменные

Для MinIO можно переопределить:

- `MINIO_ROOT_USER`
- `MINIO_ROOT_PASSWORD`
- `MINIO_SERVER_URL` (по умолчанию `http://localhost:9000`)

Пример запуска с переменными:

```bash
MINIO_ROOT_USER=admin MINIO_ROOT_PASSWORD=admin123 docker compose up -d
```

PowerShell:

```powershell
$env:MINIO_ROOT_USER="admin"
$env:MINIO_ROOT_PASSWORD="admin123"
docker compose up -d
```

## Smoke test из Go-кода

После старта контейнеров можно проверить интеграцию (upload + presigned URL + download через Nginx):

```bash
go test ./test -run TestMinIOSmoke_ThroughNginx -v
```

Тест использует переменные окружения (с дефолтами):

- `MINIO_ENDPOINT` (default: `localhost:9000`)
- `MINIO_ROOT_USER` (default: `minioadmin`)
- `MINIO_ROOT_PASSWORD` (default: `minioadmin`)
- `MINIO_USE_SSL` (default: `false`)
