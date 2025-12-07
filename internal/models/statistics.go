package models

import "time"

type StatisticsPeriod string

const (
	PeriodDay   StatisticsPeriod = "day"
	PeriodWeek  StatisticsPeriod = "week"
	PeriodMonth StatisticsPeriod = "month"
	PeriodYear  StatisticsPeriod = "year"
)

type CategoryStatistics struct {
	CategoryID    int     `json:"category_id"`    // Идентификатор категории
	CategoryName  string  `json:"category_name"`  // Название категории
	CategoryColor string  `json:"category_color"` // Цвет категории
	TotalAmount   float64 `json:"total_amount"`   // Общая сумма расходов в категории
	Count         int     `json:"count"`          // Количество расходов в категории
	Percentage    float64 `json:"percentage"`     // Процент от общей суммы всех расходов
}

type PeriodStatistics struct {
	Period        StatisticsPeriod     `json:"period"`         // Период статистики день неделя месяц год
	StartDate     time.Time            `json:"start_date"`     // Начальная дата периода
	EndDate       time.Time            `json:"end_date"`       // Конечная дата периода
	TotalAmount   float64              `json:"total_amount"`   // Общая сумма расходов за период
	Count         int                  `json:"count"`          // Количество расходов за период
	AverageAmount float64              `json:"average_amount"` // Средняя сумма расхода за период
	ByCategory    []CategoryStatistics `json:"by_category"`    // Статистика по каждой категории
}

type ExpenseDistribution struct {
	CategoryID    int     `json:"category_id"`    // Идентификатор категории
	CategoryName  string  `json:"category_name"`  // Название категории
	CategoryColor string  `json:"category_color"` // Цвет категории
	Amount        float64 `json:"amount"`         // Сумма расходов в категории
	Percentage    float64 `json:"percentage"`     // Процент расходов в категории от общей суммы
}
