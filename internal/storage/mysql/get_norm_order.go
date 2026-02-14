package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"vue-golang/internal/storage"
)

func (s *Storage) GetNormOrder(ctx context.Context, id int64) (*storage.GetOrderDetails, error) {
	const op = "storage.mysql.GetNormOrder"

	stmtOrder := "SELECT order_num, name, count, total_time, created_at, updated_at, type FROM dem_product_instances_al WHERE id = ?"

	stmtOperation := "SELECT operation_name, operation_label, count, value, minutes FROM dem_operation_values_al WHERE product_id = ? ORDER BY sort_operation ASC"

	var res storage.GetOrderDetails

	err := s.db.QueryRowContext(ctx, stmtOrder, id).Scan(&res.OrderNum, &res.Name, &res.Count, &res.TotalTime, &res.CreatedAT, &res.UpdatedAT, &res.Type)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: нормировка не найдена: %w", op, err)
		}
		return nil, fmt.Errorf("%s: ошибка запроса: %w", op, err)
	}

	rows, err := s.db.QueryContext(ctx, stmtOperation, id)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка получения операций: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var opr storage.NormOperation
		err := rows.Scan(&opr.Name, &opr.Label, &opr.Count, &opr.Value, &opr.Minutes)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования операции: %w", op, err)
		}
		res.Operations = append(res.Operations, opr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка при итерации: %w", op, err)
	}

	return &res, nil
}

func (s *Storage) GetNormOrdersByOrderNum(ctx context.Context, orderNum string) ([]*storage.GetOrderDetails, error) {
	const op = "storage.mysql.GetNormOrdersByOrderNum"

	// SQL: получаем все наряды по order_num
	stmt := `
		SELECT
			id, name, count, total_time, created_at, updated_at, type, part_type, parent_assembly, parent_product_id
		FROM dem_product_instances_al
		WHERE order_num = ?
		ORDER BY
			CASE WHEN part_type = 'main' THEN 0 ELSE 1 END,
			id
	`

	// Операции по product_id
	stmtOps := `
		SELECT operation_name, operation_label, count, value, minutes
		FROM dem_operation_values_al
		WHERE product_id = ?
	`

	rows, err := s.db.QueryContext(ctx, stmt, orderNum)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка выполнения запроса: %w", op, err)
	}
	defer rows.Close()

	var results []*storage.GetOrderDetails

	for rows.Next() {
		var detail storage.GetOrderDetails
		var parentAssembly sql.NullString // parent_assembly может быть NULL

		err := rows.Scan(
			&detail.ID,
			&detail.Name,
			&detail.Count,
			&detail.TotalTime,
			&detail.CreatedAT,
			&detail.UpdatedAT,
			&detail.Type,
			&detail.PartType,
			&parentAssembly,
			&detail.ParentProductID,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования наряда: %w", op, err)
		}

		// Обработка parent_assembly
		if parentAssembly.Valid {
			detail.ParentAssembly = parentAssembly.String
		} else {
			detail.ParentAssembly = ""
		}

		// Получаем операции для этого наряда
		opsRows, err := s.db.QueryContext(ctx, stmtOps, detail.ID)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка получения операций для product_id=%d: %w", op, detail.ID, err)
		}

		for opsRows.Next() {
			var opr storage.NormOperation
			if err := opsRows.Scan(&opr.Name, &opr.Label, &opr.Count, &opr.Value, &opr.Minutes); err != nil {
				opsRows.Close()
				return nil, fmt.Errorf("%s: ошибка сканирования операции: %w", op, err)
			}
			detail.Operations = append(detail.Operations, opr)
		}
		opsRows.Close()

		if err := opsRows.Err(); err != nil {
			return nil, fmt.Errorf("%s: ошибка при чтении операций: %w", op, err)
		}

		results = append(results, &detail)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка при итерации строк: %w", op, err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s: не найдено нарядов для order_num=%s: %w", op, orderNum, sql.ErrNoRows)
	}

	return results, nil
}

