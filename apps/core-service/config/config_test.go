package config

import (
	"testing"
)

// ============================================================================
// CONFIG — UNIT TESTS (white-box, package config)
// ============================================================================

func TestGetEnv_Default(t *testing.T) {
	t.Setenv("TEST_RK_VAR_UNDEF", "")
	if got := getEnv("TEST_RK_VAR_UNDEF", "fallback"); got != "fallback" {
		t.Errorf("getEnv default = %q, want fallback", got)
	}
}

func TestGetEnv_Override(t *testing.T) {
	t.Setenv("TEST_RK_VAR_SET", "value-123")
	if got := getEnv("TEST_RK_VAR_SET", "fallback"); got != "value-123" {
		t.Errorf("getEnv override = %q, want value-123", got)
	}
}

func TestGetEnvInt(t *testing.T) {
	t.Setenv("TEST_RK_INT", "42")
	if got := getEnvInt("TEST_RK_INT", 0); got != 42 {
		t.Errorf("getEnvInt = %d, want 42", got)
	}

	t.Setenv("TEST_RK_INT_BAD", "not-a-number")
	if got := getEnvInt("TEST_RK_INT_BAD", 7); got != 7 {
		t.Errorf("getEnvInt invalid = %d, want default 7", got)
	}

	if got := getEnvInt("TEST_RK_INT_UNSET", 99); got != 99 {
		t.Errorf("getEnvInt unset = %d, want default 99", got)
	}
}

func TestGetEnvInt64(t *testing.T) {
	t.Setenv("TEST_RK_INT64", "43113")
	if got := getEnvInt64("TEST_RK_INT64", 0); got != 43113 {
		t.Errorf("getEnvInt64 = %d, want 43113", got)
	}
}

func TestLoadConfig_HappyPath(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("PORT", "9090")
	t.Setenv("APP_ENV", "staging")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("MIDTRANS_SERVER_KEY", "SB-MID-xxx")
	t.Setenv("AVALANCHE_RPC_URL", "https://api.avax-test.network/ext/bc/C/rpc")
	t.Setenv("DEPLOYER_PRIVATE_KEY", "0xabc")
	t.Setenv("CONTRACT_ADDRESS", "0xcontract")

	cfg := LoadConfig()

	if cfg.App.Port != "9090" {
		t.Errorf("App.Port = %q, want 9090", cfg.App.Port)
	}
	if cfg.App.Env != "staging" {
		t.Errorf("App.Env = %q, want staging", cfg.App.Env)
	}
	if cfg.Database.URL != "postgres://u:p@localhost:5432/db" {
		t.Errorf("Database.URL not read")
	}
	if cfg.Database.MaxOpenConns != 100 {
		t.Errorf("MaxOpenConns default = %d, want 100", cfg.Database.MaxOpenConns)
	}
	if cfg.JWT.Secret != "super-secret" {
		t.Errorf("JWT.Secret not read")
	}
	if cfg.Blockchain.ChainID != 43113 {
		t.Errorf("ChainID default = %d, want 43113 (Fuji)", cfg.Blockchain.ChainID)
	}
	if cfg.Redis.TTL.IdempotencySec != 86400 {
		t.Errorf("Idempotency TTL default = %d, want 86400", cfg.Redis.TTL.IdempotencySec)
	}
	// AI: backend-ai owns the Groq key; core-service owns the decision threshold.
	if cfg.AI.ScreeningEnabled {
		t.Error("AI.ScreeningEnabled default should be false")
	}
	if cfg.AI.UnsafeThreshold != 0.5 {
		t.Errorf("AI.UnsafeThreshold default = %v, want 0.5", cfg.AI.UnsafeThreshold)
	}
}

func TestLoadConfig_DevDefaultsJWT(t *testing.T) {
	// In development, an empty JWT_SECRET may use the default (not fatal).
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db")
	t.Setenv("APP_ENV", "development")
	t.Setenv("JWT_SECRET", "")

	cfg := LoadConfig()
	if cfg.JWT.Secret == "" {
		t.Error("JWT.Secret should be filled with development default when empty")
	}
	if cfg.JWT.Secret == "super-secret" {
		t.Error("default JWT secret must not match test input")
	}
}

func TestFlags(t *testing.T) {
	cfg := &Config{
		Redis:    RedisConfig{URL: "redis://localhost:6379"},
		Midtrans: MidtransConfig{ServerKey: "SB-MID-xxx"},
		Blockchain: BlockchainConfig{
			AvalancheRPCURL:    "https://rpc",
			DeployerPrivateKey: "0xkey",
			ContractAddress:    "0xcontract",
		},
	}
	if !cfg.RedisEnabled() {
		t.Error("RedisEnabled should be true")
	}
	if !cfg.MidtransEnabled() {
		t.Error("MidtransEnabled should be true")
	}
	if !cfg.BlockchainEnabled() {
		t.Error("BlockchainEnabled should be true")
	}

	empty := &Config{}
	if empty.RedisEnabled() || empty.MidtransEnabled() || empty.BlockchainEnabled() {
		t.Error("empty config should have all flags false")
	}
}

func TestAppEnvHelpers(t *testing.T) {
	if !(AppConfig{Env: "production"}).IsProduction() {
		t.Error("IsProduction should be true for 'production'")
	}
	if !(AppConfig{Env: "PRODUCTION"}).IsProduction() {
		t.Error("IsProduction should be case-insensitive")
	}
	if !(AppConfig{Env: "development"}).IsDevelopment() {
		t.Error("IsDevelopment should be true for 'development'")
	}
	if !(AppConfig{Env: "dev"}).IsDevelopment() {
		t.Error("IsDevelopment should be true for 'dev'")
	}
	if (AppConfig{Env: "staging"}).IsProduction() {
		t.Error("IsProduction should be false for staging")
	}
}
