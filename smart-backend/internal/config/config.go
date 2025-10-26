package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database         DatabaseConfig
	JWT              JWTConfig
	Server           ServerConfig
	CORS             CORSConfig
	InitAdmin        InitAdminConfig
	ThirdPartyAPIURL string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type ServerConfig struct {
	Port string
	Env  string
}

type CORSConfig struct {
	AllowedOrigins string
}

type InitAdminConfig struct {
	UUID     string
	Username string
	Password string
}

var AppConfig *Config

// LoadConfig loads environment variables and initializes the global config
func LoadConfig() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Parse token expiry durations
	accessExpiry, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		log.Fatal("Invalid JWT_ACCESS_EXPIRY format:", err)
	} else {
		log.Println("JWT_ACCESS_EXPIRY set to:", accessExpiry)
	}


	refreshExpiry, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "720h"))
	if err != nil {
		log.Fatal("Invalid JWT_REFRESH_EXPIRY format:", err)
	} else{
		log.Println("JWT_REFRESH_EXPIRY set to:", refreshExpiry)
	}

	AppConfig = &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "ololo_gate"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		},
		InitAdmin: InitAdminConfig{
			UUID:     getEnv("INIT_ADMIN_UUID", "00000000-0000-0000-0000-000000000001"),
			Username: getEnv("INIT_ADMIN", "admin"),
			Password: getEnv("INIT_ADMIN_PASSWORD", "admin"),
		},
		ThirdPartyAPIURL: getEnv("THIRD_PARTY_API_URL", "https://localhost:3000"),
	}

	log.Println("✅ Configuration loaded successfully")
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
