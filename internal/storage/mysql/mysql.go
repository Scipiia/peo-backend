package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"vue-golang/internal/config"
)

type Storage struct {
	db *sql.DB
}

func New(cfg config.Config) (*Storage, error) {
	const op = "storage.mysql.New"

	db, err := sql.Open("mysql", DSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open db: %w", op, err)
	}

	return &Storage{db: db}, nil
}
