package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"
	"vue-golang/internal/storage"

	"github.com/go-sql-driver/mysql"
)

// GetGutterOrderDetails универсальный вход для водоотливов/нащельников
func (s *Storage) GetGutterOrderDetails(ctx context.Context, requestedID int64) (*storage.GetOrderDetails, error) {
	const op = "storage.mysql.GetGutterOrderDetails"

	var item storage.GetOrderDetails
	err := s.db.QueryRowContext(ctx, `
		SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position
		FROM dem_product_instances_al 
		WHERE id = ? AND type = 'vodootliv' AND part_type = 'main'
	`, requestedID).Scan(&item.ID, &item.OrderNum, &item.Name, &item.Count, &item.TotalTime, &item.CreatedAT, &item.Type, &item.PartType, &item.ParentProductID, &item.ParentAssembly, &item.Status, &item.Position)

	if err == nil {
		ops, _ := s.loadOperationsForProduct(ctx, item.ID)
		item.Operations = ops
		return &item, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("%s: search by new ID error: %w", op, err)
	}

	var orderNum string
	err = s.db.QueryRowContext(ctx, `SELECT numorders FROM dem_orders WHERE idorders = ? AND class_id = 6`, requestedID).Scan(&orderNum)
	if err != nil {
		return nil, fmt.Errorf("%s: legacy order not found (id=%d): %w", op, requestedID, err)
	}

	var nashchelnikCount int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dem_param_vo WHERE orderid = ? AND mat IN (3,4,7,9)`, requestedID).Scan(&nashchelnikCount)
	if nashchelnikCount > 0 {
		return nil, fmt.Errorf("REQUIRES_CALCULATOR")
	}

	err = s.db.QueryRowContext(ctx, `
		SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position
		FROM dem_product_instances_al 
		WHERE order_num = ? AND type = 'vodootliv' AND part_type = 'main'
		ORDER BY id DESC LIMIT 1
	`, orderNum).Scan(&item.ID, &item.OrderNum, &item.Name, &item.Count, &item.TotalTime, &item.CreatedAT, &item.Type, &item.PartType, &item.ParentProductID, &item.ParentAssembly, &item.Status, &item.Position)

	if err == nil {
		ops, _ := s.loadOperationsForProduct(ctx, item.ID)
		item.Operations = ops
		return &item, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("%s: search by order_num error: %w", op, err)
	}

	newItem, err := s.importGutterFromLegacy(ctx, requestedID, orderNum)
	if err != nil {
		// Если ошибка дубля — рекурсивно пробуем найти созданную запись
		if err.Error() == "DUPLICATE_ENTRY" {
			time.Sleep(100 * time.Millisecond)
			return s.GetGutterOrderDetails(ctx, requestedID)
		}
		return nil, err
	}

	ops, _ := s.loadOperationsForProduct(ctx, newItem.ID)
	newItem.Operations = ops
	return newItem, nil
}

func (s *Storage) importGutterFromLegacy(ctx context.Context, legacyID int64, orderNum string) (*storage.GetOrderDetails, error) {
	const op = "storage.mysql.importGutterFromLegacy"

	stmtOrder := `SELECT idorders, numorders, ordername, 'Водоотлив/Оцинковка' as name, 'vodootliv' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = 6`

	var detail storage.GetOrderDetails
	err := s.db.QueryRowContext(ctx, stmtOrder, legacyID).Scan(
		&detail.ID, &detail.OrderNum, &detail.Customer, &detail.Name, &detail.Type, &detail.PartType,
		&detail.Status, &detail.TemplateCode)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%s: gutter order %d not found", op, legacyID)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: query order: %w", op, err)
	}

	var nashchelnikCount int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM dem_param_vo 
		WHERE orderid = ? AND mat IN (3, 4, 7, 9)
	`, legacyID).Scan(&nashchelnikCount)

	if err != nil {
		return nil, fmt.Errorf("%s: check nashchelnik: %w", op, err)
	}

	if nashchelnikCount > 0 {
		return nil, fmt.Errorf("HAS_NASHCHELNIK")
	}

	stmtParamVo := `SELECT SUM(kol_vo) as kol_vo, SUM(kol_vo * h * b) as sqr, SUM(kol_vo * h) as pgm 
					FROM dem_param_vo 
					WHERE orderid = ? AND mat IN (1, 5, 8) AND mat NOT IN (10, 11)`

	var count, sqr, pgm float64
	err = s.db.QueryRowContext(ctx, stmtParamVo, legacyID).Scan(&count, &sqr, &pgm)
	if err != nil {
		slog.Warn("failed to get vo params", "op", op, "err", err)
	}

	val := sqr / 1000000.0
	allSqr := math.Round(val*1000) / 1000
	detail.Count = count
	detail.Sqr = allSqr

	stmtTypeVo := `SELECT mat, SUM(kol_vo) FROM dem_param_vo WHERE orderid = ? AND mat IN (1, 5, 8) GROUP BY mat`
	typeCounts := map[string]int{"vo": 0, "ocn": 0}
	tRows, err := s.db.QueryContext(ctx, stmtTypeVo, legacyID)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var mat, cnt int
			if err := tRows.Scan(&mat, &cnt); err == nil {
				if mat == 5 || mat == 8 {
					typeCounts["ocn"] += cnt
				} else if mat == 1 {
					typeCounts["vo"] += cnt
				}
			}
		}
	}

	vodootl := typeCounts["vo"]
	ocinkov := typeCounts["ocn"]
	var typeIzdVo string
	var name string
	switch {
	case vodootl > 0 && ocinkov > 0:
		name = "Водоотлив/Оцинковка"
		typeIzdVo = fmt.Sprintf("%dvo+%docn", vodootl, ocinkov)
	case ocinkov > 0:
		name = "Оцинковка"
		typeIzdVo = "ocn"
	default:
		name = "Водоотлив"
		typeIzdVo = "vo"
	}

	detail.Operations = make([]storage.NormOperation, 0)

	stmtOps := `
		SELECT 
			d.name_mat,
			SUM(d.allowances) as total_hours,
			SUM(d.kol_vo) as total_count
		FROM dem_orderdetails d
		LEFT JOIN dem_type_works t ON d.type_m_id = t.type_m_id
		WHERE d.orderid = ? 
		  AND t.type_code LIKE 'trud'
		GROUP BY d.name_mat
		ORDER BY FIELD(d.name_mat, 'Разметка', 'Резка', 'Гиб', 'Отбортовка', 'Упаковка')
	`

	opsRows, err := s.db.QueryContext(ctx, stmtOps, legacyID)
	if err != nil {
		return nil, fmt.Errorf("%s: query operations: %w", op, err)
	} else {
		defer opsRows.Close()

		var totalTimeVo float64

		for opsRows.Next() {
			var oper storage.NormOperation
			var label string
			var hours float64

			err = opsRows.Scan(&label, &hours, &oper.Count)
			if err != nil {
				continue
			}

			if hours == 0.00 {
				continue
			}

			oper.Label = label
			oper.Name = strings.ToLower(strings.ReplaceAll(label, " ", "_"))
			oper.Value = hours
			oper.Minutes = hours * 60.0

			detail.Operations = append(detail.Operations, oper)
			totalTimeVo += hours
		}

		detail.TotalTime = math.Round(totalTimeVo*1000) / 1000

		if len(detail.Operations) == 0 {
			return nil, fmt.Errorf("%s: no operations found", op)
		}
	}

	status := "in_production"
	newItem := &storage.GetOrderDetails{
		OrderNum:     detail.OrderNum,
		TemplateCode: detail.TemplateCode,
		Name:         name,
		Customer:     detail.Customer,
		Count:        detail.Count,
		Sqr:          detail.Sqr,
		TotalTime:    detail.TotalTime,
		Type:         "vodootliv",
		PartType:     "main",
		Status:       &status,
		Position:     0,
		TypeIzd:      typeIzdVo,
		Profile:      "",
		Systema:      "",
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var existingID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM dem_product_instances_al 
		WHERE order_num = ? AND type = 'vodootliv' LIMIT 1
	`, orderNum).Scan(&existingID)

	if err == nil {
		return &storage.GetOrderDetails{ID: existingID, OrderNum: orderNum}, nil
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO dem_product_instances_al (
			order_num, template_code, name, customer, count, total_time, 
			type, part_type, parent_assembly, status, position, sqr, type_izd, profile, systema
		) VALUES (?, '0', ?, ?, ?, ?, 'vodootliv', 'main', '', ?, 0, ?, ?, '', '')
	`, newItem.OrderNum, newItem.Name, newItem.Customer, newItem.Count, newItem.TotalTime, newItem.Status, newItem.Sqr, newItem.TypeIzd)

	if err != nil {
		// Обработка ошибки дубля (1062) от MySQL
		// Это случится, если проверка выше не сработала, но уникальный индекс сработал
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, fmt.Errorf("DUPLICATE_ENTRY")
		}
		return nil, fmt.Errorf("%s: insert failed: %w", op, err)
	}

	newID, _ := res.LastInsertId()

	for i, dop := range detail.Operations {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO dem_operation_values_al (product_id, operation_name, operation_label, count, value, minutes, sort_operation)
			VALUES (?, ?, ?, ?, ?, ?, ?)`, newID, dop.Name, dop.Label, dop.Count, dop.Value, dop.Minutes, i)

		if err != nil {
			return nil, fmt.Errorf("%s: insert operation %s: %w", op, dop.Name, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: commit failed: %w", op, err)
	}

	newItem.ID = newID
	return newItem, nil
}

func (s *Storage) SaveNashchelnikNorm(ctx context.Context, legacyID int64, orderNum string, a, b, c, d, sqr, count float64, operations []storage.NormOperation) (*storage.GetOrderDetails, error) {
	const op = "storage.mysql.SaveNashchelnikNorm"

	var customerName string
	err := s.db.QueryRowContext(ctx, `SELECT ordername FROM dem_orders WHERE idorders = ?`, legacyID).Scan(&customerName)
	if err != nil {
		customerName = ""
	}

	// TODO в будущем для обновления и предзаполнения данными
	//paramsMap := map[string]float64{"a": a, "b": b, "c": c, "d": d}
	//paramsJSON, _ := json.Marshal(paramsMap)
	//paramsStr := string(paramsJSON)

	var totalTime float64
	for _, opr := range operations {
		totalTime += opr.Value
	}
	totalTime = math.Round(totalTime*1000) / 1000

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("Ошибка транзакции %s, %w", op, err)
	}
	defer tx.Rollback()

	var existingID int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM dem_product_instances_al WHERE order_num = ? AND type = 'vodootliv' LIMIT 1 FOR UPDATE`, orderNum).Scan(&existingID)

	status := "in_production"
	var newItem storage.GetOrderDetails

	if existingID > 0 {
		_, err = tx.ExecContext(ctx,
			`UPDATE dem_product_instances_al SET total_time = ?, customer = ?, count = ?, sqr = ? WHERE id = ?`,
			totalTime,
			customerName,
			count,
			sqr,
			//paramsStr,
			existingID,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: update failed: %w", op, err)
		}

		_, err = tx.ExecContext(ctx, `DELETE FROM dem_operation_values_al WHERE product_id = ?`, existingID)
		if err != nil {
			return nil, fmt.Errorf("%s: удаление не удалось: %w", op, err)
		}

		newItem.ID = existingID
	} else {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO dem_product_instances_al (
				order_num, template_code, name, customer, count, total_time,
				type, part_type, parent_assembly, status, position, sqr, type_izd, profile
			) VALUES (
				?, '0', 'Водоотлив', ?, ?, ?,
				'vodootliv', 'main', '', ?, 0, ?, 'vo', ''
			)
		`,
			orderNum,
			customerName,
			count,
			totalTime,
			&status,
			sqr,
		)

		if err != nil {
			return nil, fmt.Errorf("%s: insert failed: %w", op, err)
		}
		id, _ := res.LastInsertId()
		newItem.ID = id
	}

	// Вставляем новые операции
	for i, dop := range operations {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO dem_operation_values_al (product_id, operation_name, operation_label, count, value, minutes, sort_operation)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, newItem.ID, dop.Name, dop.Label, dop.Count, dop.Value, dop.Minutes, i)
		if err != nil {
			// Если ошибка дубля, значит фронт прислал повторяющиеся operation_name для одного продукта
			return nil, fmt.Errorf("%s: duplicate op '%s' for product %d: %w", op, dop.Label, newItem.ID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	newItem.OrderNum = orderNum
	newItem.TotalTime = totalTime
	newItem.Operations = operations
	newItem.Type = "vodootliv"
	newItem.Count = count
	newItem.Sqr = sqr

	return &newItem, nil
}

func (s *Storage) GetNashchelnikRawData(ctx context.Context, legacyID int64) (*storage.NashchelnikRawData, error) {
	const op = "storage.mysql.GetNashchelnikRawData"

	var data storage.NashchelnikRawData
	data.LegacyID = legacyID

	err := s.db.QueryRowContext(ctx, `
		SELECT numorders, ordername 
		FROM dem_orders 
		WHERE idorders = ? AND class_id = 6
	`, legacyID).Scan(&data.OrderNum, &data.Customer)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%s: order not found", op)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: query header: %w", op, err)
	}

	var sqr, count, pgm float64
	stmtSqr := `SELECT SUM(kol_vo) as kol_vo, SUM(kol_vo * h * b) as sqr, SUM(kol_vo * CASE WHEN h > b THEN h ELSE b END) as pgm FROM dem_param_vo WHERE orderid = ? AND mat NOT IN (10,11)`
	err = s.db.QueryRowContext(ctx, stmtSqr, legacyID).Scan(&count, &sqr, &pgm)
	if err != nil {
		slog.Warn("failed to get sqr/pgm", "op", op, "err", err)
	}

	data.Sqr = math.Round((sqr/1000000.0)*1000) / 1000
	data.Pgm = pgm
	data.Count = count

	// 2. Получаем операции ИЗ ЛЕГАСИ, НО БЕЗ ГИБА И ОТБОРТОВКИ
	// Эти операции мы возьмем "как есть", а Гиб/Отбортовку посчитаем по новым формулам
	stmtOps := `
		SELECT 
			d.name_mat,
			SUM(d.allowances) as total_hours,
			SUM(d.kol_vo) as total_count
		FROM dem_orderdetails d
		LEFT JOIN dem_type_works t ON d.type_m_id = t.type_m_id
		WHERE d.orderid = ? 
		  AND t.type_code LIKE 'trud'
		  AND d.name_mat NOT IN ('Гиб', 'Отбортовка')
		GROUP BY d.name_mat
		ORDER BY FIELD(d.name_mat, 'Разметка', 'Резка', 'Упаковка')
	`

	rows, err := s.db.QueryContext(ctx, stmtOps, legacyID)
	if err != nil {
		slog.Warn("failed to query legacy ops", "op", op, "err", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var label string
			var hours, count float64

			if err := rows.Scan(&label, &hours, &count); err != nil {
				continue
			}

			data.ExistingOps = append(data.ExistingOps, storage.NormOperation{
				Label:   label,
				Name:    strings.ToLower(strings.ReplaceAll(label, " ", "_")),
				Value:   hours,
				Minutes: hours * 60,
				Count:   count,
			})
		}
	}

	return &data, nil
}
