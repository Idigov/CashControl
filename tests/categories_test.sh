#!/bin/bash

# Тесты для Category эндпоинтов

source "$(dirname "$0")/common.sh"

echo "=== Category эндпоинты ==="

# Создаем тестового пользователя
USER_ID=$(create_test_user "category_test_$(date +%s)@example.com" "categorytest")
if [ -z "$USER_ID" ]; then
    echo "  ⚠ Не удалось создать пользователя, используем ID=1"
    USER_ID=1
fi

# Список категорий пользователя
test_endpoint "GET" "/categories/$USER_ID" "" "Список категорий пользователя"

# Создание категории
test_endpoint "POST" "/categories/$USER_ID" \
    '{"name":"Тестовая категория","description":"Описание категории","color":"#FF5733"}' \
    "Создание категории" "CATEGORY_ID"

# Получение категории по ID
if [ -n "$CATEGORY_ID" ]; then
    test_endpoint "GET" "/categories/detail/$CATEGORY_ID" "" "Получение категории по ID"
    
    # Обновление категории
    test_endpoint "PATCH" "/categories/$CATEGORY_ID" \
        '{"name":"Обновленная категория","color":"#33FF57"}' \
        "Обновление категории"
    
    # Удаление категории
    test_endpoint "DELETE" "/categories/$CATEGORY_ID" "" "Удаление категории"
fi

print_stats

