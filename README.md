# AIGateway

Сервис приема событий с асинхронным обогащением через AI (OpenAI-совместимый API).
Событие принимается по REST, дедуплицируется через Redis, сохраняется в Postgres
и уходит в конвейер воркеров, которые обращаются к AI и пишут результат обратно
в БД. Прочитать обогащение можно по REST или gRPC.

## Архитектура

```
        POST /events                 GET /events/:id
              │                            │
              ▼                            ▼
        ┌───────────┐               ┌───────────┐
        │  gin REST │               │  gin REST │
        └─────┬─────┘               └─────┬─────┘
              │                            │
              ▼                            ▼
        EventsService                 EventsRepo
     (dedup через Redis SETNX)          (Postgres)
              │
              ▼
        worker.Pipeline
     ┌─────────────────────┐
     │ Producer → in chan  │
     │ N x Worker (fan-out)│──► aiclient (OpenAI-compatible, rate-limit через Redis)
     │ Merge (fan-in)      │
     │ Save → EnrichmentsRepo
     └─────────────────────┘

        GetEnrichment(id)
              │
              ▼
      gRPC EnrichmentServer ──► EnrichmentsService ──► EnrichmentsRepo (Postgres)
```

- **REST** (`gin`) — прием событий и чтение событий/обогащений.
- **gRPC** — чтение обогащения по id (`internal/transport/grpc`).
- **worker.Pipeline** — fan-out/fan-in конвейер: N воркеров разбирают общий канал
  `Job`, ходят в AI-клиент, результаты сливаются в один канал и пишутся в БД.
  Graceful shutdown: канал `Job` закрывается после остановки REST/gRPC,
  воркеры дорабатывают очередь, а `sync.WaitGroup` (`gs`) дожидается,
  пока последний результат не запишется в БД.
- **aiclient** — обертка над OpenAI SDK с рейт-лимитом на основе Redis
  (счетчик с TTL, `INCR` + `EXPIRE`).
- **Postgres** (`sqlx` + `pgx`) — хранилище событий (`events`) и результатов
  обогащения (`enrichments`).
- **Redis** — дедупликация входящих событий по `external_id` (`SETNX`) и
  счетчик для рейт-лимита AI-клиента.

## Структура проекта

```
cmd/main.go                          точка входа, сборка зависимостей, graceful shutdown
config/config.yml                    конфиг для локального запуска (localhost)
config/config.docker.yml             конфиг для docker-compose (хосты postgres/redis)
internal/config                      загрузка конфига (cleanenv + .env)
internal/db                          подключение к Postgres
internal/models                      Event, Enrichment
internal/repository                  доступ к Postgres (events, enrichments)
internal/service                     бизнес-логика (дедуп, обертки над репозиториями)
internal/aiclient                    клиент AI + Redis rate-limiter
internal/worker                      fan-out/fan-in конвейер обогащения
internal/transport/rest/handlers     REST-хендлеры (gin)
internal/transport/grpc/server       gRPC-сервер (EnrichmentService)
internal/transport/grpc/pb           сгенерированный код из proto
migrations                           SQL-миграции (events, enrichments)
```

## Конфигурация

Настройки читаются из `config/config.yml` (`cleanenv`), секреты — из `.env`
или переменных окружения (`DB_PASS`, `API_KEY` — обязательны).

```yaml
rest_service: { host, port, timeout }
grpc_service: { port, timeout }
db:           { host, port, name, user, ssl_mode }   # + DB_PASS из env
redis:        { addr }
ai_client:    { base_url }                            # + API_KEY из env
pipeline:     { worker_count, buffer_size }
```

`.env` — не коммитится (см. `.gitignore`), пример:

```
DB_PASS=postgres
API_KEY=sk-...
```

## Запуск локально

```bash
go run ./cmd
```

Требуется поднятый Postgres и Redis с адресами из `config/config.yml`
(по умолчанию `localhost:5432` и `localhost:6379`) и применённые миграции
из `migrations/` (например, через `migrate` CLI).

## Запуск в Docker

```bash
docker compose up --build
```

Поднимает `app` (REST на `:8080`, gRPC на `:7070`), `postgres:16-alpine` и
`redis:7-alpine`. `DB_PASS`/`API_KEY` подтягиваются из `.env` в корне проекта.
Внутри контейнера сервис использует `config/config.docker.yml`
(хосты `postgres`/`redis` вместо `localhost`) — файл подмонтирован поверх
запечённого в образ `config.yml`.

Миграции контейнер не применяет автоматически — прогнать их нужно отдельно
(например, `migrate -path migrations -database "postgres://..." up`).

## API

### REST

- `POST /events` — принять событие.
  ```json
  { "external_id": "abc-123", "payload": "текст для обогащения" }
  ```
  При повторном `external_id` (в пределах TTL дедупа) вернет `400` с
  `duplicate event`. При успехе событие сразу уходит в конвейер обогащения.

- `GET /events/:id` — получить событие по внутреннему `id`.

### gRPC

`EnrichmentService.GetEnrichment(id int64) → { id, event_id, response, model, created_at }`
— читает результат AI-обогащения по id из `enrichments`.