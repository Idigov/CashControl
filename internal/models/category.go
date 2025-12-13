package models

import (
	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	UserID    uint   `gorm:"not null;index" json:"user_id"`   // Идентификатор пользователя владельца категории
	Name      string `gorm:"not null" json:"name"`            // Название категории
	Color     string `gorm:"default:'#3B82F6'" json:"color"`  // Цвет категории
	Icon      string `json:"icon"`                            // Иконка категории
	IsDefault bool   `gorm:"default:false" json:"is_default"` // Флаг системной категории по умолчанию

	// Связи
	User     User      `gorm:"foreignKey:UserID" json:"-"`     // Пользователь владелец категории
	Expenses []Expense `gorm:"foreignKey:CategoryID" json:"-"` // Все расходы в этой категории
}

type CreateCategoryRequest struct {
	Name  string `json:"name" binding:"required"` // Название новой категории
	Color string `json:"color"`                   // Цвет категории
	Icon  string `json:"icon"`                    // Иконка категории
}

type UpdateCategoryRequest struct {
	Name  *string `json:"name,omitempty"`  // Новое название категории
	Color *string `json:"color,omitempty"` // Новый цвет категории
	Icon  *string `json:"icon,omitempty"`  // Новая иконка категории
}
