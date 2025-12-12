package models

import (
	"time"

	"gorm.io/gorm"
)

type ActivityType string

const (
	ActivityTypeExpenseCreated   ActivityType = "expense_created"
	ActivityTypeExpenseUpdated   ActivityType = "expense_updated"
	ActivityTypeExpenseDeleted   ActivityType = "expense_deleted"
	ActivityTypeCategoryCreated  ActivityType = "category_created"
	ActivityTypeCategoryUpdated  ActivityType = "category_updated"
	ActivityTypeCategoryDeleted  ActivityType = "category_deleted"
	ActivityTypeBudgetCreated    ActivityType = "budget_created"
	ActivityTypeBudgetUpdated    ActivityType = "budget_updated"
	ActivityTypeRecurringCreated ActivityType = "recurring_created"
	ActivityTypeRecurringUpdated ActivityType = "recurring_updated"
	ActivityTypeRecurringDeleted ActivityType = "recurring_deleted"
)

type ActivityHistory struct {
	gorm.Model
	UserID       int          `gorm:"not null;index" json:"user_id"`       // Идентификатор пользователя
	ActivityType ActivityType `gorm:"not null;index" json:"activity_type"` // Тип действия создание обновление удаление
	EntityType   string       `gorm:"not null" json:"entity_type"`         // Тип сущности расход категория бюджет регулярный расход
	EntityID     int          `gorm:"not null" json:"entity_id"`           // Идентификатор сущности над которой выполнено действие
	Description  string       `json:"description"`                         // Текстовое описание выполненного действия
	Metadata     string       `gorm:"type:jsonb" json:"metadata"`          // Дополнительные данные действия в формате JSON

	// Связи
	User User `gorm:"foreignKey:UserID" json:"-"` // Пользователь выполнивший действие
}

type CreateActivityLogRequest struct {
	UserID       int                    `json:"user_id"`       // Кто совершил действие
	ActivityType ActivityType           `json:"activity_type"` // Тип действия
	EntityType   string                 `json:"entity_type"`   // На какой сущности
	EntityID     int                    `json:"entity_id"`     // ID сущности
	Description  string                 `json:"description"`   // Текстовое описание
	Metadata     map[string]interface{} `json:"metadata"`      // Дополнительные данные (будут сериализованы в JSONB)
}

type ActivityFilter struct {
	UserID       int           // Идентификатор пользователя для фильтрации
	ActivityType *ActivityType // Тип действия для фильтрации
	EntityType   *string       // Тип сущности для фильтрации
	StartDate    *time.Time    // Начальная дата периода для фильтрации
	EndDate      *time.Time    // Конечная дата периода для фильтрации
	Limit        *int          // количество записей
	Offset       *int          // смещение
}
