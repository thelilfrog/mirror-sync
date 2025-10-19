package storage

import (
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func (r *Repository) Migrate() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}

	if err := goose.Up(r.db, "migrations"); err != nil {
		return fmt.Errorf("failed to migrate the database: %s", err)
	}

	return nil
}
