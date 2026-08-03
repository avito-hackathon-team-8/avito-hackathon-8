# Avito Hackathon 8

Приложение для хакатона Avito.

## Стек

- Frontend: React, TypeScript, Vite
- Backend: Go, PocketBase
- База данных: SQLite
- Запуск: Docker Compose

SQLite выбрана как встроенная база PocketBase: она не требует отдельного сервиса, хранит данные в одном volume и подходит для быстрого прототипирования.

## Запуск

```sh
make up
```

- Приложение: `http://localhost:3000`
- PocketBase: `http://localhost:8090/_/`

## Команды

```sh
make help
make logs
make lint
make feature NAME=auth
make down
```

## Администратор

```sh
make admin EMAIL=email@example.com PASS=password
```
