#!/bin/bash

# Тесты для Auth эндпоинтов

source "$(dirname "$0")/common.sh"

echo "=== Auth эндпоинты ==="

# Регистрация
test_endpoint "POST" "/auth/register" \
    '{"email":"auth_test@example.com","username":"authtest","password":"password123"}' \
    "Регистрация пользователя"

# Вход
test_endpoint "POST" "/auth/login" \
    '{"email":"auth_test@example.com","password":"password123"}' \
    "Вход пользователя"

# Вход с неверными данными
test_endpoint "POST" "/auth/login" \
    '{"email":"nonexistent@example.com","password":"wrongpassword"}' \
    "Вход с неверными данными (ожидается ошибка)"

print_stats

