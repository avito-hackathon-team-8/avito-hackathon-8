# Weekly login API scenarios

Black-box integration-тесты для `/api/v1/weekly-login`. Тесты обращаются к
запущенному api-service по HTTP и создают изолированных пользователей в Compose
Postgres через `docker compose exec postgres psql`.

Основной сценарий выполняется в порядке:

1. `POST /api/v1/weekly-login/activity`;
2. `GET /api/v1/weekly-login` и проверка дня со статусом `AVAILABLE`;
3. `POST /api/v1/weekly-login/claim`;
4. повторный `GET` и проверка статуса `CLAIMED`.

## Запуск

Из корня проекта:

```sh
make up
cd api-service
RUN_API_SERVICE_E2E=1 GOCACHE="${TMPDIR:-/tmp}/avito-go-build-cache" go test ./test/weekly_login -count=1 -v
```
