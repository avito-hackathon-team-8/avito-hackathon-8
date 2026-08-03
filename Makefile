SHELL := /bin/sh

.DEFAULT_GOAL := help

.PHONY: help up down restart build logs ps lint lint-frontend lint-backend admin branch feature bugfix chore

help:
	@printf '%s\n' \
		'make up                   Build and start the application' \
		'make down                 Stop the application' \
		'make restart              Restart the application' \
		'make build                Build all Docker images' \
		'make logs                 Follow service logs' \
		'make ps                   Show service status' \
		'make lint                 Run all checks' \
		'make admin EMAIL=x PASS=y Create PocketBase admin' \
		'make feature NAME=login   Create feature/login' \
		'make bugfix NAME=api      Create bugfix/api' \
		'make chore NAME=deps      Create chore/deps' \
		'make branch TYPE=x NAME=y Create x/y'

up:
	docker compose up --build -d

down:
	docker compose down

restart:
	docker compose restart

build:
	docker compose build

logs:
	docker compose logs -f

ps:
	docker compose ps

lint: lint-frontend lint-backend

lint-frontend:
	docker run --rm -v "$(CURDIR)/frontend:/app" -v /app/node_modules -w /app node:24-alpine sh -c "npm ci && npm run lint && npm run lint:styles"

lint-backend:
	docker run --rm -v "$(CURDIR):/app:ro" -w /app/backend golangci/golangci-lint:v2.12.2-alpine golangci-lint run --config ../.golangci.yaml

admin:
	@test -n "$(EMAIL)" || (printf '%s\n' 'EMAIL is required' && exit 1)
	@test -n "$(PASS)" || (printf '%s\n' 'PASS is required' && exit 1)
	docker compose exec backend ./backend superuser create "$(EMAIL)" "$(PASS)"

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
