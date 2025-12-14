.PHONY: run build test fmt vet lint tidy clean dev seed docker-up docker-down docker-stop docker-restart docker-logs test-endpoints test-auth test-users test-categories test-expenses test-budgets test-recurring test-statistics

GO           ?= go
BINARY       ?= cashcontrol
CMD_MAIN     := ./cmd/cashcontrol/main.go

run: ## Запуск основного приложения (HTTP-сервер)
	$(GO) run $(CMD_MAIN)

dev: ## Запуск в режиме разработки с hot reload (air)
	air -c .air.toml

build: ## Сборка бинарника приложения
	$(GO) build -o tmp/$(BINARY) $(CMD_MAIN)

test: ## Запуск всех тестов
	$(GO) test ./...

fmt: ## Форматирование кода
	$(GO) fmt ./...

vet: ## Статический анализ кода
	$(GO) vet ./...

lint: ## Линтинг кода с помощью golangci-lint
	golangci-lint run

tidy: ## Обновление зависимостей (go.mod / go.sum)
	$(GO) mod tidy

clean: ## Удаление собранных бинарников
	rm -rf tmp

docker-up: ## Запуск приложения в Docker (сборка и запуск одной командой)
	docker-compose up -d --build

docker-down: ## Остановка и удаление Docker контейнеров
	docker-compose down

docker-stop: ## Остановка Docker контейнеров (без удаления)
	docker-compose stop

docker-restart: ## Перезапуск Docker контейнеров
	docker-compose restart

docker-logs: ## Просмотр логов Docker контейнеров
	docker-compose logs -f

test-endpoints: ## Тестирование эндпоинтов API (все тесты)
	./tests/run_all.sh

test-auth: ## Тестирование Auth эндпоинтов
	./tests/auth_test.sh

test-users: ## Тестирование User эндпоинтов
	./tests/users_test.sh

test-categories: ## Тестирование Category эндпоинтов
	./tests/categories_test.sh

test-expenses: ## Тестирование Expense эндпоинтов
	./tests/expenses_test.sh

test-budgets: ## Тестирование Budget эндпоинтов
	./tests/budgets_test.sh

test-recurring: ## Тестирование Recurring Expense эндпоинтов
	./tests/recurring_expenses_test.sh

test-statistics: ## Тестирование Statistics эндпоинтов
	./tests/statistics_test.sh
