package mysql

import (
	"context"
	"fmt"
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

	// 1. Получаем всех сотрудников
	employeeQuery := `
        SELECT id, name, is_active 
        FROM dem_employees_al 
        ORDER BY id`

	rows, err := s.db.QueryContext(ctx, employeeQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка получения сотрудников %w", op, err)
	}
	defer rows.Close()

	employeesMap := make(map[int]*storage.EmployeesAdmin)

	for rows.Next() {
		emp := &storage.EmployeesAdmin{
			Teams: make([]*storage.TeamAdmin, 0), // или []*storage.Team
		}
		if err := rows.Scan(&emp.ID, &emp.Name, &emp.IsActive); err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования сотрудника %w", op, err)
		}
		employeesMap[int(emp.ID)] = emp
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка итерации сотрудников %w", op, err)
	}

	// 2. Получаем все связи сотрудник ↔ команда (один запрос на все)
	teamQuery := `
        SELECT et.employee_id, t.id AS team_id, t.name AS team_name, 
               t.slug AS team_slug, t.is_active AS team_is_active 
        FROM dem_employee_teams_al et
        JOIN dem_teams_al t ON et.team_id = t.id
        ORDER BY et.employee_id`

	rows, err = s.db.QueryContext(ctx, teamQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка получения связей с командами %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var employeeID int
		team := &storage.TeamAdmin{}
		if err := rows.Scan(&employeeID, &team.ID, &team.Name, &team.Slug, &team.IsActive); err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования команды %w", op, err)
		}

		if emp, ok := employeesMap[employeeID]; ok {
			emp.Teams = append(emp.Teams, team)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка итерации связей %w", op, err)
	}

	// Преобразуем map в slice
	var employees []*storage.EmployeesAdmin
	for _, emp := range employeesMap {
		employees = append(employees, emp)
	}

	return employees, nil
}

func (s *Storage) GetAllTeamsAdmin(ctx context.Context) ([]*storage.TeamAdmin, error) {
	const op = "storage.mysql.sql.GetAllTeamsAdmin"

	stmt := `SELECT id, name, slug, is_active FROM dem_teams_al WHERE is_active = 1`

	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка получения команд %w", op, err)
	}

	defer rows.Close()

	var teams []*storage.TeamAdmin
	for rows.Next() {
		team := &storage.TeamAdmin{}
		err := rows.Scan(&team.ID, &team.Name, &team.Slug, &team.IsActive)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования команд %w", op, err)
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (s *Storage) UpdateAllEmployeesAdmin(ctx context.Context, id int64, input storage.UpdateEmployeeInput) error {
	const op = "storage.mysql.sql.UpdateEmployeeWithTeams"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: начало транзакции: %w", op, err)
	}
	defer tx.Rollback()

	// 1. Обновляем данные сотрудника
	_, err = tx.ExecContext(ctx,
		`UPDATE dem_employees_al SET name = ?, is_active = ? WHERE id = ?`,
		input.Name, input.IsActive, id,
	)
	if err != nil {
		return fmt.Errorf("%s: обновление сотрудника: %w", op, err)
	}

	// 2. Удаляем старые связи
	_, err = tx.ExecContext(ctx,
		`DELETE FROM dem_employee_teams_al WHERE employee_id = ?`, id,
	)
	if err != nil {
		return fmt.Errorf("%s: удаление старых связей: %w", op, err)
	}

	// 3. Вставляем новые связи
	if len(input.TeamIDs) > 0 {
		stmt := `INSERT INTO dem_employee_teams_al (employee_id, team_id) VALUES (?, ?)`
		for _, team := range input.TeamIDs {
			_, err := tx.ExecContext(ctx, stmt, id, team)
			if err != nil {
				return fmt.Errorf("%s: вставка новой связи: %w", op, err)
			}
		}
	}

	return tx.Commit()
}

func (s *Storage) CreateEmployeeAdmin(ctx context.Context, input storage.CreateEmployeeInput) error {
	const op = "storage.mysql.sql.CreateEmployeeWithTeams"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: начало транзакции: %w", op, err)
	}
	defer tx.Rollback()

	empStmt := `INSERT INTO dem_employees_al (name, is_active) VALUES (?, ?)`
	res, err := tx.ExecContext(ctx, empStmt, input.Name, input.IsActive)
	if err != nil {
		return fmt.Errorf("%s: вставка сотрудника: %w", op, err)
	}

	empID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("%s: получение последнего вставленного ID: %w", op, err)
	}

	if len(input.TeamIDs) > 0 {
		teamStmt := `INSERT INTO dem_employee_teams_al (employee_id, team_id) VALUES (?, ?)`
		for _, teamID := range input.TeamIDs {
			if teamID <= 0 {
				continue
			}
			_, err := tx.ExecContext(ctx, teamStmt, empID, teamID)
			if err != nil {
				return fmt.Errorf("%s: вставка связи с командой %d: %w", op, teamID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: коммит транзакции: %w", op, err)
	}

	return nil
}
