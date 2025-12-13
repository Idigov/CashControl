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
│   │   ├── user.go                    # Модель пользователя
│   │   ├── category.go                # Модель категории
│   │   ├── expense.go                 # Модель расхода
│   │   ├── budget.go                  # Модель бюджета
│   │   ├── recurring_expense.go      # Модель регулярного расхода
│   │   ├── activity_history.go        # Модель истории действий
│   │   └── statistics.go              # Модели статистики
│   ├── repository/
│   │   ├── user_repository.go         # Репозиторий пользователей
│   │   ├── category_repository.go     # Репозиторий категорий
│   │   ├── expense_repository.go      # Репозиторий расходов
│   │   ├── budget_repository.go       # Репозиторий бюджета
│   │   ├── recurring_expense_repository.go # Репозиторий регулярных расходов
│   │   └── activity_log_repository.go # Репозиторий истории действий
│   └── services/
│       ├── auth_service.go            # Сервис аутентификации
│       ├── user_service.go            # Сервис пользователей
│       ├── category_service.go        # Сервис категорий
│       ├── expense_service.go         # Сервис расходов
│       ├── budget_service.go          # Сервис бюджета
│       ├── recurring_expense_service.go # Сервис регулярных расходов
│       └── activity_log_service.go    # Сервис истории действий
├── .env                               # Переменные окружения
├── .env.example                       # Пример переменных окружения
├── .air.toml                          # Конфигурация Air для hot reload
├── .dockerignore                      # Исключения для Docker
├── Dockerfile                         # Конфигурация Docker образа
├── docker-compose.yml                # Конфигурация Docker Compose
├── .gitignore                         # Игнорируемые файлы
├── go.mod                             # Зависимости Go
├── go.sum                             # Checksums зависимостей
├── Makefile                           # Make команды
├── README.md                          # Документация
├── TASKS.md                           # Список задач проекта
├── TEST_ENDPOINTS.md                  # Документация по тестированию
├── POSTMAN_EXPENSES.md                # Документация Postman для Expenses
└── test_endpoints.sh                  # Скрипт для тестирования эндпоинтов
```

## Запуск

### Вариант 1: Запуск через Docker (рекомендуется)

Этот способ не требует локальной установки PostgreSQL - используется облачная база данных Supabase.

#### Настройка Supabase

1. Создайте аккаунт на [supabase.com](https://supabase.com) (бесплатно)
2. Создайте новый проект
3. Перейдите в **Settings → Database**
4. Скопируйте **Connection string** (URI mode)
5. Формат строки: `postgresql://postgres:[PASSWORD]@[HOST]:5432/postgres?sslmode=require`

#### Настройка переменных окружения

Создайте файл `.env` в корне проекта:

```bash
# Адрес сервера
SERVER_ADDRESS=:8080
ENVIRONMENT=production

# Строка подключения к Supabase (замените на вашу)
DATABASE_URL=postgresql://postgres:your-password@db.xxxxx.supabase.co:5432/postgres?sslmode=require

# JWT секрет (измените на случайную строку!)
JWT_SECRET=your-secret-key-change-in-production
```

#### Запуск через Docker Compose

```bash
# Сборка и запуск контейнера
docker-compose up --build

# Запуск в фоновом режиме
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Остановка
docker-compose down
```

Приложение будет доступно по адресу: `http://localhost:8080`

#### Прямая сборка Docker образа

```bash
# Сборка образа
docker build -t cashcontrol .

# Запуск контейнера
docker run -p 8080:8080 --env-file .env cashcontrol
```

### Вариант 2: Локальный запуск (требуется локальный PostgreSQL)

1. Установите зависимости:
```bash
go mod download
```

2. Настройте `.env` файл с параметрами подключения к локальному PostgreSQL

3. Запустите сервер:
```bash
go run cmd/cashcontrol/main.go
```

Или с использованием Air для hot reload:
```bash
air
```

## API Эндпоинты

### Auth
- `POST /auth/register` - Регистрация пользователя
- `POST /auth/login` - Вход пользователя

### Users
- `GET /users` - Список пользователей
- `POST /users` - Создание пользователя
- `GET /users/:id` - Получение пользователя
- `PATCH /users/:id` - Обновление пользователя
- `DELETE /users/:id` - Удаление пользователя

### Categories
- `GET /categories/:userId` - Список категорий пользователя
- `POST /categories/:userId` - Создание категории
- `GET /categories/detail/:id` - Получение категории
- `PATCH /categories/:id` - Обновление категории
- `DELETE /categories/:id` - Удаление категории

### Expenses
- `GET /expenses?user_id=X` - Список расходов (с фильтрацией)
- `POST /expenses?user_id=X` - Создание расхода
- `GET /expenses/:id` - Получение расхода
- `PATCH /expenses/:id` - Обновление расхода
- `DELETE /expenses/:id` - Удаление расхода

### Budgets
- `GET /budgets?user_id=X` - Список бюджетов пользователя
- `POST /budgets?user_id=X` - Создание бюджета
- `GET /budgets/status?user_id=X&month=Y&year=Z` - Статус бюджета
- `GET /budgets/by-month?user_id=X&month=Y&year=Z` - Бюджет по месяцу
- `GET /budgets/:id` - Получение бюджета
- `PATCH /budgets/:id` - Обновление бюджета
- `DELETE /budgets/:id` - Удаление бюджета

### Recurring Expenses
- `GET /recurring-expenses?user_id=X` - Список регулярных расходов
- `POST /recurring-expenses?user_id=X` - Создание регулярного расхода
- `GET /recurring-expenses/active?user_id=X` - Активные регулярные расходы
- `GET /recurring-expenses/:id` - Получение регулярного расхода
- `PATCH /recurring-expenses/:id` - Обновление регулярного расхода
- `DELETE /recurring-expenses/:id` - Удаление регулярного расхода
- `POST /recurring-expenses/:id/activate` - Активация
- `POST /recurring-expenses/:id/deactivate` - Деактивация

## Технологии

- **Go** - Язык программирования
- **Gin** - HTTP веб-фреймворк
- **GORM** - ORM для работы с БД
- **PostgreSQL** - База данных (Supabase для облачного варианта)
- **Docker** - Контейнеризация приложения
- **Air** - Hot reload для разработки

## Деплой

Приложение готово к деплою на любую платформу, поддерживающую Docker:
- **Railway** - автоматический деплой из Git репозитория
- **Render** - простой деплой с поддержкой Docker
- **Fly.io** - быстрый деплой с глобальной сетью
- **DigitalOcean App Platform** - масштабируемый деплой

Для деплоя:
1. Подключите ваш Git репозиторий
2. Укажите Dockerfile как источник сборки
3. Настройте переменные окружения (DATABASE_URL, JWT_SECRET)
4. Деплой произойдет автоматически
