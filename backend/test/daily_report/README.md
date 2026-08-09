# Daily report API scenarios

Black-box e2e-сценарии для HTTP и WebSocket API ежедневной сводки. Тесты
обращаются к запущенному Compose backend и создают изолированных пользователей
в Postgres.

Проверяются:

- авторизация и пустая сводка нового пользователя;
- появление выполненных заданий, заработанных листьев, посещения, повышения
  уровня и полученной награды в `GET /api/v1/daily-report`;
- первоначальный снимок и последующие полные обновления через
  `GET /api/v1/daily-report/ws`.

## Запуск

Из корня проекта:

```sh
make up
cd backend
RUN_BACKEND_E2E=1 GOCACHE="${TMPDIR:-/tmp}/avito-go-build-cache" go test ./test/daily_report -count=1 -v
```
