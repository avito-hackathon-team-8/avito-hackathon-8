SHELL := /bin/sh

.DEFAULT_GOAL := help

.PHONY: help up down restart build logs ps migrate test-api-service-e2e lint lint-frontend lint-api-service lint-daily-tasks-service lint-pet-state branch feature bugfix chore

help:
	@printf '%s\n' \
		'make up                   Build and start the application' \
		'make down                 Stop the application' \
		'make restart              Restart the application' \
		'make build                Build all Docker images' \
		'make logs                 Follow service logs' \
		'make ps                   Show service status' \
		'make migrate              Apply database migrations' \
		'make test-api-service-e2e     Run api-service end-to-end tests' \
		'make lint                 Run all checks' \
		'make feature NAME=login   Create feature/login' \
		'make bugfix NAME=api      Create bugfix/api' \
		'make chore NAME=deps      Create chore/deps' \
		'make branch TYPE=x NAME=y Create x/y'

up:
	docker compose run --rm migrator
	docker compose up --build -d

down:
	docker compose down

restart:
	docker compose restart

build:
	docker compose build

logs:
	docker compose logs --since=1m -f

ps:
	docker compose ps

migrate:
	docker compose run --rm migrator

test-api-service-e2e:
	cd api-service && RUN_API_SERVICE_E2E=1 go test ./test/...

lint: lint-frontend lint-api-service lint-daily-tasks-service
	$(MAKE) lint-pet-state

lint-frontend:
	docker run --rm -v "$(CURDIR)/frontend:/app" -v /app/node_modules -w /app node:24-alpine sh -c "npm ci && npm run lint && npm run lint:styles"

lint-api-service:
	docker run --rm -v "$(CURDIR):/app:ro" -w /app/api-service golangci/golangci-lint:v2.12.2-alpine golangci-lint run --config ../.golangci.yaml

lint-daily-tasks-service:
	docker run --rm -v "$(CURDIR):/app:ro" -w /app/daily-tasks-service golangci/golangci-lint:v2.12.2-alpine golangci-lint run --config ../.golangci.yaml

lint-pet-state:
	docker run --rm -v "$(CURDIR):/app:ro" -w /app/pet-state-service golangci/golangci-lint:v2.12.2-alpine golangci-lint run --config ../.golangci.yaml

branch:
	@test -n "$(TYPE)" || (printf '%s\n' 'TYPE is required' && exit 1)
	@test -n "$(NAME)" || (printf '%s\n' 'NAME is required' && exit 1)
	@printf '%s' "$(NAME)" | grep -Eq '^[a-z0-9][a-z0-9._-]*$$' || (printf '%s\n' 'NAME must use lowercase letters, numbers, dots, dashes, or underscores' && exit 1)
	git switch -c "$(TYPE)/$(NAME)"

feature:
	@$(MAKE) branch TYPE=feature NAME="$(NAME)"

bugfix:
	@$(MAKE) branch TYPE=bugfix NAME="$(NAME)"

chore:
	@$(MAKE) branch TYPE=chore NAME="$(NAME)"
