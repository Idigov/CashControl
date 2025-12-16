package repository

import (
	"cashcontrol/internal/database"

	"gorm.io/gorm"
)

// TxProvider минимальный интерфейс для передачи *gorm.DB в репозитории
type TxProvider interface {
	DB() *gorm.DB
}

type txWrapper struct {
	db *gorm.DB
}

func (t txWrapper) DB() *gorm.DB {
	return t.db
}

// RunInTransaction запускает функцию в транзакции.
// Если fn вернёт ошибку, будет откат.
func RunInTransaction(fn func(tx TxProvider) error) error {
	if database.DB == nil {
		return gorm.ErrInvalidDB
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		return fn(txWrapper{db: tx})
	})
}

