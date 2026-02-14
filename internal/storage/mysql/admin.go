package mysql

import (
	"context"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"vue-golang/internal/storage"
)

func (s *Storage) GetAllCoefficientAdmin(ctx context.Context) ([]*storage.CoefficientPEOAdmin, error) {
	const op = "storage.mysql.sql.GetAllCoefficientAdmin"

	stmt := `SELECT id, type, coefficient, is_active FROM dem_coefficient_al`

	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка получения всех коэффициентов ПЭО %w", op, err)
	}
	defer rows.Close()

	var coefs []*storage.CoefficientPEOAdmin

	for rows.Next() {
		coef := &storage.CoefficientPEOAdmin{}

		err := rows.Scan(&coef.ID, &coef.Type, &coef.Coefficient, &coef.IsActive)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования строки для получения всех коэффициентов: %w", op, err)
		}

		coefs = append(coefs, coef)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка при итерации по строкам: %w", op, err)
	}

	return coefs, nil
}

func (s *Storage) UpdateCoefficientPEOAdmin(ctx context.Context, coeffs []storage.CoefficientPEOAdmin) error {
	const op = "storage.mysql.sql.UpdateCoefficientPEOAdmin"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: не удалось начать транзакцию: %w", op, err)
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		UPDATE dem_coefficient_al 
		SET coefficient = ?, is_active = ? 
		WHERE id = ? AND type = ?
	`)
	if err != nil {
		return fmt.Errorf("%s: не удалось подготовить запрос для обновления коэффициентов: %w", op, err)
	}

	for _, coef := range coeffs {
		_, err := stmt.ExecContext(ctx, coef.Coefficient, coef.IsActive, coef.ID, coef.Type)
		if err != nil {
			return fmt.Errorf("%s: ошибка обновления коэффициента id=%d: %w", op, coef.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: ошибка коммита транзакции: %w", op, err)
	}

	return nil
}

func (s *Storage) GetAllEmployeesAdmin(ctx context.Context) ([]*storage.EmployeesAdmin, error) {
	const op = "storage.mysql.sql.GetAllEmployeesAdmin"

	stmt := `SELECT id, name, is_active FROM dem_employees_al`

	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка получения всех сотрудников в админке %w", op, err)
	}
	defer rows.Close()

	var employees []*storage.EmployeesAdmin

	for rows.Next() {
		employer := &storage.EmployeesAdmin{}

		err := rows.Scan(&employer.ID, &employer.Name, &employer.IsActive)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования строк для получения всех сотрудников %w", op, err)
		}

		employees = append(employees, employer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка при итерации по строкам: %w", op, err)
	}

	return employees, nil
}

func (s *Storage) UpdateAllEmployeesAdmin(ctx context.Context, emps []storage.EmployeesAdmin) error {
	const op = "storage.mysql.sql.UpdateAllEmployeesAdmin"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: ошибка при создании транзакции: %w", op, err)
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`UPDATE dem_employees_al 
			SET name = ?, is_active = ? 
			WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("%s: ошибка при подготовке запроса: %w", op, err)
	}

	for _, emp := range emps {
		_, err := stmt.ExecContext(ctx, emp.Name, emp.IsActive, emp.ID)
		if err != nil {
			return fmt.Errorf("%s: ошибка при обновлении сотрудников: %w", op, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: ошибка коммита транзакции: %w", op, err)
	}

	return nil
}

func (s *Storage) CreateEmployerAdmin(ctx context.Context, emp storage.EmployeesAdmin) error {
	const op = "storage.mysql.sql.CreateEmployerAdmin"

	stmt := `INSERT INTO dem_employees_al (name, is_active) VALUES (?, ?)`

	_, err := s.db.ExecContext(ctx, stmt, emp.Name, emp.IsActive)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1452 {
			return fmt.Errorf("%s: Ошибка сохранения шаблона в базу='%s'", op, err)
		}
		return fmt.Errorf("%s: Ошибка сохранения шаблона в базу='%s'", op, err)
	}

	return nil
}
