package core_config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Config_DB
	Config_Minio
	Config_Logger
	Config_Server
}
type Config_DB struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Sslmode  string
	Timeout  time.Duration
}

type Config_Minio struct {
	Host         string
	BucketName   string
	RootUser     string
	RootPassword string
	Sslmode      string
}
type Config_Logger struct {
	LogLevel string
}
type Config_Server struct {
	Addr            string
	ShutdownTimeout time.Duration
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("%s не найден, использую дефолт: %s", key, fallback)
	return fallback
}

func Load() *Config {
	godotenv.Load("internal/core/config/.env")
	return &Config{
		Config_DB{
			Host:     getEnv("DB_HOST", "192.168.31.234"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "discord_user"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "discord"),
			Sslmode:  getEnv("DB_SSLMODE", "disable"),
			Timeout:  60 * time.Second,
		},
		Config_Minio{
			Host:         getEnv("MINIO_HOST", "192.168.31.195:9000"),
			BucketName:   getEnv("MINIO_BUCKET_NAME", "onlinechat"),
			RootUser:     getEnv("MINIO_ROOT_USER", "admin"),
			RootPassword: getEnv("MINIO_ROOT_PASSWORD", "password123"),
		},
		Config_Logger{
			LogLevel: getEnv("LOGGER_LEVEL", "INFO"),
		},
		Config_Server{
			Addr:            getEnv("HTTP_ADDR", ":8080"),
			ShutdownTimeout: 60 * time.Second,
		},
	}
}
