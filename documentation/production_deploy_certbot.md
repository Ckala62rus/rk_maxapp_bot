# Прод: Ubuntu + Docker + certbot (HTTPS) + код на сервере

Эта инструкция описывает сценарий, когда **код размещается на боевом сервере**,
и сервер **сам собирает** Docker-образы. HTTPS сертификат получаем через `certbot`.

## 0) Что будет в итоге
- Вход по домену: `https://your.domain`
- TLS завершается **внутри контейнера** `web` (nginx).
- Приложение запускается через `docker-compose.yml` (prod).
- `docker-compose.deploy.yml` оставлен как альтернатива, если вы не хотите хранить код на сервере.
- Миграции выполняются через сервис `migrate` (profile `tools`).

---

## 1) Установка Docker + Docker Compose

### 1.1 Проверить архитектуру
```bash
uname -m
```
Обычно это `x86_64` (amd64). Ниже команды для amd64.

### 1.2 Установка Docker и compose-плагина
```bash
sudo apt update
sudo apt install -y docker.io docker-compose-plugin
sudo systemctl enable --now docker
```

### 1.3 (Важно) Плагин логирования Loki
В `docker-compose.yml` используется драйвер логирования `loki`, поэтому плагин обязателен.

Для **amd64**:
```bash
docker plugin install grafana/loki-docker-driver:3.6.7-amd64 \
  --alias loki --grant-all-permissions
```

Если архитектура **arm64**, используй:
```bash
docker plugin install grafana/loki-docker-driver:3.6.7-arm64 \
  --alias loki --grant-all-permissions
```

Проверить, что плагин есть:
```bash
docker plugin ls
```

---

## 2) Установка certbot
```bash
sudo apt update
sudo apt install -y certbot
```

---

## 3) Клонировать код на сервер
```bash
git clone <YOUR_REPO_URL> bot
cd bot
```

---

## 4) Подготовить `.env`
Скопируйте `.env.example` в `.env` и заполните минимум:
- `BOT_TOKEN`
- `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- `DAX_MSSQL_HOST`, `DAX_MSSQL_USER`, `DAX_MSSQL_PASSWORD`, `DAX_MSSQL_DB` (если используется)
- `CERT_DOMAIN=your.domain`
- `PGADMIN_DEFAULT_EMAIL`, `PGADMIN_DEFAULT_PASSWORD` (если нужен pgAdmin)

---

## 5) Подготовить директории для certbot (webroot)
```bash
sudo mkdir -p /var/www/certbot
sudo mkdir -p /etc/letsencrypt
```

Открой порты `80` и `443` наружу (firewall/security group).

---

## 6) Сборка образов на сервере
Собираем все нужные образы **на сервере**:
```bash
docker build -t maxapp-backend:prod ./backend
docker build -t maxapp-web:prod -f deploy/nginx/Dockerfile .
docker build -t maxapp-migrate:prod -f deploy/migrate/Dockerfile .
```

---

## 7) Запуск prod compose (пока HTTP для certbot)
```bash
docker compose up -d
```

На этом шаге контейнер `web` в HTTP-режиме отдаёт:
- `/.well-known/acme-challenge/` из `/var/www/certbot`

---

## 8) Получить сертификат (certbot, webroot)
Пример:
```bash
sudo certbot certonly \
  --webroot -w /var/www/certbot \
  -d "${CERT_DOMAIN}" \
  --email your-email@domain.com \
  --agree-tos --no-eff-email
```

После успешного выполнения появятся файлы:
- `/etc/letsencrypt/live/${CERT_DOMAIN}/fullchain.pem`
- `/etc/letsencrypt/live/${CERT_DOMAIN}/privkey.pem`

---

## 9) Включить HTTPS в контейнере `web`
После получения сертификата перезапускаем nginx:
```bash
docker compose restart web
```

Теперь доступно:
- `https://${CERT_DOMAIN}`

---

## 10) Миграции
```bash
docker compose --profile tools run --rm migrate \
  -path /migrations \
  -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" up
```

---

## 11) Проверка
1) Открыть в браузере:
   - `https://${CERT_DOMAIN}`

2) Проверить health endpoints внутри контейнера `backend`:
```bash
docker compose exec backend sh -lc 'curl -s localhost:3000/healthz || true'
docker compose exec backend sh -lc 'curl -s localhost:3000/readyz || true'
```

3) Логи:
```bash
docker compose logs -f web
docker compose logs -f backend
```

---

## 12) Обновление кода на сервере
После изменений в репозитории:
```bash
git pull
docker build -t maxapp-backend:prod ./backend
docker build -t maxapp-web:prod -f deploy/nginx/Dockerfile .
docker build -t maxapp-migrate:prod -f deploy/migrate/Dockerfile .
docker compose up -d
```

---

## 13) Renew (обновление сертификатов)
1) Обновить сертификаты:
```bash
sudo certbot renew
```

2) Перезагрузить nginx в контейнере `web`:
```bash
docker compose exec web nginx -s reload
```

Если reload не применил изменения — перезапустите контейнер:
```bash
docker compose restart web
```

---

## 14) Если сертификаты выдают не через certbot
Для нашей схемы нужно, чтобы внутри контейнера были видны:
- `fullchain.pem`
- `privkey.pem`

Самый простой вариант:
1) положить `fullchain.pem` и `privkey.pem` в стандартный путь:
   - `/etc/letsencrypt/live/${CERT_DOMAIN}/fullchain.pem`
   - `/etc/letsencrypt/live/${CERT_DOMAIN}/privkey.pem`
2) выполнить:
```bash
docker compose restart web
```

Если сертификаты лежат в другом месте — подскажу, как смонтировать и какие env выставить (`SSL_CERT_PATH` / `SSL_KEY_PATH`).

