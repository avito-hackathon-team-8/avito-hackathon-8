# Chest API scenarios

Black-box e2e-сценарий для открытия сундука. 

Проверяются открытие сундука, немедленное появление награды в `GET /api/app/rewards`,
списание 200 листьев и отказ при недостаточном балансе.

## Запуск

```sh
make up
cd backend
GOCACHE="${TMPDIR:-/tmp}/avito-go-build-cache" go test ./test/chest -count=1 -v
```
