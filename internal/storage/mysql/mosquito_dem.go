package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"vue-golang/internal/storage"

	"github.com/go-sql-driver/mysql"
)

func (s *Storage) GetMosquitoOrderDetails(ctx context.Context, requestedID int64) (*storage.GetOrderDetails, error) {
	const op = "storage.mysql.GetMosquitoOrderDetails"

	stmt := `SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position, ready_date
		FROM dem_product_instances_al 
		WHERE id = ? AND type = 'mosquito' AND part_type = 'main'`

	var item storage.GetOrderDetails
	err := s.db.QueryRowContext(ctx, stmt, requestedID).Scan(&item.ID, &item.OrderNum, &item.Name, &item.Count, &item.TotalTime, &item.CreatedAT, &item.Type, &item.PartType,
		&item.ParentProductID, &item.ParentAssembly, &item.Status, &item.Position, &item.ReadyDate)

	if err == nil {
		ops, err := s.loadOperationsForProduct(ctx, item.ID)
		if err != nil {
			slog.Warn("failed to load operations", "op", op, "product_id", item.ID, "err", err)
		} else {
			item.Operations = ops
		}
		return &item, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("%s: ошибка поиска по newID: %w", op, err)
	}

	var orderNum string
	err = s.db.QueryRowContext(ctx, `SELECT numorders FROM dem_orders WHERE idorders = ? AND class_id = 4`, requestedID).Scan(&orderNum)
	if err != nil {
		return nil, fmt.Errorf("%s: не удалось получить номер заказа из архива (legacy_id=%d): %w", op, requestedID, err)
	}

	err = s.db.QueryRowContext(ctx, `
		SELECT id, order_num, name, count, total_time, created_at, type, 
		       part_type, parent_product_id, parent_assembly, status, position
		FROM dem_product_instances_al 
		WHERE order_num = ? AND type = 'mosquito' AND part_type = 'main' 
		ORDER BY id DESC LIMIT 1
	`, orderNum).Scan(
		&item.ID, &item.OrderNum, &item.Name, &item.Count, &item.TotalTime,
		&item.CreatedAT, &item.Type, &item.PartType, &item.ParentProductID,
		&item.ParentAssembly, &item.Status, &item.Position,
	)

	if err == nil {
		ops, err := s.loadOperationsForProduct(ctx, item.ID)
		if err != nil {
			slog.Warn("failed to load operations", "op", op, "product_id", item.ID, "err", err)
		} else {
			item.Operations = ops
		}
		return &item, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("%s: ошибка поиска по order_num: %w", op, err)
	}

	newItem, err := s.importMosquitoFromLegacy(ctx, requestedID, orderNum)
	if err != nil {
		return nil, fmt.Errorf("%s: импорт не удался: %w", op, err)
	}

	ops, err := s.loadOperationsForProduct(ctx, newItem.ID)
	if err != nil {
		slog.Warn("failed to load operations after import", "op", op, "product_id", newItem.ID, "err", err)
	} else {
		newItem.Operations = ops
	}

	slog.Info("mosquito order imported", "op", op, "legacy_id", requestedID, "new_id", newItem.ID, "order_num", newItem.OrderNum, "ops_count", len(newItem.Operations))

	return newItem, nil
}

