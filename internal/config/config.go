package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Config_DB
}
type Config_DB struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Sslmode  string
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("%s не найден, использую дефолт: %s", key, fallback)
	return fallback
}

func Load() *Config {
	godotenv.Load("internal/config/.env")
	return &Config{
		Config_DB{
			Host:     getEnv("DB_HOST", "192.168.31.234"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "discord_user"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "discord"),
			Sslmode:  getEnv("SSLMODE", "disable"),
		},
	}
}
