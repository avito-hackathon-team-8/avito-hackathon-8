# Коробыш

Игровой слой поверх Авито: пользователь получает виртуального питомца, выполняет
ежедневные задания и получает листья за активность. Листья
повышают уровень питомца, открывают персональные награды и учитываются в
ежемесячном лидерборде.



## Архитектура

```text
frontend (React/Vite, :3000)
        |
        v
backend (Go API, :8090) -----> postgres (:5432)
        |
        +---- internal HTTP ----> puppeteer (:8091)
                                  задания, лидерборд, health-check
```


| Сервис | Назначение | Локальный адрес |
| --- | --- | --- |
| `postgres` | хранение пользователей, питомцев, листьев, заданий и наград | `localhost:5432` |
| `backend` | HTTP API, авторизация и игровые операции | `http://localhost:8090` |
| `puppeteer` | фоновые задания: назначение задач и расчёт лидерборда | `http://localhost:8091` |
| `frontend` | веб-интерфейс | `http://localhost:3000` |

## Быстрый запуск

### Требования

- Docker Engine с Docker Compose v2;
- свободные порты `3000`, `5432`, `8090` и `8091`;
- для запуска тестов без Docker: Go 1.25+, Node.js 24+ и npm.

### Запуск всего проекта

Из корня репозитория:

```sh
cp .env.example .env
```

Для локальной разработки можно оставить `EMAIL_MODE=log`. Сгенерируйте два
секрета длиной не менее 32 символов:

```sh
openssl rand -hex 32
```

Вставьте разные значения в `.env`:

```dotenv
JWT_SECRET=<значение_1>
INTERNAL_SERVICE_TOKEN=<значение_2>
```

Запустите сервисы:

```sh
make up
```

Эта команда собирает Docker-образы и запускает PostgreSQL, backend, puppeteer и
frontend.

Остановить проект:

```sh
make down
```

Дополнительные команды:

```sh
make ps       # состояние сервисов
make logs     # логи всех сервисов
make restart  # перезапуск без пересборки
make build    # пересборка образов
```

## Проверка после запуска

Проверка backend:

```sh
curl http://localhost:8090/api/health
```

Команда проверяет, что backend запущен и готов принимать запросы.

Проверка puppeteer:

```sh
curl http://localhost:8091/health
```

Откройте интерфейс: <http://localhost:3000>.

## API

Все пользовательские endpoint-ы, кроме health-check и методов авторизации,
требуют заголовок:

```text
Authorization: Bearer <JWT>
```

Внутренние endpoint-ы используют заголовок:

```text
X-Service-Token: <INTERNAL_SERVICE_TOKEN>
```

Все даты передаются в формате `YYYY-MM-DD`, а временные метки — в формате
RFC3339 и UTC.

### Сервис и авторизация

| Метод | Endpoint | Описание |
|---|---|---|
| `GET` | `/api/health` | Проверяет доступность и готовность backend-сервиса. |
| `POST` | `/api/app/auth/request-otp` | Создаёт пользователя при необходимости и отправляет одноразовый код на email. |
| `POST` | `/api/app/auth/verify-otp` | Проверяет одноразовый код и создаёт JWT-сессию пользователя. |
| `GET` | `/api/app/auth/me` | Определяет текущего пользователя по JWT. |


### Питомец

| Метод | Endpoint | Описание |
|---|---|---|
| `GET` | `/api/v1/pet` | Получает имя, уровень и текущий баланс листьев питомца. |
| `PATCH` | `/api/v1/pet` | Изменяет имя питомца. |
| `GET` | `/api/v1/pet/levels` | Получает список уровней питомца и доступных наград. |
| `GET` | `/api/v1/pet/ws` | Открывает WebSocket-соединение для обновлений прогресса питомца. |
| `POST` | `/api/v1/pet/level-rewards/{rewardId}/claim` | Забирает награду, связанную с достигнутым уровнем питомца. |
| `POST` | `/api/v1/pet/chests/open` | Открывает сундук и списывает необходимое количество листьев. |

