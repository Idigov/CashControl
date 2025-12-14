#!/bin/bash

# Запуск всех тестов

BASE_URL="${1:-http://localhost:8080}"
export BASE_URL

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Общие счетчики
TOTAL_ALL=0
SUCCESS_ALL=0
CLIENT_ERROR_ALL=0
SERVER_ERROR_ALL=0

echo "=========================================="
echo "Запуск всех тестов API: $BASE_URL"
echo "=========================================="
echo ""

# Запускаем каждый тест
for test_file in "$SCRIPT_DIR"/*_test.sh; do
    if [ -f "$test_file" ] && [ "$(basename "$test_file")" != "run_all.sh" ]; then
        echo ""
        # Запускаем тест и сохраняем вывод
        output=$(bash "$test_file" 2>&1)
        echo "$output"
        
        # Извлекаем статистику из вывода (убираем цветовые коды для парсинга)
        stats_output=$(echo "$output" | sed 's/\x1b\[[0-9;]*m//g')
        
        # Парсим статистику с учетом формата "  - Всего тестов: X"
        test_total=$(echo "$stats_output" | grep "Всего тестов:" | grep -oE '[0-9]+' | head -1)
        test_success=$(echo "$stats_output" | grep "Успешно:" | grep -oE '[0-9]+' | head -1)
        test_client=$(echo "$stats_output" | grep "Клиентские ошибки" | grep -oE '[0-9]+' | head -1)
        test_server=$(echo "$stats_output" | grep "Серверные ошибки" | grep -oE '[0-9]+' | head -1)
        
        # Добавляем к общим счетчикам (проверяем что значение не пустое и является числом)
        if [ -n "$test_total" ] && [ "$test_total" -eq "$test_total" ] 2>/dev/null; then
            TOTAL_ALL=$((TOTAL_ALL + test_total))
        fi
        if [ -n "$test_success" ] && [ "$test_success" -eq "$test_success" ] 2>/dev/null; then
            SUCCESS_ALL=$((SUCCESS_ALL + test_success))
        fi
        if [ -n "$test_client" ] && [ "$test_client" -eq "$test_client" ] 2>/dev/null; then
            CLIENT_ERROR_ALL=$((CLIENT_ERROR_ALL + test_client))
        fi
        if [ -n "$test_server" ] && [ "$test_server" -eq "$test_server" ] 2>/dev/null; then
            SERVER_ERROR_ALL=$((SERVER_ERROR_ALL + test_server))
        fi
    fi
done

echo ""
echo "=========================================="
echo "Итоговая статистика всех тестов:"
echo "=========================================="
echo -e "Всего тестов: ${TOTAL_ALL}"
echo -e "${GREEN}Успешно: ${SUCCESS_ALL}${NC}"
echo -e "${YELLOW}Клиентские ошибки (4xx): ${CLIENT_ERROR_ALL}${NC}"
echo -e "${RED}Серверные ошибки (5xx): ${SERVER_ERROR_ALL}${NC}"
echo ""

if [ $SERVER_ERROR_ALL -eq 0 ]; then
    echo -e "${GREEN}✓ Все эндпоинты доступны!${NC}"
    if [ $CLIENT_ERROR_ALL -gt 0 ]; then
        echo -e "${YELLOW}⚠ Некоторые запросы вернули 4xx (это может быть нормально для валидации)${NC}"
    fi
    if [ $TOTAL_ALL -gt 0 ] && [ $SUCCESS_ALL -gt 0 ]; then
        success_rate=$((SUCCESS_ALL * 100 / TOTAL_ALL))
        echo -e "${GREEN}✓ Успешность: ${success_rate}%${NC}"
    fi
else
    echo -e "${RED}✗ Обнаружены серверные ошибки!${NC}"
fi
echo "=========================================="
