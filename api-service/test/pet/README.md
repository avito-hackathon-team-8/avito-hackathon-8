# Pet API scenarios

Black-box e2e-сценарии для API и WebSocket питомца. Тесты используют
запущенный Compose api-service и создают изолированных пользователей в Postgres.

Проверяются:

- стартовое состояние демонстрационного питомца: уровень 10 и 1000 листьев;
- fallback-создание питомца при первом `GET /api/v1/pet`;
- авторизация и порядок первоначальной настройки имени;
- валидация и сохранение имени;
- initial snapshot WebSocket;
- получение листьев после `record -> claim` задачи и WebSocket-обновление;
- отклонение неавторизованного WebSocket-подключения.

## Запуск

```sh
make up
cd api-service
RUN_API_SERVICE_E2E=1 GOCACHE="${TMPDIR:-/tmp}/avito-go-build-cache" go test ./test/pet -count=1 -v
```
