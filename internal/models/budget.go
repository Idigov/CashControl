package models

import (
	"gorm.io/gorm"
)

type Budget struct {
	gorm.Model
	UserID uint    `gorm:"not null;index" json:"user_id"`             // Идентификатор пользователя
	Amount float64 `gorm:"not null;type:decimal(10,2)" json:"amount"` // Сумма месячного бюджета
	Month  int     `gorm:"not null" json:"month"`                     // Номер месяца от 1 до 12 (валидация в сервисе)
	Year   int     `gorm:"not null" json:"year"`                      // Год бюджета

	// Связи
	User User `gorm:"foreignKey:UserID" json:"-"` // Пользователь владелец бюджета
}

type BudgetSummary struct {
	Limit     float64 `json:"limit"`
	Spent     float64 `json:"spent"`
	Remaining float64 `json:"remaining"`
	Percent   int     `json:"percent"`
}

type CreateBudgetRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`        // Сумма бюджета должна быть больше нуля
	Month  int     `json:"month" binding:"required,min=1,max=12"` // Номер месяца от 1 до 12
	Year   int     `json:"year" binding:"required"`               // Год бюджета
}

type UpdateBudgetRequest struct {
	Amount *float64 `json:"amount,omitempty"` // Новая сумма бюджета
	Month  *int     `json:"month,omitempty"`  // Новый номер месяца
	Year   *int     `json:"year,omitempty"`   // Новый год бюджета
}

type BudgetStatus struct {
	Budget      *Budget `json:"budget"`        // Информация о бюджете
	Spent       float64 `json:"spent"`         // Потраченная сумма за период
	Remaining   float64 `json:"remaining"`     // Оставшаяся сумма бюджета
	Percentage  float64 `json:"percentage"`    // Процент использования бюджета
	IsExceeded  bool    `json:"is_exceeded"`   // Флаг превышения бюджета
	IsNearLimit bool    `json:"is_near_limit"` // Флаг приближения к лимиту бюджета
}
