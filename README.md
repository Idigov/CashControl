# CashControl

Серверное приложение для управления расходами на Go.

## Структура проекта

```
CashControl/
├── cmd/
│   └── cashcontrol/
│       └── main.go              # Точка входа приложения
├── internal/
│   ├── config/
│   │   └── config.go            # Конфигурация приложения
│   ├── database/
│   │   └── database.go          # Подключение к БД и миграции
│   ├── handlers/
│   │   └── handlers.go           # HTTP обработчики
│   ├── models/
│   │   ├── user.go               # Модель пользователя
│   │   ├── category.go           # Модель категории
│   │   ├── expense.go           # Модель расхода
│   │   ├── budget.go            # Модель бюджета
│   │   ├── recurring_expense.go # Модель регулярного расхода
│   │   ├── activity_history.go   # Модель истории действий
│   │   └── statistics.go        # Модели статистики
│   └── repository/
│       └── repository.go         # Репозитории
├── .env                         # Переменные окружения
├── .air.toml                    # Конфигурация Air
├── go.mod                       # Зависимости Go
└── README.md                    # Документация
```

## Запуск

1. Установите зависимости:
```bash
go mod download
```

2. Настройте `.env` файл с параметрами подключения к PostgreSQL

3. Запустите сервер:
```bash
go run cmd/cashcontrol/main.go
```

Или с использованием Air для hot reload:
```bash
air
```

