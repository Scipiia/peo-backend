package migration

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

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
		return fmt.Errorf("apply migration: %w", err)
	}

	version, dirty, err := m.Version()

	if err == nil {
		fmt.Printf("migration version: %d dirty: %v\n", version, dirty)
	}

	return nil
}
