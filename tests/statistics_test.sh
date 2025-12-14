#!/bin/bash

# Тесты для Statistics эндпоинтов

source "$(dirname "$0")/common.sh"

echo "=== Statistics эндпоинты ==="

# Создаем тестового пользователя
USER_ID=$(create_test_user "statistics_test_$(date +%s)@example.com" "statisticstest")
if [ -z "$USER_ID" ]; then
    echo "  ⚠ Не удалось создать пользователя, используем ID=1"
    USER_ID=1
else
    echo "  ✓ Тестовый пользователь создан (ID: $USER_ID)"
fi

# Создаем несколько категорий для тестирования
category1_response=$(curl -s -w "\n%{http_code}" -X "POST" "$BASE_URL/categories/$USER_ID" \
    -H "Content-Type: application/json" \
    -d '{"name":"Категория 1","description":"Первая категория","color":"#FF5733"}')
category1_code=$(echo "$category1_response" | tail -n1)
category1_body=$(echo "$category1_response" | sed '$d')
if [ "$category1_code" -ge 200 ] && [ "$category1_code" -lt 300 ]; then
    CATEGORY1_ID=$(extract_id "$category1_body")
    echo "  ✓ Категория 1 создана (ID: $CATEGORY1_ID)"
else
    echo "  ⚠ Не удалось создать категорию 1, используем ID=1"
    CATEGORY1_ID=1
fi

category2_response=$(curl -s -w "\n%{http_code}" -X "POST" "$BASE_URL/categories/$USER_ID" \
    -H "Content-Type: application/json" \
    -d '{"name":"Категория 2","description":"Вторая категория","color":"#33FF57"}')
category2_code=$(echo "$category2_response" | tail -n1)
category2_body=$(echo "$category2_response" | sed '$d')
if [ "$category2_code" -ge 200 ] && [ "$category2_code" -lt 300 ]; then
    CATEGORY2_ID=$(extract_id "$category2_body")
    echo "  ✓ Категория 2 создана (ID: $CATEGORY2_ID)"
else
    echo "  ⚠ Не удалось создать категорию 2, используем ID=2"
    CATEGORY2_ID=2
fi

