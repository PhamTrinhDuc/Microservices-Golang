package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBConfig struct {
	Host              string        `mapstructure:"DB_HOST"`
	Port              int           `mapstructure:"DB_PORT"`
	User              string        `mapstructure:"DB_USER"`
	Password          string        `mapstructure:"DB_PASSWORD"`
	DBName            string        `mapstructure:"DB_NAME"`
	SSLMode           string        `mapstructure:"DB_SSL_MODE"`
	MaxConns          int32         `mapstructure:"DB_MAX_CONNS"`
	MinConns          int32         `mapstructure:"DB_MIN_CONNS"`
	MaxConnLifetime   time.Duration `mapstructure:"DB_MAX_CONN_LIFETIME"`
	MaxConnIdleTime   time.Duration `mapstructure:"DB_MAX_CONN_IDLE_TIME"`
	HealthCheckPeriod time.Duration `mapstructure:"DB_HEALTH_CHECK_PERIOD"`
}

type DB struct {
	pool *pgxpool.Pool
}

func (cfg *DBConfig) Validate() error {
	if cfg.Host == "" || cfg.User == "" || cfg.DBName == "" {
		return fmt.Errorf("host, user, and dbname are required")
	}

	// Default values
	if cfg.Port <= 0 {
		cfg.Port = 5432
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.MaxConns <= 0 {
		cfg.MaxConns = 25
	}
	if cfg.MinConns < 0 {
		cfg.MinConns = 2
	}
	if cfg.MaxConnLifetime == 0 {
		cfg.MaxConnLifetime = 1 * time.Hour
	}
	if cfg.MaxConnIdleTime == 0 {
		cfg.MaxConnIdleTime = 30 * time.Minute
	}
	if cfg.HealthCheckPeriod == 0 {
		cfg.HealthCheckPeriod = 1 * time.Minute
	}

	return nil
}

func NewDB(ctx context.Context, cfg DBConfig) (*DB, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("db config validation failed: %w", err)
	}

	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *DB) GetPool() *pgxpool.Pool {
	return db.pool
}

func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}