func (s *Storage) importMosquitoFromLegacy(ctx context.Context, legacyID int64, orderNum string) (*storage.GetOrderDetails, error) {
	const op = "storage.mysql.importMosquitoFromLegacy"

	stmtOrder := `SELECT idorders, numorders, ordername, 'Москитная сетка', 'mosquito' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = '4'`

	var detail storage.GetOrderDetails

	err := s.db.QueryRowContext(ctx, stmtOrder, legacyID).Scan(&detail.ID, &detail.OrderNum, &detail.Customer, &detail.Name, &detail.Type, &detail.PartType,
		&detail.Status, &detail.TemplateCode)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%s: mosquito order %d not found", op, legacyID)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: query order: %w", op, err)
	}

	stmtParamMs := `SELECT SUM(kol_vo) as kol_vo,SUM(kol_vo * weight * hight) sqr FROM dem_param_moskit WHERE orderid = ?`

	var count, sqr float64
	err = s.db.QueryRowContext(ctx, stmtParamMs, legacyID).Scan(&count, &sqr)
	if err != nil {
		// Если таблица пуста или ошибка, не ломаем весь процесс, просто логируем
		slog.Warn("failed to get mosquito params", "op", op, "err", err)
		count = 0
		sqr = 0
	}

	val := sqr / 1000000.0
	allSqr := math.Round(val*1000) / 1000
	detail.Count = count
	detail.Sqr = allSqr

	stmtTypeMs := `SELECT vid, SUM(kol_vo) FROM dem_param_moskit WHERE orderid = ? GROUP BY vid`
	typeCounts := map[string]int{"vsn": 0, "ms": 0}
	tRows, err := s.db.QueryContext(ctx, stmtTypeMs, legacyID)
	if err != nil {
		return nil, fmt.Errorf("%s: query type ms: %w", op, err)
	}

	defer tRows.Close()
	for tRows.Next() {
		var vid, cnt int
		if err := tRows.Scan(&vid, &cnt); err == nil {
			if vid == 5 || vid == 6 {
				typeCounts["vsn"] += cnt
			} else {
				typeCounts["ms"] += cnt
			}
		}
	}

	vsn := typeCounts["vsn"]
	reg := typeCounts["ms"]
	var typeIzd string
	switch {
	case vsn > 0 && reg > 0:
		typeIzd = fmt.Sprintf("%dvsn+%dms", vsn, reg)
	case vsn > 0:
		typeIzd = fmt.Sprintf("vsn")
	default:
		typeIzd = fmt.Sprintf("ms")
	}

	detail.Operations = make([]storage.NormOperation, 0)

	stmtOps := `
    SELECT 
        d.name_mat,
        SUM(d.allowances) as total_value,
        SUM(d.kol_vo) as total_count
    FROM dem_orderdetails d
    WHERE d.orderid = ? 
      AND d.type_m_id = 30
    GROUP BY d.articul_mat, d.name_mat, d.messure
    ORDER BY FIELD(d.name_mat, 
        'Напиловка',
        'Сборка, опрессовка',
        'Сборка',
        'Скатка',
        'Установка крепежа',
        'Изготовление',
        'Установка защиты (вилатерм)'
    ), d.name_mat ASC`

	opsRows, err := s.db.QueryContext(ctx, stmtOps, legacyID)
	if err != nil {
		return nil, fmt.Errorf("%s: query operations: %w", op, err)
	}
	defer opsRows.Close()

	for opsRows.Next() {
		var oper storage.NormOperation
		var rawHours float64
		var name string

		err := opsRows.Scan(
			&name,
			&rawHours,
			&oper.Count,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: scan operation: %w", op, err)
		}

		oper.Label = name
		oper.Name = name
		oper.Value = rawHours
		oper.Minutes = rawHours * 60.0

		detail.Operations = append(detail.Operations, oper)
	}

	var totalTime float64
	for _, oper := range detail.Operations {
		totalTime += oper.Value
	}
	detail.TotalTime = math.Round(totalTime*1000) / 1000

	status := "in_production"
	newItem := &storage.GetOrderDetails{
		OrderNum:     detail.OrderNum,
		TemplateCode: detail.TemplateCode,
		Name:         detail.Name,
		Customer:     detail.Customer,
		Count:        detail.Count,
		Sqr:          detail.Sqr,
		TotalTime:    detail.TotalTime,
		Type:         "mosquito",
		PartType:     "main",
		Status:       &status,
		Position:     0,
		TypeIzd:      typeIzd,
		Profile:      "",
		Systema:      "",
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: транзакция: %w", op, err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
        INSERT INTO dem_product_instances_al (
            order_num, template_code, name, customer, count, total_time, 
            type, part_type, parent_assembly, status, position, sqr, type_izd, profile, systema
        ) VALUES (?, '0', ?, ?, ?, ?, ?, ?, '', ?, ?, ?, ?, '', '')
    `,
		newItem.OrderNum, newItem.Name, newItem.Customer, newItem.Count, newItem.TotalTime,
		newItem.Type, newItem.PartType, newItem.Status, newItem.Position, newItem.Sqr, newItem.TypeIzd,
	)

	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Рекурсивно читаем то, что вставил конкурент
			return s.GetMosquitoOrderDetails(ctx, legacyID)
		}
		return nil, fmt.Errorf("%s: вставка шапки: %w", op, err)
	}

	newID, _ := res.LastInsertId()
	newItem.ID = newID

	for i, dop := range detail.Operations {
		_, err = tx.ExecContext(ctx, `
            INSERT INTO dem_operation_values_al (
                product_id, operation_name, operation_label, count, value, minutes, sort_operation
            ) VALUES (?, ?, ?, ?, ?, ?, ?)
        `, newID, dop.Name, dop.Name, dop.Count, dop.Value, dop.Minutes, i)

		if err != nil {
			return nil, fmt.Errorf("%s: вставка операции '%s': %w", op, dop.Name, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: коммит: %w", op, err)
	}

	return newItem, nil
}

func (s *Storage) loadOperationsForProduct(ctx context.Context, productID int64) ([]storage.NormOperation, error) {
	const op = "storage.mysql.loadOperationsForProduct"

	stmtOps := `SELECT id, operation_name, operation_label, count, value, minutes, sort_operation 
                FROM dem_operation_values_al 
                WHERE product_id = ? 
                ORDER BY sort_operation ASC, id ASC`

	rowsOps, err := s.db.QueryContext(ctx, stmtOps, productID)
	if err != nil {
		return nil, fmt.Errorf("%s: query operations: %w", op, err)
	}
	defer rowsOps.Close()

	var ops []storage.NormOperation

	stmtExecOper := `SELECT employee_id, actual_minutes, actual_value 
                     FROM dem_operation_executors_al 
                     WHERE product_id = ? AND operation_name = ?`

	for rowsOps.Next() {
		var oper storage.NormOperation
		var opID int64
		var sortOp int

		err := rowsOps.Scan(&opID, &oper.Name, &oper.Label, &oper.Count, &oper.Value, &oper.Minutes, &sortOp)
		if err != nil {
			return nil, fmt.Errorf("%s: scan operation: %w", op, err)
		}

		execRows, err := s.db.QueryContext(ctx, stmtExecOper, productID, oper.Name)
		if err != nil {
			slog.Warn("failed to query executors", "op", op, "err", err)
			// Не прерываем весь процесс из-за одной операции без исполнителей
			oper.AssignedWorkers = []storage.AssignedWorker{}
			ops = append(ops, oper)
			continue
		}

		var workers []storage.AssignedWorker
		for execRows.Next() {
			var ex storage.AssignedWorker
			if err := execRows.Scan(&ex.EmployeeID, &ex.ActualMinutes, &ex.ActualValue); err != nil {
				slog.Warn("failed to scan executor", "op", op, "err", err)
				continue
			}
			workers = append(workers, ex)
		}

		execRows.Close()

		if err = execRows.Err(); err != nil {
			slog.Warn("error iterating executors", "op", op, "err", err)
		}

		oper.AssignedWorkers = workers
		ops = append(ops, oper)
	}

	if err = rowsOps.Err(); err != nil {
		return nil, fmt.Errorf("%s: iteration error: %w", op, err)
	}

	return ops, nil
}
