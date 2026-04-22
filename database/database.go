package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Service struct {
	db *sql.DB
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("%s не найден, использую дефолт: %s", key, fallback)
	return fallback
}

func (sv *Service) InitDB() error {

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "192.168.31.234"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "discord_user"),
		getEnv("DB_PASSWORD", "password"),
		getEnv("DB_NAME", "discord"),
		getEnv("SSLMODE", "disable"),
	)

	// 2. Открываем соединение
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	sv.db = db

	// 3. Проверяем соединение (sql.Open не создает само подключение сразу)
	err = db.Ping()
	if err != nil {
		return err
	}

	// создание таблиц
	if _, err = db.Exec(Users); err != nil {
		return err
	}
	if _, err = db.Exec(History_messages); err != nil {
		return err
	}

	fmt.Println("Успешное подключение к базе данных!")
	return nil
}
