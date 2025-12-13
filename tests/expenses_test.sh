#!/bin/bash

# Тесты для Expense эндпоинтов

source "$(dirname "$0")/common.sh"

echo "=== Expense эндпоинты ==="

# Создаем тестового пользователя
USER_ID=$(create_test_user "expense_test_$(date +%s)@example.com" "expensetest")
if [ -z "$USER_ID" ]; then
    echo "  ⚠ Не удалось создать пользователя, используем ID=1"
    USER_ID=1
fi

# Создаем категорию
category_response=$(curl -s -w "\n%{http_code}" -X "POST" "$BASE_URL/categories/$USER_ID" \
    -H "Content-Type: application/json" \
    -d '{"name":"Категория для расходов","description":"Тестовая категория"}')
category_code=$(echo "$category_response" | tail -n1)
category_body=$(echo "$category_response" | sed '$d')
if [ "$category_code" -ge 200 ] && [ "$category_code" -lt 300 ]; then
    CATEGORY_ID=$(extract_id "$category_body")
    echo "  ✓ Категория создана (ID: $CATEGORY_ID)"
else
    echo "  ⚠ Не удалось создать категорию, используем ID=1"
    CATEGORY_ID=1
fi

# Список расходов
test_endpoint "GET" "/expenses?user_id=$USER_ID" "" "Список расходов"

# Создание расхода
if [ -n "$CATEGORY_ID" ]; then
    current_date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    test_endpoint "POST" "/expenses?user_id=$USER_ID" \
        "{\"category_id\":$CATEGORY_ID,\"amount\":1000.50,\"description\":\"Тестовый расход\",\"date\":\"$current_date\"}" \
        "Создание расхода" "EXPENSE_ID"
    
    # Получение расхода по ID
    if [ -n "$EXPENSE_ID" ]; then
        test_endpoint "GET" "/expenses/$EXPENSE_ID" "" "Получение расхода по ID"
        
        # Обновление расхода
        test_endpoint "PATCH" "/expenses/$EXPENSE_ID" \
            '{"amount":1500.75,"description":"Обновленный расход"}' \
            "Обновление расхода"
        
        # Удаление расхода
        test_endpoint "DELETE" "/expenses/$EXPENSE_ID" "" "Удаление расхода"
    fi
fi

print_stats

