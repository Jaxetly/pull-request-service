# Переменные
APP_NAME=pr-service
DOCKER_COMPOSE=docker-compose

.PHONY: all build run clean lint docker-up docker-down docker-logs

all: build

# Сборка локально
build:
	go build -o bin/$(APP_NAME) cmd/main.go

# Запуск локально
run: build
	./bin/$(APP_NAME)

# Запуск линтера
lint:
	golangci-lint run

# Поднятие в докере (с пересборкой)
docker-up:
	$(DOCKER_COMPOSE) up --build -d

# Остановка докера
docker-down:
	$(DOCKER_COMPOSE) down

# Просмотр логов докера
docker-logs:
	$(DOCKER_COMPOSE) logs

# Очистка бинарников
clean:
	rm -rf bin/