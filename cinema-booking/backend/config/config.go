package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	// Server
	Port string

	// MongoDB
	MongoURI string
	MongoDB  string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// JWT
	JWTSecret string

	// RabbitMQ
	RabbitMQURL string

	// Frontend
	FrontendURL string

	// Admin
	AdminEmails string

	// Lock TTL (seconds)
	SeatLockTTL int
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		MongoURI:           getEnv("MONGO_URI", "mongodb://mongo:27017"),
		MongoDB:            getEnv("MONGO_DB", "cinema"),
		RedisAddr:          getEnv("REDIS_ADDR", "redis:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            getEnvInt("REDIS_DB", 0),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production"),
		RabbitMQURL:        getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		AdminEmails:        getEnv("ADMIN_EMAILS", ""),
		SeatLockTTL:        getEnvInt("SEAT_LOCK_TTL", 300),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			log.Printf("Invalid int for %s: %s", key, v)
			return fallback
		}
		return n
	}
	return fallback
}
