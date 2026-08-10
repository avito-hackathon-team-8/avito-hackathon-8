# Коробыш — puppeteer

Фоновый Go-сервис игрового слоя для Авито. Puppeteer назначает пользователям ежедневные задания, пересчитывает месячный лидерборд, выдаёт награды победителям завершившегося сезона и предоставляет backend внутренний API для создания заданий по запросу.

## Запуск

Для локального запуска нужны Go 1.25+ и PostgreSQL с применёнными миграциями из `../migrations/`. Puppeteer и backend используют одну базу данных.

```bash
cd puppeteer

export DATABASE_URL=postgres://hackathon:change-me@localhost:5432/hackathon?sslmode=disable
export INTERNAL_SERVICE_TOKEN="$(openssl rand -hex 32)"

go run .
```

Сервис будет доступен на `http://localhost:8091`. Миграции он автоматически не применяет.

Рекомендуемый запуск всего проекта из корня репозитория:

```bash
cp .env.example .env
make up
```

Перед запуском замените значения `JWT_SECRET` и `INTERNAL_SERVICE_TOKEN` в `.env` на разные случайные строки длиной не менее 32 символов. `make up` применит миграции и запустит PostgreSQL, backend, puppeteer и frontend.

Проверка состояния сервиса:

```bash
curl http://localhost:8091/health
```

Основные переменные окружения:

```dotenv
DATABASE_URL=postgres://user:password@localhost:5432/database?sslmode=disable
INTERNAL_SERVICE_TOKEN=<строка длиной не менее 32 символов>
PUPPETEER_HTTP_ADDRESS=:8091
PUPPETEER_POLL_INTERVAL=10m
TASK_DEFINITIONS_CONFIG=../config/task_definitions.yaml
LEADERBOARD_REWARDS_CONFIG=../config/leaderboard_rewards.yaml
```

`PUPPETEER_POLL_INTERVAL` задаётся в формате Go duration, не может быть меньше одной секунды и управляет периодичностью пересчёта лидерборда.

## Технологии

- Go 1.25 и стандартный `net/http`;
- PostgreSQL 17 и GORM;
- YAML-конфигурация заданий и наград лидерборда;
- PostgreSQL advisory locks для координации фоновых задач;
- Docker Compose;
- unit- и integration-тесты на Go;
- golangci-lint.

## Структура

```text
puppeteer/
  main.go          # сборка зависимостей, запуск worker и graceful shutdown
  internal/
    config/         # переменные окружения и их валидация
    database/       # подключение к PostgreSQL
    health/         # health-check и внутренний HTTP API
    jobs/           # scheduler, задания, лидерборд и загрузка YAML-каталогов
    models/         # используемые worker модели GORM
```

Конфигурация механик хранится вне сервиса:

- [`config/task_definitions.yaml`](../config/task_definitions.yaml) — варианты ежедневных заданий, слоты, награды и уровни открытия;
- [`config/leaderboard_rewards.yaml`](../config/leaderboard_rewards.yaml) — награды за первые три места и срок их действия.

## Фоновые задачи

### Ежедневные задания

При старте Puppeteer синхронизирует YAML-каталог заданий с базой данных. Затем задача `distribute-daily-tasks` запускается при старте и в полночь по UTC:

- помечает незавершённые задания предыдущего дня как `EXPIRED`;
- выбирает по одному заданию для каждого из четырёх слотов;
- учитывает уровень питомца и блокирует пока недоступные задания;
- подбирает варианты по близости категорий к интересам пользователя;
- использует стабильное разрешение ничьей для пользователя, даты и слота.

Задания назначаются только подтверждённым пользователям. Уже занятые слоты и существующие назначения повторно не создаются.

### Лидерборд

Задача `calculate-leaderboard` запускается при старте и повторяется через `PUPPETEER_POLL_INTERVAL`. Она формирует рейтинг подтверждённых пользователей за текущий календарный месяц по сумме заработанных листьев.

После завершения месяца предыдущий сезон финализируется один раз. Пользователи на первых трёх местах получают награды из YAML-каталога; повторный запуск не выдаёт ту же награду повторно.

## HTTP API

`GET /health` проверяет соединение с PostgreSQL и возвращает время и результат последнего запуска каждой фоновой задачи. При недоступной базе данных endpoint отвечает `503 Service Unavailable`.

Backend может потребовать немедленно назначить задания конкретному пользователю:

```text
POST /internal/v1/users/{userId}/daily-tasks/ensure
X-Service-Token: <INTERNAL_SERVICE_TOKEN>
```

Успешный запрос отвечает статусом `204 No Content`. Внутренний endpoint защищён тем же сервисным токеном, который настроен в backend.

## Конкурентность и отказоустойчивость

Каждая фоновая задача выполняется внутри транзакции PostgreSQL. Advisory lock не позволяет нескольким экземплярам Puppeteer одновременно выполнять одну задачу, а таблица запусков гарантирует, что ежедневная задача будет обработана не более одного раза за дату.

Назначения заданий и сезонные награды дополнительно защищены уникальными ограничениями. Ошибка фоновой задачи сохраняется в состоянии health-check, после чего scheduler продолжает работу и повторит периодические задачи по расписанию.

При `SIGINT` или `SIGTERM` сервис прекращает scheduler, завершает HTTP-сервер и закрывает соединение с базой данных.

## Тесты и проверки

Основные тесты запускаются без PostgreSQL; integration-тест advisory lock пропускается, если не задан `TEST_DATABASE_URL`:

```bash
cd puppeteer
go test ./...
go test -race ./...
```

Запуск integration-теста с PostgreSQL:

```bash
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/puppeteer_test?sslmode=disable \
  go test ./internal/jobs -run TestRunnerRunLockedUsesAdvisoryLock -count=1 -v
```

Запуск линтера из корня репозитория:

```bash
make lint-puppeteer
```
