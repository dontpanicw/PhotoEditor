# Image Processor

Микросервис для обработки изображений с использованием Kafka, MinIO и PostgreSQL.

## Возможности

- Загрузка изображений через REST API
- Асинхронная обработка изображений через Kafka
- Изменение размера изображений (Resize)
- Создание миниатюр (Thumbnail)
- Добавление водяных знаков (Watermark)
- Хранение изображений в MinIO
- Метаданные в PostgreSQL
- Web-интерфейс для управления

## Архитектура

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│   Client    │─────▶│   App API   │─────▶│   Kafka     │
└─────────────┘      └─────────────┘      └─────────────┘
                            │                     │
                            ▼                     ▼
                     ┌─────────────┐      ┌─────────────┐
                     │  PostgreSQL │      │   Worker    │
                     └─────────────┘      └─────────────┘
                            ▲                     │
                            │                     ▼
                            │              ┌─────────────┐
                            └──────────────│   MinIO     │
                                          └─────────────┘
```

## Технологии

- **Go 1.21+** - основной язык
- **Kafka** - очередь сообщений
- **PostgreSQL** - база данных
- **MinIO** - объектное хранилище
- **Docker** - контейнеризация
- **libvips** - обработка изображений

## Быстрый старт

### Требования

- Docker и Docker Compose
- Go 1.21+ (для локальной разработки)

### Запуск

```bash
# Клонировать репозиторий
git clone <repository-url>
cd ImageProcessor

# Запустить все сервисы
docker-compose up --build

# Или через Makefile
make docker-up-build
```

Приложение будет доступно по адресу: http://localhost:8080

### API Endpoints

- `POST /upload` - загрузка изображения
- `GET /image/{id}` - получение обработанного изображения
- `GET /image/{id}/status` - проверка статуса обработки
- `DELETE /image/{id}` - удаление изображения

### Пример использования

```bash
# Загрузка изображения с действиями
curl -X POST http://localhost:8080/upload \
  -F "image=@photo.jpg" \
  -F "actions=Resize,Watermark"

# Проверка статуса
curl http://localhost:8080/image/{id}/status

# Получение обработанного изображения
curl http://localhost:8080/image/{id} -o processed.jpg
```

## Разработка

### Установка зависимостей

```bash
make deps
```

### Запуск тестов

```bash
# Все тесты
make test

# С покрытием
make test-coverage

# Только unit-тесты
make test-unit
```

### Линтинг и форматирование

```bash
# Форматирование кода
make fmt

# Линтинг
make lint

# Все проверки
make check
```

### Локальный запуск

```bash
# Запустить приложение
make run

# Запустить worker
make run-worker
```

## Структура проекта

```
.
├── cmd/                    # Точки входа приложения
├── config/                 # Конфигурация
├── internal/
│   ├── adapter/           # Адаптеры (Kafka, MinIO, PostgreSQL)
│   ├── app/               # Инициализация приложения
│   ├── domain/            # Доменные модели
│   ├── input/             # HTTP handlers
│   ├── port/              # Интерфейсы
│   └── usecases/          # Бизнес-логика
├── image_worker/          # Worker для обработки изображений
├── pkg/                   # Общие пакеты
├── web/                   # Frontend
├── docker-compose.yml     # Docker конфигурация
└── Makefile              # Команды для разработки
```

## Переменные окружения

```env
# HTTP
HTTP_PORT=8080

# PostgreSQL
MASTER_DSN=postgres://user:pass@localhost:5432/db?sslmode=disable
SLAVE_DSN=postgres://user:pass@localhost:5432/db?sslmode=disable

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
BUCKET_NAME=images

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TASK_TOPIC=image-tasks
```

## Тестирование

Проект покрыт unit-тестами:

- Handlers (HTTP)
- Use cases (бизнес-логика)
- Image processor (обработка изображений)
- Domain models

Запуск тестов:

```bash
make test
```

Покрытие кода:

```bash
make test-coverage
open coverage.html
```

## CI/CD

Проект использует GitHub Actions для автоматического тестирования и линтинга при каждом push и pull request.

## Лицензия

MIT
