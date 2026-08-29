package migrate

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
)

type Migrator struct {
	db *sql.DB
}

func New(db *sql.DB) *Migrator {
	return &Migrator{
		db: db,
	}
}

func Up(path string, dsn string) error {

	m, err := migrate.New(
		"file://"+path,
		"mysql://"+dsn,
	)

	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	defer m.Close()

	err = m.Up()

	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
