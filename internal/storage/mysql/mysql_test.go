package mysql

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Подключаемся к тестовой БД
	var err error
	testDB, err = sql.Open("mysql", "valera:pass@tcp(localhost:3306)/aaaa?parseTime=true")
	if err != nil {
		panic(fmt.Errorf("не удалось подключиться к тестовой БД: %w", err))
	}
	defer testDB.Close()

	// Проверяем подключение
	if err := testDB.Ping(); err != nil {
		panic(fmt.Errorf("ping failed: %w", err))
	}

	// Запускаем все тесты
	code := m.Run()

	os.Exit(code)
}
