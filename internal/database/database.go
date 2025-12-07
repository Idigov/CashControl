package database

import (
	"database/sql"
	"fmt"

	"cashcontrol/internal/config"
	"cashcontrol/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// createDatabaseIfNotExists создает базу данных, если её не существует
func createDatabaseIfNotExists(cfg *config.Config) error {
	// Подключаемся к базе данных postgres
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBPort,
		cfg.DBSSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("ошибка подключения к postgres: %w", err)
	}
	defer db.Close()

	// Проверяем, существует ли база данных
	var exists bool
	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		cfg.DBName,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка проверки существования БД: %w", err)
	}

	// Создаем базу данных, если её нет
	if !exists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
		if err != nil {
			return fmt.Errorf("ошибка создания БД: %w", err)
		}
	}

	return nil
}

// Init инициализирует подключение к базе данных
func Init(cfg *config.Config) error {
	// Создаем базу данных, если её нет (только если используется отдельные параметры, не DATABASE_URL)
	if cfg.DatabaseURL == "" {
		if err := createDatabaseIfNotExists(cfg); err != nil {
			return fmt.Errorf("ошибка создания БД: %w", err)
		}
	}
	var err error

	dsn := cfg.DatabaseURL
	if dsn == "" {
		// Формируем DSN из отдельных параметров, если DATABASE_URL не указан
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			cfg.DBHost,
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBName,
			cfg.DBPort,
			cfg.DBSSLMode,
		)
	}

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// Проверка подключения
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("ошибка получения sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ошибка ping БД: %w", err)
	}

	return nil
}

// Migrate выполняет миграции базы данных
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("база данных не инициализирована")
	}

	err := DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Expense{},
		&models.Budget{},
		&models.RecurringExpense{},
		&models.ActivityHistory{},
	)
	if err != nil {
		return fmt.Errorf("ошибка миграции: %w", err)
	}

	return nil
}

// Close закрывает подключение к базе данных
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
