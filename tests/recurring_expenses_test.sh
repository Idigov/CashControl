#!/bin/bash

# Тесты для Recurring Expense эндпоинтов

source "$(dirname "$0")/common.sh"

echo "=== Recurring Expense эндпоинты ==="

# Создаем тестового пользователя
USER_ID=$(create_test_user "recurring_test_$(date +%s)@example.com" "recurringtest")
if [ -z "$USER_ID" ]; then
    echo "  ⚠ Не удалось создать пользователя, используем ID=1"
    USER_ID=1
fi

# Создаем категорию
category_response=$(curl -s -w "\n%{http_code}" -X "POST" "$BASE_URL/categories/$USER_ID" \
    -H "Content-Type: application/json" \
    -d '{"name":"Категория для регулярных расходов","description":"Тестовая категория"}')
category_code=$(echo "$category_response" | tail -n1)
category_body=$(echo "$category_response" | sed '$d')
if [ "$category_code" -ge 200 ] && [ "$category_code" -lt 300 ]; then
    CATEGORY_ID=$(extract_id "$category_body")
    echo "  ✓ Категория создана (ID: $CATEGORY_ID)"
else
    echo "  ⚠ Не удалось создать категорию, используем ID=1"
    CATEGORY_ID=1
fi

# Список регулярных расходов
test_endpoint "GET" "/recurring-expenses?user_id=$USER_ID" "" "Список регулярных расходов"

# Создание регулярного расхода
if [ -n "$CATEGORY_ID" ]; then
    # Получаем дату через месяц
    next_month_date=$(date -u -v+1m +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "+1 month" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    test_endpoint "POST" "/recurring-expenses?user_id=$USER_ID" \
        "{\"category_id\":$CATEGORY_ID,\"amount\":500,\"description\":\"Ежемесячная подписка\",\"type\":\"monthly\"}" \
        "Создание регулярного расхода" "RECURRING_EXPENSE_ID"
    
    # Активные регулярные расходы
    test_endpoint "GET" "/recurring-expenses/active?user_id=$USER_ID" "" "Активные регулярные расходы"
    
    # Получение регулярного расхода по ID
    if [ -n "$RECURRING_EXPENSE_ID" ]; then
        test_endpoint "GET" "/recurring-expenses/$RECURRING_EXPENSE_ID" "" "Получение регулярного расхода по ID"
        
        # Обновление регулярного расхода
        test_endpoint "PATCH" "/recurring-expenses/$RECURRING_EXPENSE_ID" \
            '{"amount":600,"description":"Обновленная подписка"}' \
            "Обновление регулярного расхода"
        
        # Активация
        test_endpoint "POST" "/recurring-expenses/$RECURRING_EXPENSE_ID/activate" "" "Активация регулярного расхода"
        
        # Деактивация
        test_endpoint "POST" "/recurring-expenses/$RECURRING_EXPENSE_ID/deactivate" "" "Деактивация регулярного расхода"
        
        # Удаление
        test_endpoint "DELETE" "/recurring-expenses/$RECURRING_EXPENSE_ID" "" "Удаление регулярного расхода"
    fi
fi

print_stats

