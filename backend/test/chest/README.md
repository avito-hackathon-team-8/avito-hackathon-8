# Chest API scenarios

Black-box e2e-сценарий для открытия сундука. 

Проверяются стартовый баланс демонстрационного питомца (`1000` листьев), пять
последовательных открытий сундука, немедленное появление каждой награды в
`GET /api/app/rewards`, списание 200 листьев и отказ при нулевом балансе.

## Запуск

```sh
make up
cd backend
GOCACHE="${TMPDIR:-/tmp}/avito-go-build-cache" go test ./test/chest -count=1 -v
```
