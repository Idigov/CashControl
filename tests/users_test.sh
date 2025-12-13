#!/bin/bash

# Тесты для User эндпоинтов

source "$(dirname "$0")/common.sh"

echo "=== User эндпоинты ==="

# Создаем тестового пользователя
USER_ID=$(create_test_user "user_test_$(date +%s)@example.com" "usertest")
if [ -z "$USER_ID" ]; then
    echo "  ⚠ Не удалось создать пользователя, используем ID=1"
    USER_ID=1
else
    echo "  ✓ Тестовый пользователь создан (ID: $USER_ID)"
fi

# Список пользователей
test_endpoint "GET" "/users" "" "Список пользователей"

# Создание пользователя
test_endpoint "POST" "/users" \
    "{\"email\":\"newuser_$(date +%s)@example.com\",\"username\":\"newuser\",\"password\":\"password123\"}" \
    "Создание пользователя" "NEW_USER_ID"

# Получение пользователя по ID
if [ -n "$USER_ID" ]; then
    test_endpoint "GET" "/users/$USER_ID" "" "Получение пользователя по ID"
    
    # Обновление пользователя
    test_endpoint "PATCH" "/users/$USER_ID" \
        "{\"username\":\"updateduser_$(date +%s)\"}" \
        "Обновление пользователя"
fi

print_stats

