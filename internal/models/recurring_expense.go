package models

import "time"

type RecurringExpenseType string

const (
	RecurringTypeDaily   RecurringExpenseType = "daily"
	RecurringTypeWeekly  RecurringExpenseType = "weekly"
	RecurringTypeMonthly RecurringExpenseType = "monthly"
	RecurringTypeYearly  RecurringExpenseType = "yearly"
)

type RecurringExpense struct {
	ID          int                  `gorm:"primaryKey" json:"id"`
	UserID      int                  `gorm:"not null;index" json:"user_id"`             // Идентификатор пользователя
	CategoryID  int                  `gorm:"not null;index" json:"category_id"`         // Идентификатор категории расхода
	Amount      float64              `gorm:"not null;type:decimal(10,2)" json:"amount"` // Сумма регулярного расхода
	Description string               `json:"description"`                               // Описание регулярного расхода
	Type        RecurringExpenseType `gorm:"not null" json:"type"`                      // Тип повторения ежедневно еженедельно ежемесячно ежегодно
	DayOfMonth  *int                 `json:"day_of_month"`                              // День месяца для ежемесячных расходов от 1 до 31
	DayOfWeek   *int                 `json:"day_of_week"`                               // День недели для еженедельных расходов от 0 до 6 где 0 воскресенье
	IsActive    bool                 `gorm:"default:true" json:"is_active"`             // Флаг активности регулярного расхода
	NextDate    time.Time            `gorm:"not null;index" json:"next_date"`           // Следующая дата автоматического создания расхода
	CreatedAt   time.Time            `json:"created_at"`                                // Дата и время создания записи
	UpdatedAt   time.Time            `json:"updated_at"`                                // Дата и время последнего обновления записи

	// Связи
	User     User     `gorm:"foreignKey:UserID" json:"-"`            // Пользователь владелец регулярного расхода
	Category Category `gorm:"foreignKey:CategoryID" json:"category"` // Категория регулярного расхода
}

type CreateRecurringExpenseRequest struct {
	CategoryID  int                  `json:"category_id" binding:"required"`                            // Идентификатор категории расхода
	Amount      float64              `json:"amount" binding:"required,gt=0"`                            // Сумма расхода должна быть больше нуля
	Description string               `json:"description"`                                               // Описание регулярного расхода
	Type        RecurringExpenseType `json:"type" binding:"required,oneof=daily weekly monthly yearly"` // Тип повторения
	DayOfMonth  *int                 `json:"day_of_month"`                                              // День месяца для ежемесячных расходов
	DayOfWeek   *int                 `json:"day_of_week"`                                               // День недели для еженедельных расходов
}

type UpdateRecurringExpenseRequest struct {
	CategoryID  *int                  `json:"category_id,omitempty"`  // Новый идентификатор категории
	Amount      *float64              `json:"amount,omitempty"`       // Новая сумма расхода
	Description *string               `json:"description,omitempty"`  // Новое описание расхода
	Type        *RecurringExpenseType `json:"type,omitempty"`         // Новый тип повторения
	DayOfMonth  *int                  `json:"day_of_month,omitempty"` // Новый день месяца
	DayOfWeek   *int                  `json:"day_of_week,omitempty"`  // Новый день недели
	IsActive    *bool                 `json:"is_active,omitempty"`    // Новый статус активности
}
