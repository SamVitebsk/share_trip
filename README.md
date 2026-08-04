# Share Trip

## О проекте

Share Trip - backend-сервис на Go для проекта совместных поездок. Приложение запускает HTTP API на Fiber, подключается к PostgreSQL через `pgx` и содержит endpoint готовности сервиса:

```text
GET /api/ready
```

Endpoint проверяет соединение с базой данных и возвращает успешный HTTP-статус, если приложение готово принимать запросы.

## Технологии

- Go
- Fiber
- PostgreSQL
- pgx
- goose
- Docker Compose
- golangci-lint

## Сборка и запуск

Скачать зависимости:

```bash
make deps
```

Запустить PostgreSQL:

```bash
make up
```

Применить миграции:

```bash
make migrate-up
```

Запустить приложение:

```bash
make run
```

По умолчанию приложение доступно на порту `9090`.

## Использование

Проверить готовность сервиса:

```bash
curl http://localhost:9090/api/ready
```

Ожидаемый ответ:

```text
OK
```

## Проверки

Форматирование кода:

```bash
make fmt
```

Запуск линтера:

```bash
make lint
```

Запуск тестов:

```bash
make test
```

Сборка бинарного файла:

```bash
make build
```

Запуск всех основных проверок:

```bash
make check
```

## Настройки окружения

Приложение использует следующие переменные окружения:

| Переменная | Значение по умолчанию | Описание |
| --- | --- | --- |
| `SERVER_PORT` | `:9090` | Порт HTTP-сервера |
| `DB_HOST` | `localhost` | Хост PostgreSQL |
| `DB_PORT` | `6543` | Порт PostgreSQL |
| `DB_USER` | `postgres` | Пользователь базы данных |
| `DB_PASSWORD` | `admin` | Пароль базы данных |
| `DB_NAME` | `share_trip` | Имя базы данных |
| `DB_SSLMODE` | `disable` | Режим SSL для подключения к PostgreSQL |