Для `PATCH /api/v1/pet` передаётся новое имя питомца, а для WebSocket токен можно
передать через заголовок `Authorization` или query-параметр `token`.

### Награды

| Метод | Endpoint | Описание |
|---|---|---|
| `GET` | `/api/app/rewards` | Получает персональные награды пользователя, сгруппированные по категориям. |
| `GET` | `/api/app/rewards/{rewardId}` | Получает подробную информацию о конкретной награде пользователя. |
| `POST` | `/api/app/rewards/{rewardId}/redeem` | Использует персональную награду пользователя. |

### Ежедневные задания

| Метод | Endpoint | Описание |
|---|---|---|
| `GET` | `/api/v1/tasks` | Получает четыре задания текущего дня и их прогресс. |
| `GET` | `/api/v1/tasks/progress` | Получает количество выполненных заданий за текущий день. |
| `POST` | `/api/v1/tasks/record` | Записывает действия пользователя для обновления прогресса заданий. |
| `POST` | `/api/v1/tasks/{taskId}/claim` | Забирает награду за выполненное задание. |

Метод `POST /api/v1/tasks/record` принимает список событий с полями `taskId`,
`type` и `count`;

Поддерживаемые типы заданий: `VIEW_LISTINGS`, `ADD_TO_FAVORITES`,
`PUBLISH_LISTING`, `BOOST_LISTING`, `LEAVE_REVIEW`, `COMPLETE_DEAL` и
`ORDER_WITH_DELIVERY`.

### Еженедельный вход

| Метод | Endpoint | Описание |
|---|---|---|
| `GET` | `/api/v1/weekly-login` | Получает состояние наград за вход в текущей календарной неделе. |
| `POST` | `/api/v1/weekly-login/activity` | Засчитывает активность пользователя за текущий день в UTC. |
| `POST` | `/api/v1/weekly-login/claim` | Забирает доступную награду за вход в указанный день. |

Метод `POST /api/v1/weekly-login/activity` вызывается без тела запроса, а
`POST /api/v1/weekly-login/claim` принимает дату в формате `YYYY-MM-DD`.

### Лидерборд

| Метод | Endpoint | Описание |
|---|---|---|
| `GET` | `/api/v1/leaderboard` | Получает текущий топ-10 игроков за календарный месяц. |
| `GET` | `/api/v1/leaderboard/me` | Получает позицию текущего пользователя в лидерборде. |

### Внутренние события

| Метод | Endpoint | Описание |
|---|---|---|
| `POST` | `/api/internal/v1/users/{userId}/events` | Принимает подтверждённые события активности от доверенного сервиса Авито. |

Метод принимает список событий с полями `eventId`, `type`, `count` и
`occurredAt`.

## Разработка и тесты

### Backend

```sh
cd backend
go test ./...
go test -race ./...
```

### Puppeteer

```sh
cd puppeteer
go test ./...
```

### Frontend

```sh
cd frontend
npm ci
npm run build
npm run lint
npm run lint:styles
```

### Все проверки

```sh
make lint
```

`make lint` запускает frontend-проверки и golangci-lint для backend в Docker.

При таком режиме backend должен быть доступен на `http://localhost:8090`.

## Структура проекта

```text
backend/
  main.go                 # сборка зависимостей и HTTP-сервер
  internal/auth/          # OTP, JWT и аутентификация
  internal/database/      # PostgreSQL, GORM и миграции
  internal/events/        # доверенные события активности
  internal/handlers/      # HTTP handlers и маршруты
  internal/leaves/        # журнал листьев и повышение уровня
  internal/pet/           # питомец, награды уровней и WebSocket
  internal/rewards/       # персональные награды и redemption
  internal/tasks/         # чтение и выдача дневных заданий
  internal/weekly_login/  # недельные награды за вход
puppeteer/
  main.go                 # фоновые jobs и health endpoint
config/
  task_definitions.yaml   # варианты заданий
  leaderboard_rewards.yaml# награды топ-3
frontend/                 # React/Vite интерфейс
docs/
  openapi.yaml            # контракт API
  logic.md                # продуктовая логика MVP
compose.yaml              # локальный production-like запуск
```
