package client

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"rakit-tiket-be/pkg/entity"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"gitlab.com/threetopia/envgo"
)

type PostgreSQLClient interface {
	GetSQLDB() *sql.DB
	Migration() error
}

type postgreSQLClient struct {
	sqlDB *sql.DB
	dsn   entity.DSNEntity
}

func MakePostgreSQLClientFromEnv() PostgreSQLClient {
	dsn := entity.DSNEntity{
		Host:     envgo.GetString("DB_HOST", "localhost"),
		User:     envgo.GetString("DB_USER", "bramasta"),
		Password: envgo.GetString("DB_PASS", "bramasta"),
		Port:     envgo.GetInt("DB_PORT", 5432),
		SSLMode:  envgo.GetBool("DB_SSL_MODE", false),
		Database: envgo.GetString("DB_DATABASE", "reservation"),
		Schema:   envgo.GetString("DB_SCHEMA", "public"),
		TimeZone: envgo.GetString("DB_TZ", "Asia/Jakarta"),
	}
	return MakePostgreSQLClient(dsn)
}

func MakePostgreSQLClient(dsn entity.DSNEntity) PostgreSQLClient {
	return postgreSQLClient{
		dsn:   dsn,
		sqlDB: dbConn(dsn),
	}
}

func dbConn(dsn entity.DSNEntity) *sql.DB {
	sqlDB, err := sql.Open("postgres", dsn.GetPostgresParam())
	if err != nil {
		return nil
	}

	// Increase pool size for high concurrency war scenarios
	sqlDB.SetMaxOpenConns(500)
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Set lock timeout to prevent long-running locks
	sqlDB.Exec("SET lock_timeout = '10s'")
	sqlDB.Exec("SET statement_timeout = '30s'")

	return sqlDB
}

func (c postgreSQLClient) GetSQLDB() *sql.DB {
	return c.sqlDB
}

func (c postgreSQLClient) Migration() error {
	log.Printf("DSN %s", c.dsn.GetPostgresParam())

	driver, err := postgres.WithInstance(c.sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", envgo.GetString("MIGRATION_PATH", "../scripts")),
		"postgres", driver)

	if err != nil {
		return fmt.Errorf("open migration failed: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migration up failed: %w", err)
	}

	return nil
}
