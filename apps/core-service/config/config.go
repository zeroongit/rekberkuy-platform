package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from the environment.
// Grouped by domain so it is easy to inject into the relevant layer.
type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Supabase   SupabaseConfig
	AI         AIConfig
	Blockchain BlockchainConfig
	Midtrans   MidtransConfig
	SMTP       SMTPConfig
}

type AppConfig struct {
	Env                string // development | staging | production
	Port               string
	CORSAllowedOrigins []string // comma-separated via CORS_ALLOWED_ORIGINS
}

type DatabaseConfig struct {
	URL                string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetimeHrs int
}

type RedisConfig struct {
	URL string
	TTL struct {
		IdempotencySec int
	}
}

type JWTConfig struct {
	Secret        string
	TokenLifetime string // e.g. "24h"
}

type SupabaseConfig struct {
	URL            string
	AnonKey        string
	ServiceRoleKey string
}

type AIConfig struct {
	ServiceURL string
	// ScreeningEnabled turns on real fraud scoring via the backend-ai verification
	// service. false (default) = use the always-safe stub (development).
	ScreeningEnabled bool
	// UnsafeThreshold is the platform's OWN fraud decision threshold: a transaction
	// is considered unsafe (release refused) when the backend-ai score >= this.
	// core-service owns the decision; backend-ai only provides the score.
	UnsafeThreshold float64
	// FailOpen controls behaviour when backend-ai is unreachable or returns an error.
	// false (default) = fail-closed: refuse to release funds when screening is unavailable.
	// true = fail-open: assume safe (availability over strictness). Not recommended for production.
	FailOpen bool
}

type BlockchainConfig struct {
	AvalancheRPCURL    string
	DeployerPrivateKey string
	ContractAddress    string
	ChainID            int64
}

type MidtransConfig struct {
	ServerKey   string
	ClientKey   string
	Environment string // sandbox | production
	WebhookKey  string
	// DisbursementFee is the ESTIMATED Midtrans disbursement cost per wallet
	// withdrawal, booked at request time. The user-facing gross fee
	// (domain.WithdrawFeeToUser) already covers it; the ledger is trued-up to
	// the real cost when the disbursement is confirmed.
	DisbursementFee int64
}

type SMTPConfig struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

