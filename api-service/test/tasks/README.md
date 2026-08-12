# Tasks API scenarios

Black-box сценарии для `/api/v1/tasks`, написанные на Go. Тесты ходят в
запущенный api-service по HTTP и создают изолированных пользователей в Compose
Postgres через `docker compose exec postgres psql`.

## Запуск

Из корня проекта:

```sh
make up
cd api-service
RUN_API_SERVICE_E2E=1 GOCACHE="${TMPDIR:-/tmp}/avito-go-build-cache" go test ./test/tasks -count=1 -v
```


