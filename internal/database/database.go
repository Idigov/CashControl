package database

import (
	"database/sql"
	"fmt"
	"strings"

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

	if !exists {
		quotedDBName := fmt.Sprintf(`"%s"`, strings.ReplaceAll(cfg.DBName, `"`, `""`))
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", quotedDBName))
		if err != nil {
			return fmt.Errorf("ошибка создания БД: %w", err)
		}
	}

	return nil
}

func Init(cfg *config.Config) error {
	if !cfg.UseDatabaseURL && (cfg.DBHost == "localhost" || cfg.DBHost == "127.0.0.1") {
		if err := createDatabaseIfNotExists(cfg); err != nil {
			return fmt.Errorf("ошибка создания БД: %w", err)
		}
	}
	var err error

	var dsn string
	if cfg.UseDatabaseURL {
		dsn = cfg.DatabaseURL
	} else {
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

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("ошибка получения sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "no route to host") || strings.Contains(errStr, "ipv6") || strings.Contains(errStr, "network is unreachable") {
			return fmt.Errorf("ошибка подключения к БД (IPv6 недоступен): %w", err)
		}
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	return nil
}

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
