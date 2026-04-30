package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Config_DB
	Config_Minio
}
type Config_DB struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Sslmode  string
}

type Config_Minio struct {
	Host         string
	BucketName   string
	RootUser     string
	RootPassword string
	Sslmode      string
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
		Config_Minio{
			Host:         getEnv("MINIO_HOST", "192.168.31.195:9000"),
			BucketName:   getEnv("MINIO_BUCKET_NAME", "onlinechat"),
			RootUser:     getEnv("MINIO_ROOT_USER", "admin"),
			RootPassword: getEnv("MINIO_ROOT_PASSWORD", "password123"),
		},
	}
}
