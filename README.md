# Avito Hackathon Тамагочи - Коробыш
Разработано командой номер 8

## Стек

- Frontend: React, TypeScript, Vite
- Backend: Go
- База данных: PostgreSQL + GORM
- Авторизация: одноразовый код из email + JWT
- Запуск: Docker Compose

## Запуск

Скопируйте пример .env, при необходимости измените значения и запустите:

```sh
cp .env.example .env
make up
```
Приложение доступно: `http://localhost:3000`

## Email

По умолчанию `EMAIL_MODE=log`: одноразовый код выводится в лог backend для
локальной разработки:

```sh
make logs
```

В проде `EMAIL_MODE=smtp` и заполнить в `.env`:

```dotenv
EMAIL_FROM=no-reply@example.com
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=no-reply@example.com
SMTP_PASSWORD=your-password
JWT_SECRET=your-secret-key-here
```

## Структура backend

```text
backend/
├── internal/
│   ├── auth/       # OTP, JWT и сценарии авторизации
│   ├── database/   # PostgreSQL/GORM
│   ├── email/      # SMTP
│   ├── handlers/   # HTTP routes и JSON handlers
│   ├── models/     # GORM-модели
│   ├── rewards/    # выдача, список и использование наград
│   └── tasks/      # дневные задачи и YAML-конфигурация вариантов
└── main.go         # сборка зависимостей и запуск сервера
```

## Команды

```sh
make help
make logs
make lint
make feature NAME=auth
make down
```

## Ежедневные задачи

Проверка сценария: получить задачи дня, записать подтверждённые события,
дождаться статуса `COMPLETED` и забрать награду через `claim`.

```text
GET  /api/v1/tasks?level=1
GET  /api/v1/tasks/progress?level=1
POST /api/v1/tasks/record
POST /api/v1/tasks/{taskId}/claim
```

В текущем MVP уровень питомца и начисление листьев являются заглушками до
подключения соответствующего сервиса.
