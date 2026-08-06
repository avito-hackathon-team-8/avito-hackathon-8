# Pet API scenarios

Black-box e2e-сценарии для API и WebSocket питомца. Тесты используют
запущенный Compose backend и создают изолированных пользователей в Postgres.

Проверяются:

- создание питомца при первом `GET /api/v1/pet`;
- авторизация и порядок первоначальной настройки имени;
- валидация и сохранение имени;
- initial snapshot WebSocket;
- получение листьев после `record -> claim` задачи и WebSocket-обновление;
- отклонение неавторизованного WebSocket-подключения.

## Запуск

```sh
make up
cd backend
GOCACHE="${TMPDIR:-/tmp}/avito-go-build-cache" go test ./test/pet -count=1 -v
```
