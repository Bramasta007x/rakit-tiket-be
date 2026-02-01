package migration

import (
	"database/sql"
	"fmt"
	"log"
	"rakit-tiket-be/pkg/entity"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"gitlab.com/threetopia/envgo"
)

func DoMigration(dsn entity.DSNEntity) error {
	log.Printf("DSN %s", dsn.GetPostgresParam())

	db, err := sql.Open("postgres", dsn.GetPostgresParam())
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", envgo.GetString("MIGRATION_PATH", "../script")),
		"postgres", driver)
	m.Up()
	if err != nil {
		log.Fatalf("Error %s", err.Error())
		return err
	}

	return nil
}
