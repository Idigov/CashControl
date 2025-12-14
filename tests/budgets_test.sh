#!/bin/bash

# Тесты для Budget эндпоинтов

source "$(dirname "$0")/common.sh"

echo "=== Budget эндпоинты ==="

# Создаем тестового пользователя
USER_ID=$(create_test_user "budget_test_$(date +%s)@example.com" "budgettest")
if [ -z "$USER_ID" ]; then
    echo "  ⚠ Не удалось создать пользователя, используем ID=1"
    USER_ID=1
fi

# Получаем текущий месяц и год
current_month=$(date +%m | sed 's/^0//')
current_year=$(date +%Y)

# Список бюджетов
test_endpoint "GET" "/budgets?user_id=$USER_ID" "" "Список бюджетов"

# Создание бюджета
test_endpoint "POST" "/budgets?user_id=$USER_ID" \
    "{\"amount\":5000,\"month\":$current_month,\"year\":$current_year}" \
    "Создание бюджета" "BUDGET_ID"

# Статус бюджета
test_endpoint "GET" "/budgets/status?user_id=$USER_ID&month=$current_month&year=$current_year" "" "Статус бюджета"

# Бюджет по месяцу
test_endpoint "GET" "/budgets/by-month?user_id=$USER_ID&month=$current_month&year=$current_year" "" "Бюджет по месяцу"

# Получение бюджета по ID
if [ -n "$BUDGET_ID" ]; then
    test_endpoint "GET" "/budgets/$BUDGET_ID" "" "Получение бюджета по ID"
    
    # Обновление бюджета
    test_endpoint "PATCH" "/budgets/$BUDGET_ID" \
        '{"amount":6000}' \
        "Обновление бюджета"
    
    # Удаление бюджета
    test_endpoint "DELETE" "/budgets/$BUDGET_ID" "" "Удаление бюджета"
fi

print_stats

