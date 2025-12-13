#!/bin/bash

# Общие функции для тестирования API

BASE_URL="${BASE_URL:-http://localhost:8080}"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Счетчики (инициализируются при каждом вызове)
TOTAL=0
SUCCESS=0
CLIENT_ERROR=0
SERVER_ERROR=0

# Функция для извлечения ID из JSON ответа
extract_id() {
    local json=$1
    echo "$json" | grep -o '"ID":[0-9]*' | head -1 | grep -o '[0-9]*'
}

# Функция для выполнения запроса
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    local save_id_var=$5  # Имя переменной для сохранения ID (опционально)
    
    TOTAL=$((TOTAL + 1))
    echo -n "  Testing $method $endpoint ... "
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" -H "Content-Type: application/json")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        SUCCESS=$((SUCCESS + 1))
        echo -e "${GREEN}✓${NC} ($http_code)"
        if [ -n "$description" ]; then
            echo "    → $description"
        fi
        
        # Сохраняем ID если указано
        if [ -n "$save_id_var" ] && [ -n "$body" ]; then
            id=$(extract_id "$body")
            if [ -n "$id" ]; then
                eval "$save_id_var=$id"
                echo "    → Сохранен ID: $id"
            fi
        fi
        
        return 0
    elif [ "$http_code" -ge 400 ] && [ "$http_code" -lt 500 ]; then
        CLIENT_ERROR=$((CLIENT_ERROR + 1))
        echo -e "${YELLOW}⚠${NC} ($http_code) - Client Error"
        if [ -n "$description" ]; then
            echo "    → $description"
        fi
        return 1
    else
        SERVER_ERROR=$((SERVER_ERROR + 1))
        echo -e "${RED}✗${NC} ($http_code)"
        if [ -n "$description" ]; then
            echo "    → $description"
        fi
        return 1
    fi
}

# Функция для создания пользователя
create_test_user() {
    local email=$1
    local username=$2
    local password=${3:-password123}
    
    local response=$(curl -s -w "\n%{http_code}" -X "POST" "$BASE_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"username\":\"$username\",\"password\":\"$password\"}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        extract_id "$body"
        return 0
    fi
    return 1
}

# Функция для вывода статистики
print_stats() {
    echo ""
    echo "  Статистика:"
    echo "  - Всего тестов: ${TOTAL}"
    echo -e "  - ${GREEN}Успешно: ${SUCCESS}${NC}"
    echo -e "  - ${YELLOW}Клиентские ошибки (4xx): ${CLIENT_ERROR}${NC}"
    echo -e "  - ${RED}Серверные ошибки (5xx): ${SERVER_ERROR}${NC}"
}