func (s *Storage) GetNormOrders(ctx context.Context, orderNum, orderType string) ([]storage.GetOrderDetails, error) {
	const op = "storage.mysql.GetNormOrders"

	stmt := `SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status FROM dem_product_instances_al 
        	WHERE 1=1 AND (?='' OR order_num LIKE CONCAT('%', ?, '%')) AND (? = '' OR type = ?) AND part_type='main' ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, stmt, orderNum, orderNum, orderType, orderType)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка получения всех нормированных заказов %w", op, err)
	}
	defer rows.Close()

	var items []storage.GetOrderDetails
	for rows.Next() {
		var item storage.GetOrderDetails
		err = rows.Scan(
			&item.ID,
			&item.OrderNum,
			&item.Name,
			&item.Count,
			&item.TotalTime,
			&item.CreatedAT,
			&item.Type,
			&item.PartType,
			&item.ParentProductID,
			&item.ParentAssembly,
			&item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: сканирование: %w", op, err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка итерации: %w", op, err)
	}

	return items, nil
}

func (s *Storage) GetNormOrderIdSub(ctx context.Context, id int64) ([]*storage.GetOrderDetails, error) {
	const op = "storage.mysql.GetNormOrderIdSub"

	stmt := `
		SELECT 
			pi.id, pi.name, pi.count, pi.total_time, pi.created_at, pi.updated_at, pi.type, pi.part_type, pi.parent_assembly, 
			pi.parent_product_id, pi.order_num, pi.template_code, t.head_name, pi.type_izd, pi.status, pi.ready_date, pi.position
		FROM dem_product_instances_al pi
		LEFT JOIN dem_templates_al t ON pi.template_code = t.code
		WHERE pi.id = ? OR pi.parent_product_id = ?
		ORDER BY 
			CASE WHEN pi.part_type = 'main' THEN 0 ELSE 1 END, 
			pi.id
	`

	stmtOps := `SELECT operation_name, operation_label, count, value, minutes FROM dem_operation_values_al WHERE product_id = ? ORDER BY sort_operation ASC`
	stmtExecOper := ` SELECT employee_id, actual_minutes, actual_value FROM dem_operation_executors_al WHERE product_id = ? AND operation_name = ?`

	rows, err := s.db.QueryContext(ctx, stmt, id, id)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка получения основного и дочернего заказа: %w", op, err)
	}
	defer rows.Close()

	var results []*storage.GetOrderDetails

	for rows.Next() {
		var detail storage.GetOrderDetails
		var parentAssembly sql.NullString

		err := rows.Scan(
			&detail.ID,
			&detail.Name,
			&detail.Count,
			&detail.TotalTime,
			&detail.CreatedAT,
			&detail.UpdatedAT,
			&detail.Type,
			&detail.PartType,
			&parentAssembly,
			&detail.ParentProductID,
			&detail.OrderNum,
			&detail.TemplateCode,
			&detail.HeadName,
			&detail.TypeIzd,
			&detail.Status,
			&detail.ReadyDate,
			&detail.Position,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования: %w", op, err)
		}
		if parentAssembly.Valid {
			detail.ParentAssembly = parentAssembly.String
		} else {
			detail.ParentAssembly = ""
		}

		// Операции
		opsRows, err := s.db.QueryContext(ctx, stmtOps, detail.ID)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка операций для id=%d: %w", op, detail.ID, err)
		}

		for opsRows.Next() {
			var oper storage.NormOperation
			err := opsRows.Scan(&oper.Name, &oper.Label, &oper.Count, &oper.Value, &oper.Minutes)
			if err != nil {
				opsRows.Close()
				return nil, fmt.Errorf("%s: ошибка сканирования операции: %w", op, err)
			}

			// Загрузка исполнителей
			execRows, err := s.db.QueryContext(ctx, stmtExecOper, detail.ID, oper.Name)
			if err != nil {
				opsRows.Close()
				return nil, fmt.Errorf("%s: ошибка загрузки исполнителей для операции %s: %w", op, oper.Name, err)
			}
			defer execRows.Close() // ← безопасное закрытие

			var workers []storage.AssignedWorker
			for execRows.Next() {
				var ex storage.AssignedWorker
				err := execRows.Scan(&ex.EmployeeID, &ex.ActualMinutes, &ex.ActualValue)
				if err != nil {
					opsRows.Close()
					return nil, fmt.Errorf("%s: ошибка сканирования исполнителя: %w", op, err)
				}
				workers = append(workers, ex)
			}
			if err = execRows.Err(); err != nil {
				opsRows.Close()
				return nil, fmt.Errorf("%s: ошибка при чтении исполнителей: %w", op, err)
			}

			oper.AssignedWorkers = workers
			detail.Operations = append(detail.Operations, oper)
		}
		opsRows.Close()

		if err != nil {
			opsRows.Close()
			return nil, fmt.Errorf("%s: ошибка загрузки исполнителей для операции %w", op, err)
		}

		results = append(results, &detail)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: итерация строк: %w", op, err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s: наряд с id=%d не найден", op, id)
	}

	return results, nil
}
