package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr      string
	SecretKey []byte
	Database  *DatabaseConfig
}

type DatabaseConfig struct {
	URI             string
	MaxOpenConns    int
	MaxIdleConns    int
	MinIdleConns    int
	ConnMaxLifetime int
}

func New() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file found: %v", err)
	}

	return &Config{
		Addr:      GetEnv("ADDR", ":8080"),
		SecretKey: []byte(GetEnv("SECRET_KEY", "")),
		Database: &DatabaseConfig{
			URI:             GetEnv("DATABASE_URI", ""),
			MaxOpenConns:    GetEnvAsInt("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    GetEnvAsInt("DATABASE_MAX_IDLE_CONNS", 25),
			MinIdleConns:    GetEnvAsInt("DATABASE_MIN_IDLE_CONNS", 25),
			ConnMaxLifetime: GetEnvAsInt("DATABASE_CONN_MAX_LIFETIME", 60),
		},
	}
}

func GetEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("environment variable %s is not set", key)
		return defaultValue
	}

	return value

}

func GetEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("environment variable %s is not set, using default %d", key, defaultValue)
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("error converting environment variable %s to int: %v", key, err)
		return defaultValue
	}
	return intValue
}
