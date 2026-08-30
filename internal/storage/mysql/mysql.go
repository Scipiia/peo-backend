package mysql

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"vue-golang/internal/config"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) DB() *sql.DB {
	return s.db
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func New(cfg config.Config) (*Storage, error) {
	const op = "storage.mysql.New"

	db, err := sql.Open("mysql", DSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open db: %w", op, err)
	}

	return &Storage{db: db}, nil
}
