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

	//dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=%v",
	//	cfg.DBUser,
	//	cfg.DBPassword,
	//	cfg.DBHost,
	//	cfg.DBPort,
	//	cfg.DBName,
	//	cfg.ParseTime,
	//)
	//db, err := sql.Open("mysql", dsn)
	//if err != nil {
	//	return nil, fmt.Errorf("%s: failed to open db: %w", op, err)
	//}

	//db, err := sql.Open("mysql", "root:@tcp(mysql-8.0:3306)/test_new_logic?parseTime=true")
	//db, err := sql.Open("mysql", "Kuznecov_av:BV02y0Xer72a@tcp(192.168.2.10:3306)/demetra_test?parseTime=true")
	//ubuntu
	//db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/test_new_logic?parseTime=true")
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/aaaa?parseTime=true")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}
