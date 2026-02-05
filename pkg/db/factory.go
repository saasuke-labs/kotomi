package db

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// Config holds database configuration
type Config struct {
	Provider  Provider
	SQLitePath string
	FirestoreProjectID string
}

// NewStore creates a new database store based on the provider configuration
func NewStore(ctx context.Context, cfg Config) (Store, error) {
	switch cfg.Provider {
	case ProviderSQLite:
		if cfg.SQLitePath == "" {
			return nil, fmt.Errorf("SQLite path is required")
		}
		return NewSQLiteAdapter(cfg.SQLitePath)
	case ProviderFirestore:
		if cfg.FirestoreProjectID == "" {
			return nil, fmt.Errorf("Firestore project ID is required")
		}
		return NewFirestoreStore(ctx, cfg.FirestoreProjectID)
	default:
		return nil, fmt.Errorf("unsupported database provider: %s", cfg.Provider)
	}
}

// ConfigFromEnv creates a database configuration from environment variables
func ConfigFromEnv() Config {
	provider := strings.ToLower(os.Getenv("DB_PROVIDER"))
	if provider == "" {
		provider = "sqlite" // Default to SQLite for backward compatibility
	}

	cfg := Config{
		Provider: Provider(provider),
	}

	switch cfg.Provider {
	case ProviderSQLite:
		cfg.SQLitePath = os.Getenv("DB_PATH")
		if cfg.SQLitePath == "" {
			cfg.SQLitePath = "./kotomi.db"
		}
	case ProviderFirestore:
		cfg.FirestoreProjectID = os.Getenv("FIRESTORE_PROJECT_ID")
		// Also check GCP_PROJECT for convenience
		if cfg.FirestoreProjectID == "" {
			cfg.FirestoreProjectID = os.Getenv("GCP_PROJECT")
		}
	}

	return cfg
}
