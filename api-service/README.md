# Коробыш — api-service

API service игрового слоя для Авито. Сервис авторизует пользователей, хранит состояние питомца, обрабатывает задания и события активности, начисляет листья, выдаёт награды и отправляет обновления клиенту по WebSocket.

## Запуск

Запуск выполняется вместе с остальными сервисами через Docker Compose из корня
репозитория:

```bash
cp .env.example .env
make up
```

Перед запуском замените значения `JWT_SECRET` и `INTERNAL_SERVICE_TOKEN` в `.env` на разные случайные строки длиной не менее 32 символов. `make up` применит миграции и запустит PostgreSQL, api-service, daily-tasks-service и frontend.

API service будет доступен на `http://localhost:8090`.

Проверка состояния сервиса:

```bash
curl http://localhost:8090/api/health
```

При запуске через Docker Compose Swagger UI доступен на
`http://localhost:8090/swagger`. Полный контракт API находится в
[`docs/openapi.yaml`](../docs/openapi.yaml).

Основные настройки Docker-окружения:

```dotenv
JWT_SECRET=<строка длиной не менее 32 символов>
INTERNAL_SERVICE_TOKEN=<строка длиной не менее 32 символов>
JWT_TTL=24h
HTTP_ADDRESS=:8090
DAILY_TASKS_INTERNAL_URL=http://daily-tasks-service:8091
LEVEL_REWARDS_CONFIG=/app/config/level_rewards.yaml
```

## Технологии

- Go 1.25 и стандартный `net/http`;
- PostgreSQL 17 и GORM;
- JWT с подписью HS256;
- Gorilla WebSocket;
- YAML-конфигурация наград;
- Docker Compose и версионированные SQL-миграции;
- unit- и black-box e2e-тесты на Go;
- golangci-lint.

## Структура

```text
api-service/
  main.go            # сборка зависимостей, HTTP-сервер и graceful shutdown
  internal/
    auth/             # авторизация по OTP и выпуск JWT
    chest/            # покупка и открытие сундуков
    config/           # переменные окружения и их валидация
    daily_report/     # ежедневная сводка и подписки на обновления
    database/         # подключение к PostgreSQL
    events/           # обработка доверенных событий активности
    handlers/         # маршруты, HTTP- и WebSocket-обработчики
    models/           # модели предметной области и GORM
    pet/              # питомец, уровни, листья и награды уровней
    reward_catalog/   # загрузка каталога наград из YAML
    rewards/          # выдача и использование персональных наград
    tasks/            # ежедневные задания и их прогресс
    testutil/         # тестовые заглушки
    weekly_login/     # недельная активность и награды за вход
  test/               # e2e-сценарии публичного API
```

## API и авторизация

Публичные методы сгруппированы по префиксам `/api/app` и `/api/v1`. Все пользовательские методы, кроме health-check и входа, принимают JWT в заголовке:

```text
Authorization: Bearer <JWT>
```

Внутренний endpoint `/api/internal/v1/users/{userId}/events` предназначен для доверенных сервисов и защищён отдельным токеном:

```text
X-Service-Token: <INTERNAL_SERVICE_TOKEN>
```

В MVP запрос кода создаёт пользователя при необходимости, а проверка принимает любой непустой OTP. После входа API предоставляет операции с питомцем, заданиями, наградами, сундуками, недельной активностью, ежедневной сводкой и лидербордом.

## WebSocket

JWT для WebSocket можно передать в заголовке `Authorization` или query-параметре `token`.

- `/api/v1/pet/ws` сначала отправляет текущее состояние питомца, затем события `PET_PROGRESS_UPDATED` после изменения баланса или уровня;
- тот же WebSocket отправляет `PET_STATE_UPDATED` после поглаживания или кормления;
- `/api/v1/daily-report/ws` сначала отправляет актуальную сводку, затем события `DAILY_REPORT_UPDATED` при изменениях и наступлении нового дня по UTC.

Соединения поддерживаются ping/pong-сообщениями. Сервер ограничивает размер входящего сообщения и корректно освобождает подписки после отключения клиента.

`POST /api/v1/pet/care` поддерживает необязательный `Idempotency-Key` длиной до
128 символов. Для нескольких реплик api-service `API_SERVICE_INSTANCE_ID` должен быть
уникальным; если переменная пуста, используется hostname контейнера.

`/api/health/live` проверяет процесс, `/api/health/ready` — PostgreSQL, Kafka и
pet-state-service. `/api/health` оставлен как alias readiness.

## Транзакции и целостность данных

Начисление и списание листьев, повышение уровня, открытие сундука и выдача награды выполняются транзакционно. Баланс питомца блокируется на время изменения, а уникальный ключ операции защищает начисления и списания от повторного применения.

Доверенные события активности также имеют идентификатор события, поэтому их повторная доставка не должна повторно менять прогресс пользователя. После успешной транзакции сервис публикует WebSocket-обновления питомца и ежедневной сводки.

## Тесты и проверки

Unit-тесты не требуют запущенного PostgreSQL:

```bash
cd api-service
go test ./...
go test -race ./...
```

Black-box e2e-тесты обращаются к сервисам, запущенным через Docker Compose. Команды выполняются из корня репозитория:

```bash
make up
make test-api-service-e2e
```

Запуск линтера из корня репозитория:

```bash
make lint-api-service
```