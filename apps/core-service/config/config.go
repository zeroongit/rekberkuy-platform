package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	AppPort     string
	DatabaseURL string
	MidtransServerKey string
	Environment string
}

// LoadConfig memuat semua env variables secara aman dan terpusat
func LoadConfig() *Config {
	// Muat file .env jika ada (di produksi akan membaca OS environment)
	if err := godotenv.Load(); err != nil {
		log.Println("💡 Info: File .env tidak ditemukan, sistem membaca variabel lingkungan OS")
	}

	return &Config{
		AppPort:           getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		MidtransServerKey: getEnv("MIDTRANS_SERVER_KEY", ""),
		Environment:       getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue == "" {
			log.Fatalf("❌ CRITICAL CONFIG ERROR: Variabel %s wajib diisi!", key)
		}
		return defaultValue
	}
	return value
}