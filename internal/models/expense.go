package models

import (
	"time"

	"gorm.io/gorm"
)

type Expense struct {
	gorm.Model

	UserID      int       `gorm:"not null;index" json:"user_id"`             // Идентификатор пользователя
	CategoryID  int       `gorm:"not null;index" json:"category_id"`         // Идентификатор категории расхода
	Amount      float64   `gorm:"not null;type:decimal(10,2)" json:"amount"` // Сумма расхода
	Description string    `json:"description"`                               // Описание расхода
	Date        time.Time `gorm:"not null;index" json:"date"`                // Дата расхода
	// Связи
	User     User     `gorm:"foreignKey:UserID" json:"-"`            // Пользователь владелец расхода
	Category Category `gorm:"foreignKey:CategoryID" json:"category"` // Категория расхода
}

type CreateExpenseRequest struct {
	CategoryID  int       `json:"category_id" binding:"required"` // Идентификатор категории расхода
	Amount      float64   `json:"amount" binding:"required,gt=0"` // Сумма расхода должна быть больше нуля
	Description string    `json:"description"`                    // Описание расхода
	Date        time.Time `json:"date" binding:"required"`        // Дата расхода
}

type UpdateExpenseRequest struct {
	CategoryID  *int       `json:"category_id,omitempty"` // Новый идентификатор категории
	Amount      *float64   `json:"amount,omitempty"`      // Новая сумма расхода
	Description *string    `json:"description,omitempty"` // Новое описание расхода
	Date        *time.Time `json:"date,omitempty"`        // Новая дата расхода
}

type ExpenseFilter struct {
	UserID     int        // Идентификатор пользователя для фильтрации
	CategoryID *int       // Идентификатор категории для фильтрации
	StartDate  *time.Time // Начальная дата периода для фильтрации
	EndDate    *time.Time // Конечная дата периода для фильтрации
	MinAmount  *float64   // Минимальная сумма для фильтрации
	MaxAmount  *float64   // Максимальная сумма для фильтрации
}

type ExpenseGroup struct {
	Period   string    `json:"period"`   // Период группировки день неделя месяц
	Date     time.Time `json:"date"`     // Дата начала периода
	Total    float64   `json:"total"`    // Общая сумма расходов за период
	Count    int       `json:"count"`    // Количество расходов за период
	Expenses []Expense `json:"expenses"` // Список расходов в этом периоде
}