# Создаем расходы для тестирования статистики
if [ -n "$CATEGORY1_ID" ] && [ -n "$CATEGORY2_ID" ]; then
    echo "  → Создание тестовых расходов..."
    
    # Получаем текущую дату и даты для разных периодов
    current_date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    yesterday_date=$(date -u -v-1d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "1 day ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ")
    week_ago_date=$(date -u -v-7d +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -d "7 days ago" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Расходы в категории 1
    curl -s -X "POST" "$BASE_URL/expenses?user_id=$USER_ID" \
        -H "Content-Type: application/json" \
        -d "{\"category_id\":$CATEGORY1_ID,\"amount\":1000.50,\"description\":\"Расход 1\",\"date\":\"$current_date\"}" > /dev/null
    
    curl -s -X "POST" "$BASE_URL/expenses?user_id=$USER_ID" \
        -H "Content-Type: application/json" \
        -d "{\"category_id\":$CATEGORY1_ID,\"amount\":500.25,\"description\":\"Расход 2\",\"date\":\"$yesterday_date\"}" > /dev/null
    
    # Расходы в категории 2
    curl -s -X "POST" "$BASE_URL/expenses?user_id=$USER_ID" \
        -H "Content-Type: application/json" \
        -d "{\"category_id\":$CATEGORY2_ID,\"amount\":750.75,\"description\":\"Расход 3\",\"date\":\"$current_date\"}" > /dev/null
    
    curl -s -X "POST" "$BASE_URL/expenses?user_id=$USER_ID" \
        -H "Content-Type: application/json" \
        -d "{\"category_id\":$CATEGORY2_ID,\"amount\":300.00,\"description\":\"Расход 4\",\"date\":\"$week_ago_date\"}" > /dev/null
    
    echo "  ✓ Тестовые расходы созданы"
fi

# Небольшая задержка для обработки данных
sleep 1

# Тест 1: Статистика за период (день)
echo ""
echo "  → Тестирование статистики за день..."
test_endpoint "GET" "/statistics/period?user_id=$USER_ID&period=day" "" "Статистика за день"

# Тест 2: Статистика за период (неделя)
test_endpoint "GET" "/statistics/period?user_id=$USER_ID&period=week" "" "Статистика за неделю"

# Тест 3: Статистика за период (месяц)
test_endpoint "GET" "/statistics/period?user_id=$USER_ID&period=month" "" "Статистика за месяц"

# Тест 4: Статистика за период (год)
test_endpoint "GET" "/statistics/period?user_id=$USER_ID&period=year" "" "Статистика за год"

# Тест 5: Статистика за период с указанием дат
if [ -n "$yesterday_date" ] && [ -n "$current_date" ]; then
    start_date=$(echo "$yesterday_date" | cut -d'T' -f1)
    end_date=$(echo "$current_date" | cut -d'T' -f1)
    test_endpoint "GET" "/statistics/period?user_id=$USER_ID&start_date=$start_date&end_date=$end_date" "" "Статистика за период с датами"
fi

# Тест 6: Статистика за период с фильтром по категории
if [ -n "$CATEGORY1_ID" ]; then
    test_endpoint "GET" "/statistics/period?user_id=$USER_ID&period=month&category_id=$CATEGORY1_ID" "" "Статистика за месяц по категории"
fi

# Тест 7: Статистика по категориям (без фильтров)
echo ""
echo "  → Тестирование статистики по категориям..."
test_endpoint "GET" "/statistics/categories?user_id=$USER_ID" "" "Статистика по всем категориям"

# Тест 8: Статистика по категориям с фильтром по датам
if [ -n "$yesterday_date" ] && [ -n "$current_date" ]; then
    start_date=$(echo "$yesterday_date" | cut -d'T' -f1)
    end_date=$(echo "$current_date" | cut -d'T' -f1)
    test_endpoint "GET" "/statistics/categories?user_id=$USER_ID&start_date=$start_date&end_date=$end_date" "" "Статистика по категориям с датами"
fi

# Тест 9: Статистика по категориям с фильтром по категории
if [ -n "$CATEGORY2_ID" ]; then
    test_endpoint "GET" "/statistics/categories?user_id=$USER_ID&category_id=$CATEGORY2_ID" "" "Статистика по конкретной категории"
fi

# Тест 10: Распределение расходов по категориям
echo ""
echo "  → Тестирование распределения расходов..."
test_endpoint "GET" "/statistics/distribution?user_id=$USER_ID" "" "Распределение расходов по категориям"

# Тест 11: Распределение расходов с фильтром по датам
if [ -n "$yesterday_date" ] && [ -n "$current_date" ]; then
    start_date=$(echo "$yesterday_date" | cut -d'T' -f1)
    end_date=$(echo "$current_date" | cut -d'T' -f1)
    test_endpoint "GET" "/statistics/distribution?user_id=$USER_ID&start_date=$start_date&end_date=$end_date" "" "Распределение расходов с датами"
fi

# Тест 12: Валидация - отсутствует user_id
echo ""
echo "  → Тестирование валидации..."
test_endpoint "GET" "/statistics/period?period=month" "" "Отсутствует user_id (ожидается ошибка)"

# Тест 13: Валидация - некорректный период
test_endpoint "GET" "/statistics/period?user_id=$USER_ID&period=invalid" "" "Некорректный период (ожидается ошибка)"

# Тест 14: Валидация - некорректный формат даты
test_endpoint "GET" "/statistics/period?user_id=$USER_ID&start_date=invalid-date" "" "Некорректный формат даты (ожидается ошибка)"

print_stats