// LoadConfig safely and centrally loads all environment variables.
// Only DATABASE_URL is fatal when empty; the rest get safe defaults
// so the application can still run in development mode without all services.
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("💡 Info: .env file not found, system reading OS environment variables")
	}

	cfg := &Config{
		App: AppConfig{
			Env:                getEnv("APP_ENV", "development"),
			Port:               getEnv("PORT", "8080"),
			CORSAllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		},
		Database: DatabaseConfig{
			URL:                getEnv("DATABASE_URL", ""),
			MaxOpenConns:       getEnvInt("DB_MAX_OPEN_CONNS", 100),
			MaxIdleConns:       getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetimeHrs: getEnvInt("DB_CONN_MAX_LIFETIME_HRS", 1),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", ""),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", ""),
			TokenLifetime: getEnv("JWT_TOKEN_LIFETIME", "24h"),
		},
		Supabase: SupabaseConfig{
			URL:            getEnv("SUPABASE_URL", ""),
			AnonKey:        getEnv("SUPABASE_ANON_KEY", ""),
			ServiceRoleKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		},
		AI: AIConfig{
			ServiceURL:       getEnv("AI_SERVICE_URL", "http://localhost:8081"),
			ScreeningEnabled: getEnvBool("FRAUD_SCREENING_ENABLED", false),
			UnsafeThreshold:  getEnvFloat("FRAUD_UNSAFE_THRESHOLD", 0.5),
			FailOpen:         getEnvBool("FRAUD_FAIL_OPEN", false),
		},
		Blockchain: BlockchainConfig{
			AvalancheRPCURL:    getEnv("AVALANCHE_RPC_URL", ""),
			DeployerPrivateKey: getEnv("DEPLOYER_PRIVATE_KEY", ""),
			ContractAddress:    getEnv("CONTRACT_ADDRESS", ""),
			ChainID:            getEnvInt64("AVALANCHE_CHAIN_ID", 43113), // Fuji testnet default
		},
		Midtrans: MidtransConfig{
			ServerKey:       getEnv("MIDTRANS_SERVER_KEY", ""),
			ClientKey:       getEnv("MIDTRANS_CLIENT_KEY", ""),
			Environment:     getEnv("MIDTRANS_ENVIRONMENT", "sandbox"),
			WebhookKey:      getEnv("MIDTRANS_WEBHOOK_KEY", ""),
			DisbursementFee: getEnvInt64("MIDTRANS_DISBURSEMENT_FEE", 4000),
		},
		SMTP: SMTPConfig{
			Host: getEnv("SMTP_HOST", ""),
			Port: getEnv("SMTP_PORT", "587"),
			User: getEnv("SMTP_USER", ""),
			Pass: getEnv("SMTP_PASS", ""),
			From: getEnv("SMTP_FROM", "no-reply@rekberkuy.id"),
		},
	}
	cfg.Redis.TTL.IdempotencySec = getEnvInt("IDEMPOTENCY_TTL_SEC", 86400) // 24 hours

	// Critical validation
	if cfg.Database.URL == "" {
		log.Fatal("❌ CRITICAL CONFIG ERROR: DATABASE_URL variable is required!")
	}
	if cfg.JWT.Secret == "" {
		// Not fatal: use dev-only default in development mode, reject in production.
		if cfg.App.IsProduction() {
			log.Fatal("❌ CRITICAL CONFIG ERROR: JWT_SECRET is required in production!")
		}
		cfg.JWT.Secret = "rekberkuy-dev-secret-key"
		log.Println("⚠️  WARNING: JWT_SECRET is empty, using development default. DO NOT use in production!")
	}

	return cfg
}

// IsProduction helper for feature/secret gating.
func (a AppConfig) IsProduction() bool { return strings.EqualFold(a.Env, "production") }
func (a AppConfig) IsDevelopment() bool {
	return strings.EqualFold(a.Env, "development") || strings.EqualFold(a.Env, "dev")
}

// RedisEnabled indicates whether Redis is configured (for idempotency/cache).
func (c *Config) RedisEnabled() bool { return c.Redis.URL != "" }

// MidtransEnabled indicates whether Midtrans is configured.
func (c *Config) MidtransEnabled() bool { return c.Midtrans.ServerKey != "" }

// BlockchainEnabled indicates whether the on-chain relayer can be operated.
func (c *Config) BlockchainEnabled() bool {
	return c.Blockchain.AvalancheRPCURL != "" && c.Blockchain.DeployerPrivateKey != "" && c.Blockchain.ContractAddress != ""
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// getEnvSlice reads a comma-separated env var into a trimmed slice. Returns the
// default when the var is empty or contains only blank entries.
func getEnvSlice(key string, defaultValue []string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return defaultValue
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return defaultValue
	}
	return out
}

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		log.Printf("⚠️  Variable %s is invalid (%q), using default %d", key, v, defaultValue)
		return defaultValue
	}
	return n
}

// getEnvBool parses a boolean env var (accepts 1/0, true/false, t/f, yes/no, case-insensitive).
func getEnvBool(key string, defaultValue bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return defaultValue
	}
	switch v {
	case "1", "true", "t", "yes", "y":
		return true
	case "0", "false", "f", "no", "n":
		return false
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	var n int64
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		log.Printf("⚠️  Variable %s is invalid (%q), using default %d", key, v, defaultValue)
		return defaultValue
	}
	return n
}

// getEnvFloat parses a float64 env var. Returns the default when empty or invalid.
func getEnvFloat(key string, defaultValue float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	var f float64
	if _, err := fmt.Sscanf(v, "%f", &f); err != nil {
		log.Printf("⚠️  Variable %s is invalid (%q), using default %f", key, v, defaultValue)
		return defaultValue
	}
	return f
}
