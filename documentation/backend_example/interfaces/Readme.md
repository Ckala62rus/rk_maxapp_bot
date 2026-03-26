PS D:\GO\projects\interfaces> docker compose -f [docker-compose.dev](http://docker-compose.dev).yml up -d --force-recreate tempo otel-collector

[+] Running 1/1

 ✘ tempo Error                                                                                                                                                                                                                 1.7s 

Error response from daemon: manifest for grafana/tempo:2.9.2 not found: manifest unknown: manifest unknown

myproject/
├── cmd/
│   └── myapp/
│       └── main.go                 # Точка входа
├── internal/
│   ├── app/
│   │   ├── handler/               # HTTP обработчики
│   │   ├── service/               # Бизнес-логика
│   │   └── middleware/            # Middleware
│   ├── domain/
│   │   ├── entity/                # Сущности
│   │   ├── valueobject/           # Value objects
│   │   └── repository/            # Интерфейсы репозиториев
│   └── infrastructure/
│       ├── persistence/           # Реализации репозиториев
│       ├── database/              # Подключение к БД
│       └── cache/                 # Кеш
├── pkg/
│   ├── config/                    # Конфигурация
│   ├── logger/                    # Логирование
│   └── util/                      # Утилиты
├── api/
│   ├── http/                      # HTTP спецификации
│   └── proto/                     # gRPC proto файлы
├── migrations/                    # Миграции БД
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── test/
│   ├── unit/                      # Юнит-тесты
│   └── integration/               # Интеграционные тесты
├── .gitignore
├── .env.example
├── Makefile                       # Автоматизация
├── go.mod                         # Модуль Go
├── go.sum                         # Зависимости
├── README.md                      # Документация
└── LICENSE                        # Лицензия

Минимальная версия (для старта):

todo-app/
├── cmd/
│   └── server/
│       └── main.go                # Запуск приложения
├── internal/
│   ├── domain/
│   │   └── task.go               # Сущности
│   ├── service/
│   │   └── task_service.go       # Бизнес-логика
│   └── handler/
│       └── http/
│           └── task_handler.go   # HTTP обработчики
├── pkg/
│   └── config/
│       └── config.go             # Конфигурация
├── go.mod
├── go.sum
├── .env
└── README.md

Для микросервиса:

auth-service/
├── cmd/
│   └── auth-server/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── user.go
│   │   ├── token.go
│   │   └── repository.go         # Интерфейсы
│   ├── service/
│   │   ├── auth_service.go
│   │   └── user_service.go
│   ├── handler/
│   │   └── http/
│   │       ├── auth_handler.go
│   │       ├── user_handler.go
│   │       └── routes.go
│   └── infrastructure/
│       ├── postgres/
│       │   └── user_repository.go
│       └── redis/
│           └── token_repository.go
├── api/
│   └── proto/
│       └── auth.proto            # gRPC контракт
├── migrations/
│   ├── 001_init.up.sql
│   └── 001_init.down.sql
├── docker-compose.yml
├── .env.example
├── go.mod
└── Makefile

Миграции

Добавьте зависимости в проект:

go get -u github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/lib/pq  # или ваш драйвер PostgreSQL

Структура проекта

your-project/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   └── ...
├── migrations/          # ← ЭТА ПАПКА
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_add_email_index.up.sql
│   └── 000002_add_email_index.down.sql
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum

Создание первой миграции

# Создайте миграцию (автоматически создаст up/down файлы)

migrate create -ext sql -dir migrations -seq create_users_table

migrations/
├── 000001_create_users_table.up.sql
└── 000001_create_users_table.down.sql

Полезные команды

# Создать миграцию

make migrate-create

# Введите: add_priority_to_tasks

# Применить все миграции

make migrate-up

# Откатить одну миграцию

make migrate-down

# Посмотреть текущую версию

make migrate-version

# Применить 2 миграции

make migrate-up-n

# Введите: 2

# Откатить на 3 миграции

make migrate-down-n

# Введите: 3

# Принудительно установить версию (если что-то сломалось)

make migrate-force

# Введите: 1

Для продакшена
Создайте скрипт деплоя:

#!/bin/bash

# deploy.sh

# Применить миграции

migrate -path ./migrations   
-database "$DATABASE_URL"   
up

# Проверить статус

migrate -path ./migrations   
-database "$DATABASE_URL"   
version

# Как делать тесты

устанавливаем Mockery
go install github.com/vektra/mockery/v2@latest

находим интерфейс, который хотим мокнуть и делаем вот так

```bash
//go:generate mockery --name=UserRepository --filename=user_repository_mock.go --output=../mocks --case=underscore
type UserRepository interface {
    GetUserById(id int) (domain.User, error)
}
```

далее генерируем моки
go generate ./...

# Из корня проекта

go test ./test/... -v

# Или только этот тест

go test ./test -run TestUserService_GetUserByID -v

# Для сбора логов из GO приложения нужно установить плагин для докера

```bash
docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
```

---

## Docker: полный стек (app, Postgres, MinIO+Nginx, Loki, Grafana)

Из корня проекта:

```bash
docker compose up -d
```

- **Приложение:** [http://localhost:3000](http://localhost:3000)  
- **Grafana:** [http://localhost:3001](http://localhost:3001) (admin / admin)  
- **MinIO S3:** [http://localhost:9000](http://localhost:9000) (через Nginx)  
- **MinIO Console:** [http://localhost:9001](http://localhost:9001) (через Nginx)  
- **Prometheus:** [http://localhost:9090](http://localhost:9090)  
- **Loki:** логи контейнера app пишутся через драйвер Loki (поля `level`, `app_name`, `app_version`).

### Что установить в новом проекте (Loki + Grafana)

1. Установить Loki‑драйвер Docker:

```bash
docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
```

1. В `docker-compose.yml` для сервиса app:

```yaml
logging:
  driver: loki
  options:
    loki-url: "http://host.docker.internal:3100/loki/api/v1/push"
    loki-pipeline-stages: |
      - json:
          expressions:
            level: level
            app_name: app_name
            app_version: app_version
      - labels:
          level:
          app_name:
          app_version:
```

1. Поднять `loki` и `grafana` в compose (как в этом проекте).
2. Открыть Grafana → **Explore** → источник **Loki** и выполнить запрос:

```
{app_name="interfaces"}
```

Конфиг приложения в контейнере: `CONFIG_PATH=/app/config.yml` (файл `docker/config.app.yml` с хостами `postgres`, `minio`, `redis`).

### Loki URL на Windows/Mac

На Docker Desktop лог-драйвер работает на хосте, поэтому `loki-url` должен быть доступен с хоста. По умолчанию используется:

```
http://host.docker.internal:3100/loki/api/v1/push
```

Если нужно, задайте в `.env`:

```
LOKI_URL=http://localhost:3100/loki/api/v1/push
```

### Миграции из контейнера

Сначала поднять стек (`docker compose up -d`), затем:

```bash
# Накатить все миграции
make migrate-up-docker

# Откатить одну
make migrate-down-docker

# Текущая версия
make migrate-version-docker
```

Без Make (переменные из `.env`):

```bash
docker compose --profile tools run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/db?sslmode=disable" up
docker compose --profile tools run --rm migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/db?sslmode=disable" down 1
```

---

## Метрики (Prometheus + Grafana)

### Что такое метрики

Метрики — это числовые показатели работы сервиса (кол-во запросов, время ответа, ошибки).  
Обычно их собирает Prometheus, а визуализирует Grafana.

### Как это работает у нас

1. Приложение отдаёт метрики по `GET /metrics`
2. Prometheus опрашивает `app:3000/metrics`
3. Grafana подключена к Prometheus и строит графики

### Где смотреть

- Prometheus: `http://localhost:9090`
- Grafana → Explore → источник **Prometheus**

Пример запросов:

```
process_cpu_seconds_total
go_goroutines
http_requests_total
```

### Как подключить метрики в новый проект

1. Добавить зависимость:

```bash
go get github.com/prometheus/client_golang@latest
```

1. Повесить обработчик `/metrics`:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

r.Handle("/metrics", promhttp.Handler())
```

1. Добавить middleware для подсчёта HTTP‑метрик (requests, latency):

```
// пример:
r.Use(custommiddleware.MetricsMiddleware())
```

1. Добавить Prometheus в compose:

```yaml
prometheus:
  image: prom/prometheus:v2.55.1
  ports:
    - "9090:9090"
  volumes:
    - ./docker/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
```

1. Добавить datasource Prometheus в Grafana (provisioning):

```yaml
- name: Prometheus
  type: prometheus
  access: proxy
  url: http://prometheus:9090
```

---

## Трейсы (OpenTelemetry + Tempo)

### Что такое трейсы

Трейсы показывают путь одного запроса через систему и время на каждом шаге.

### Как это работает у нас

1. Приложение отправляет трейсы в **OpenTelemetry Collector** (OTLP HTTP)
2. Collector пишет в **Tempo** (OTLP/HTTP на `tempo:4318`, версия `2.9.1`)
3. Grafana читает Tempo как datasource

### Где смотреть

- Grafana → Explore → источник **Tempo**
- Порт Tempo: `http://localhost:3200`

### Как подключить трейсы в новый проект

1. Добавить зависимости:

```bash
go get go.opentelemetry.io/otel@latest
go get go.opentelemetry.io/otel/sdk@latest
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@latest
go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@latest
```

1. Инициализировать OTel при старте приложения:

```go
shutdownTelemetry, err := telemetry.Init(ctx, telemetry.LoadConfigFromEnv())
if err != nil { /* лог */ }
defer shutdownTelemetry(context.Background())
```

1. Обернуть HTTP‑router:

```go
otelHandler := otelhttp.NewHandler(router, "http-server")
srv.Run(port, otelHandler)
```

1. Добавить в compose:

```yaml
tempo:
  image: grafana/tempo:latest
  ports:
    - "3200:3200"

otel-collector:
  image: otel/opentelemetry-collector-contrib:latest
  ports:
    - "4317:4317"
    - "4318:4318"
  volumes:
    - ./docker/otel-collector/otel-collector.yml:/etc/otel-collector.yml:ro
```

1. В Grafana добавить datasource Tempo:

```yaml
- name: Tempo
  type: tempo
  access: proxy
  url: http://tempo:3200
```

### Переменные окружения

```
ENABLE_TRACING=true
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4318
```

### Проверка трейсов

1. Отправить запрос:

```bash
curl http://localhost:3000/api/trace-test
```

1. В Grafana → **Explore** → источник **Tempo**.
2. В поиске выбрать `Service Name = interfaces`.

### Как правильно добавлять спаны (manual instrumentation)

Автоматическая обёртка HTTP уже создаёт входной span.  
Чтобы видеть детали внутри бизнес‑логики, добавляйте собственные спаны:

```go
import "go.opentelemetry.io/otel"

func (s *Service) DoWork(ctx context.Context) error {
	tracer := otel.Tracer("service")
	ctx, span := tracer.Start(ctx, "Service.DoWork")
	defer span.End()

	// бизнес‑логика
	if err := s.repo.Save(ctx, data); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
```

Лучшие практики:

- **Не добавляйте `request_id` в метрики** (высокая кардинальность) — используйте трейсы.
- **Называйте span по смыслу**, а не по техническим деталям.
- **Ставьте `RecordError`** при ошибках и завершайте span `defer span.End()`.

---

## Профилирование (Grafana Pyroscope)

`pprof` — это **снимок на момент времени** (вы снимаете профиль вручную).  
Pyroscope — это **непрерывный профилировщик**, он постоянно снимает pprof‑профили и отправляет их в Grafana, чтобы было видно **почему** функция тормозит (CPU, alloc, mutex, block и т.д.).

### Как это собирает данные

- в приложении запускается агент (`pyroscope-go`);
- агент периодически снимает **pprof‑профили** из Go runtime (CPU, alloc, goroutines, mutex, block);
- профили сэмплируются (например CPU по времени) и имеют небольшой оверхед;
- агент отправляет их в Pyroscope (сервер) по HTTP;
- Grafana показывает flamegraph/таблицы, где видно «почему медленно».

### Что добавить в новый проект

1. Добавить Pyroscope в `docker-compose.yml`:

```yaml
pyroscope:
  image: grafana/pyroscope:1.18.1
  ports:
    - "4040:4040"
```

2. Добавить datasource в Grafana:

```yaml
- name: Pyroscope
  type: grafana-pyroscope-datasource
  access: proxy
  url: http://pyroscope:4040
```

3. Убедиться, что плагин Pyroscope установлен в Grafana:

```yaml
grafana:
  environment:
    GF_PLUGINS_PREINSTALL: grafana-pyroscope-datasource
```

После этого пересоздать Grafana:

```bash
docker compose up -d --force-recreate grafana
```

4. Подключить агент в коде:

```go
profiler, err := profiling.Start(profiling.LoadConfigFromEnv())
if err != nil {
	logger.Error("failed to init profiling: " + err.Error())
} else if profiler != nil {
	defer profiler.Stop()
}
```

5. Переменные окружения:

```
ENABLE_PROFILING=true
PYROSCOPE_URL=http://pyroscope:4040
```

### Проверка

1. Дайте нагрузку на API.
2. Grafana → **Explore** → источник **Pyroscope**.
3. Выберите `Service` (имя приложения) и нужный интервал времени.

---

## План развития (TODO)

- Самописный DI‑контейнер + инструкция использования + пример моков
- Middleware метрик (детализация по статусам и группам endpoint)
- Бизнес‑метрики и метрики ошибок
- Рефакторинг структуры под Clean Architecture

---

## Самописный DI (инструкция + пример моков)

### Идея

Вместо магического DI‑контейнера делаем **простую сборку зависимостей** руками.  
Это прозрачно, легко тестируется и не мешает мокам.

### Мини‑пример

```go
// contracts
type UserRepo interface {
	GetByID(ctx context.Context, id int) (User, error)
}

type UserService struct {
	repo UserRepo
}

func NewUserService(repo UserRepo) *UserService {
	return &UserService{repo: repo}
}
```

```go
// container
type Container struct {
	Config  *config.Config
	Logger  *slog.Logger
	Repo    UserRepo
	Service *UserService
	Handler *handler.Handler
}

func BuildContainer() *Container {
	cfg := pkg.MainConfig
	log := pkg.MainLogger
	repo := persistence.NewUserRepository()
	svc := NewUserService(repo)
	h := handler.NewHandler(service.NewService(repository.NewRepository()), log, nil)

	return &Container{
		Config:  cfg,
		Logger:  log,
		Repo:    repo,
		Service: svc,
		Handler: h,
	}
}
```

### Использование в реальном коде (main.go)

```go
func main() {
	container, err := di.Build(di.Deps{})
	if err != nil {
		log.Fatal(err)
	}

	srv := new(server.Server)
	// Без OTel:
	err := srv.Run(container.Config.HttpServer.Port, container.Handler.InitRoutes())
	if err != nil {
		container.Logger.Error(err.Error())
	}
}
```

Реальный контейнер лежит в `internal/di/container.go`.

Если нужен OTel, то **вы оборачиваете уже готовый handler**, контейнер тут ни при чём:

```go
otelHandler := otelhttp.NewHandler(container.Handler.InitRoutes(), "http-server")
_ = srv.Run(container.Config.HttpServer.Port, otelHandler)
```

### Важно про DI‑контейнер

- Контейнер используется **только в composition root** (обычно `main.go`), чтобы собрать зависимости.
- Хендлеры **не получают контейнер** — они получают конкретные зависимости.
- Инициализация происходит в `di.Build`: после `Build` все поля (`Handler`, `Services`, `Repo`, `Minio`, `Logger`) уже готовы.
- OTel **не участвует** в DI: он только оборачивает готовый `handler`.

Почему не передавать контейнер в хендлер:
- это превращает контейнер в **Service Locator** (антипаттерн);
- зависимости становятся неявными;
- тестировать сложнее.

### Если контейнер нужен «ещё где‑то»

Правильный путь: передавайте **только нужные части**:

- фоновые задачи → `container.Services`, `container.Logger`;
- CLI/cron → `container.Repo`, `container.Config`;
- отдельный модуль → `container.Logger` и/или нужный сервис.

Если всё же передаёте контейнер в верхний уровень (например `NewApp(container)`), то внутри него лучше сразу разложить по конкретным полям.

### Как передавать зависимости (Overrides)

Чтобы легко мокать и подменять зависимости, делайте сборку с параметрами:

```go
// di.Deps позволяет переопределить любые зависимости
container, _ := di.Build(di.Deps{
	Logger: customLogger,
	Repo:   mockRepo,
})
```

В тестах можно передавать mock‑репозиторий:

```go
container, _ := di.Build(di.Deps{Repo: mockRepo})
```

### Как мокать в тестах

```go
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, id int) (User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(User), args.Error(1)
}

func TestUserService(t *testing.T) {
	repo := new(MockUserRepo)
	repo.On("GetByID", mock.Anything, 1).Return(User{ID: 1}, nil)

	svc := NewUserService(repo)
	_, _ = svc.GetByID(context.Background(), 1)

	repo.AssertExpectations(t)
}
```

Плюсы:

- никакой магии
- зависимости явно видно
- мокается точечно любой слой

---

## Отладка в GoLand (в т.ч. без локального Go)

### Вариант A: зависимости в Docker, приложение локально

Поднять только зависимости (postgres, minio, loki и т.д.), приложение запускать и отлаживать в GoLand на хосте. В `.env`/`config.yml` использовать `localhost` и проброшенные порты.

### Вариант B: всё в Docker + удалённый отладчик (Delve) + hot reload

1. Запустить dev-стек:
  ```bash
   docker compose -f docker-compose.dev.yml up --build -d app-dev-live
  ```
   (сервис `app-dev-live` использует `docker/Dockerfile.live` и `.air.toml`.)
   Dev-стек также поднимает `minio + nginx`, `loki` и `grafana` (как в проде), чтобы окружение было одинаковым.
2. В GoLand: **Run → Edit Configurations → + → Go Remote**. Host: `localhost`, Port: `40000`. Сохранить.
3. **Run → Debug** выбранной конфигурации Go Remote — подключение к процессу в контейнере.

Так можно отлаживать без установленного Go на машине: сборка и запуск в контейнере, отладка через Go Remote из GoLand, а `air` делает hot reload.

### Чеклист для нового проекта (подключение отладки)

1. Добавить `Dockerfile.dev`:
  - сборка с debug-символами: `-gcflags="all=-N -l"`.
  - установить Delve (`dlv`).
  - запуск: `dlv exec /app/server --headless --listen=:40000 --api-version=2 --accept-multiclient`.
2. Добавить `docker-compose.dev.yml`:
  - пробросить порт `40000`.
  - поднять зависимости (postgres/minio/и т.д.).
  - сервис `app-dev` с `Dockerfile.dev`.
3. В GoLand создать конфигурацию **Go Remote**:
  - Host: `localhost`
  - Port: `40000`
  - затем **Debug**.
4. Запуск dev-стека:
  ```bash
   docker compose -f docker-compose.dev.yml up --build -d app-dev
  ```

### Как видеть изменения в коде

В `app-dev-live` изменения подхватываются автоматически — `air` пересобирает бинарник и перезапускает процесс под `dlv`. Пересборка образа не нужна.

### Hot reload + debug (без пересборки образа)

Готово в проекте:

- `docker/Dockerfile.live` — образ с Go, `dlv` и `air`
- `.air.toml` — настройка hot reload + Delve
- сервис `app-dev-live` в `docker-compose.dev.yml`

Запуск:

```bash
docker compose -f docker-compose.dev.yml up --build -d app-dev-live
```

Далее подключиться из GoLand (**Go Remote**):

- Host: `localhost`
- Port: `40000`

Примечания:

- используйте **либо** `app-dev`, **либо** `app-dev-live` (порты одинаковые).
- при изменениях кода `air` автоматически пересобирает бинарник и перезапускает процесс под `dlv`.
- если GoLand временно теряет соединение, просто нажмите **Debug** еще раз.
- в GoLand лучше выбрать **On disconnect → Leave it running**. Если нажать **Stop**, GoLand завершает `dlv`, и контейнер нужно перезапускать или сделать любое изменение файла, чтобы `air` перезапустил процесс.
- на Windows/Docker Desktop иногда не приходят файловые события, поэтому в `.air.toml` включен `poll = true` (проверка изменений по таймеру).

### Чеклист для нового проекта (hot reload + debug)

1. Добавить `docker/Dockerfile.live` (Go + `dlv` + `air`).
2. Добавить `.air.toml` с:
  - `go build -gcflags=all=-N -l`
  - `dlv exec ... --headless --listen=:40000 --accept-multiclient`
  - `poll = true` (важно для Docker Desktop/Windows)
3. В `docker-compose.dev.yml` создать сервис `app-dev-live`:
  - `volumes: .:/app`
  - проброс портов `3000` и `40000`
  - `command: air -c .air.toml` (или `CMD` в Dockerfile)
4. В GoLand создать **Go Remote** и подключиться к `localhost:40000`.

### Что именно добавляли для отладки (шаблон)

- `docker/Dockerfile.live`
- `.air.toml`
- сервис `app-dev-live` в `docker-compose.dev.yml` (с монтированием исходников)
- настройки GoLand: **Go Remote** (localhost:40000)

