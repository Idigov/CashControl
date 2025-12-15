package models

import (
	"gorm.io/gorm"
)

type TelegramAuthRequest struct {
	InitData string `json:"init_data" binding:"required"`
}

type User struct {
	gorm.Model

	TelegramID *int64  `gorm:"uniqueIndex" json:"telegram_id,omitempty"`

	Email    *string `gorm:"uniqueIndex" json:"email,omitempty"`
	Username *string `gorm:"uniqueIndex" json:"username,omitempty"`
	Password *string `json:"-"`

	// Связи
	Expenses          []Expense          `gorm:"foreignKey:UserID" json:"-"`
	Categories        []Category         `gorm:"foreignKey:UserID" json:"-"`
	Budgets           []Budget           `gorm:"foreignKey:UserID" json:"-"`
	RecurringExpenses []RecurringExpense `gorm:"foreignKey:UserID" json:"-"`
	ActivityHistory   []ActivityHistory  `gorm:"foreignKey:UserID" json:"-"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`    // Электронная почта для регистрации
	Username string `json:"username" binding:"required,min=3"` // Имя пользователя для регистрации
	Password string `json:"password" binding:"required,min=6"` // Пароль для регистрации
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"` // Электронная почта для входа
	Password string `json:"password" binding:"required"`    // Пароль для входа
}

type LoginResponse struct {
	Token string `json:"token"` // JWT токен для аутентификации
	User  *User  `json:"user"`  // Информация о пользователе
}
