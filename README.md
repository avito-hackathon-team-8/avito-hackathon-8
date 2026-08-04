# Avito Hackathon 8

## Стек

- Frontend: React, TypeScript, Vite
- Backend: Go
- База данных: PostgreSQL + GORM
- Авторизация: одноразовый код из email + JWT
- Запуск: Docker Compose

## Запуск

Скопируйте пример настроек, при необходимости измените значения и запустите
контейнеры:

```sh
cp .env.example .env
make up
```

- Приложение: `http://localhost:3000`
- Backend health check: `http://localhost:8090/api/health`

GORM автоматически создаёт таблицы `users` и `otps` при старте backend.

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

## Авторизация

Пользователь вводит email, получает 8-значный одноразовый код и подтверждает
его. Если email встречается впервые, пользователь создаётся автоматически.
Коды хранятся только как HMAC-хеши и действуют 5 минут. Для каждого кода
разрешено до трёх неверных вводов; после третьей ошибки код удаляется. Успешно
использованный или истёкший код также удаляется. API:

- `POST /api/app/auth/request-otp` — отправить код;
- `POST /api/app/auth/verify-otp` — проверить код и получить JWT;
- `GET /api/app/auth/me` — получить текущего пользователя по Bearer JWT.

## Структура backend

```text
backend/
├── internal/
│   ├── auth/       # OTP, JWT и сценарии авторизации
│   ├── database/   # PostgreSQL/GORM
│   ├── email/      # SMTP и локальный log sender
│   ├── handlers/   # HTTP routes и JSON handlers
│   └── models/     # GORM-модели User и OTP
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
