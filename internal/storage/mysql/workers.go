package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"vue-golang/internal/storage"
)

func (s *Storage) GetAllWorkers(ctx context.Context) ([]storage.GetWorkers, error) {
	const op = "storage.mysql.GetWorkers"

	stmt := "SELECT id, name FROM dem_employees_al WHERE is_active = TRUE"

	var workers []storage.GetWorkers

	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка получения всех работников: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var worker storage.GetWorkers

		err := rows.Scan(&worker.ID, &worker.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования строк для всех сотрудников: %w", op, err)
		}

		workers = append(workers, worker)
	}

	return workers, nil
}

func (s *Storage) SaveOperationWorkers(ctx context.Context, req storage.SaveWorkers) error {
	const op = "storage.mysql.SaveOperationWorkers"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin transaction: %w", op, err)
	}
	defer tx.Rollback()

	// удаляем корень + дети
	_, err = tx.ExecContext(ctx, `
		DELETE FROM dem_operation_executors_al
		WHERE product_id = ? 
		   OR product_id IN (
		       SELECT * FROM (
		           SELECT id FROM dem_product_instances_al WHERE parent_product_id = ?
		       ) AS tmp
		   )
	`, req.RootProductID, req.RootProductID)
	if err != nil {
		return fmt.Errorf("%s: ошибка удаления старых назначении с id=%d %w", op, req.RootProductID, err)
	}

	stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO dem_operation_executors_al 
        (product_id, operation_name, employee_id, actual_minutes, notes, actual_value)
        VALUES (?, ?, ?, ?, ?, ?)
        ON DUPLICATE KEY UPDATE
            actual_minutes = VALUES(actual_minutes),
            actual_value = VALUES(actual_value),
            notes = VALUES(notes),
            updated_at = CURRENT_TIMESTAMP
    `)
	if err != nil {
		return fmt.Errorf("%s: ошибка подготовки запроса: %w", op, err)
	}
	defer stmt.Close()

	for _, a := range req.Assignments {
		_, err := stmt.Exec(
			a.ProductID,
			a.OperationName,
			a.EmployeeID,
			a.ActualMinutes,
			a.Notes,
			a.ActualValue,
		)
		if err != nil {
			return fmt.Errorf("%s: ошибка вставки новых назначенных сотрудников для нормировки с id=%d , op=%s: %w", op, a.ProductID, a.OperationName, err)
		}
	}

	//Если указано — обновляем статус всей сборки
	if req.UpdateStatus != "" && req.RootProductID != 0 {
		// Обновляем main + все его sub
		if err := s.UpdateStatusTx(ctx, tx, req.RootProductID, req.UpdateStatus); err != nil {
			return fmt.Errorf("%s: ошибка обновления статуса для родительского заказа id= %d: %w", op, req.RootProductID, err)
		}

	}

	if req.ReadyDate != "" {
		if err := s.SaveReadyDate(ctx, tx, req.RootProductID, req.ReadyDate); err != nil {
			return fmt.Errorf("%s: ошибка обновления даты готовности для родительского заказа id= %d: %w", op, req.RootProductID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return nil
}

func (s *Storage) SaveReadyDate(ctx context.Context, tx *sql.Tx, rootProductID int64, readyDate string) error {
	const op = "storage.mysql.SaveReadyDate"

	stmtInsertReadyDate := `UPDATE dem_product_instances_al SET ready_date = ? WHERE id = ? OR parent_product_id = ?`
	//stmtInsertReadyDate := `INSERT INTO dem_product_instances_al (ready_date) VALUES (?) WHERE id = ? OR parent_product_id = ?`

	_, err := tx.ExecContext(ctx, stmtInsertReadyDate, readyDate, rootProductID, rootProductID)
	if err != nil {
		return fmt.Errorf("%s: ошибка обновления даты готовности для родительского заказа id=%d: %w", op, rootProductID, err)
	}

	return nil
}
