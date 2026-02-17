.PHONY: test test-coverage test-unit test-integration build run docker-up docker-down clean lint

# Переменные
BINARY_NAME=imageprocessor
WORKER_BINARY_NAME=worker
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean

# Цвета для вывода
GREEN=\033[0;32m
NC=\033[0m # No Color

# Тесты
test:
	@echo "$(GREEN)Running all tests...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage: test
	@echo "$(GREEN)Generating coverage report...$(NC)"
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

test-unit:
	@echo "$(GREEN)Running unit tests...$(NC)"
	$(GOTEST) -v -short ./...

test-integration:
	@echo "$(GREEN)Running integration tests...$(NC)"
	$(GOTEST) -v -run Integration ./...

# Сборка
build:
	@echo "$(GREEN)Building application...$(NC)"
	$(GOBUILD) -o bin/$(BINARY_NAME) cmd/main.go
	$(GOBUILD) -o bin/$(WORKER_BINARY_NAME) image_worker/internal/cmd/main.go
	@echo "$(GREEN)Build complete!$(NC)"

build-linux:
	@echo "$(GREEN)Building for Linux...$(NC)"
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-linux cmd/main.go
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(WORKER_BINARY_NAME)-linux image_worker/internal/cmd/main.go

# Запуск
run:
	@echo "$(GREEN)Running application...$(NC)"
	$(GO) run cmd/main.go

run-worker:
	@echo "$(GREEN)Running worker...$(NC)"
	$(GO) run image_worker/internal/cmd/main.go

# Docker
docker-up:
	@echo "$(GREEN)Starting Docker containers...$(NC)"
	docker-compose up -d

docker-up-build:
	@echo "$(GREEN)Building and starting Docker containers...$(NC)"
	docker-compose up --build -d

docker-down:
	@echo "$(GREEN)Stopping Docker containers...$(NC)"
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-clean:
	@echo "$(GREEN)Cleaning Docker containers and volumes...$(NC)"
	docker-compose down -v
	docker system prune -f

# Линтинг
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

# Форматирование
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO) fmt ./...

# Очистка
clean:
	@echo "$(GREEN)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

# Зависимости
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod tidy

# Проверка
check: fmt lint test
	@echo "$(GREEN)All checks passed!$(NC)"

# База данных
db-migrate:
	@echo "$(GREEN)Running migrations...$(NC)"
	$(GO) run cmd/main.go migrate

# Помощь
help:
	@echo "Available targets:"
	@echo "  test              - Run all tests"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  test-unit         - Run unit tests only"
	@echo "  build             - Build binaries"
	@echo "  run               - Run application"
	@echo "  run-worker        - Run worker"
	@echo "  docker-up         - Start Docker containers"
	@echo "  docker-up-build   - Build and start Docker containers"
	@echo "  docker-down       - Stop Docker containers"
	@echo "  docker-logs       - Show Docker logs"
	@echo "  docker-clean      - Clean Docker containers and volumes"
	@echo "  lint              - Run linter"
	@echo "  fmt               - Format code"
	@echo "  clean             - Clean build artifacts"
	@echo "  deps              - Download dependencies"
	@echo "  check             - Run fmt, lint, and test"
	@echo "  help              - Show this help"
