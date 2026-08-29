package mysql

import (
	"fmt"
	"vue-golang/internal/config"
)

func DSN(cfg config.Config) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=%v",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.ParseTime,
	)
}

func MigrateDSN(cfg config.Config) string {
	return fmt.Sprintf(
		"mysql://%s:%s@tcp(%s:%d)/%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
}
