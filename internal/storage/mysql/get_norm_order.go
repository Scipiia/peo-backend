package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
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

func (s *Storage) GetNormOrdersByOrderNum(ctx context.Context, orderNum string, position int) ([]*storage.GetOrderDetails, error) {
	const op = "storage.mysql.GetNormOrdersByOrderNum"

	stmt := `
        SELECT id, name, count, total_time, created_at, updated_at, type, part_type, parent_assembly, parent_product_id, position
        FROM dem_product_instances_al
        WHERE order_num = ? AND position = ?
        ORDER BY CASE WHEN part_type = 'main' THEN 0 ELSE 1 END, id
    `

	rows, err := s.db.QueryContext(ctx, stmt, orderNum, position)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка выполнения запроса: %w", op, err)
	}
	defer rows.Close()

	var results []*storage.GetOrderDetails
	for rows.Next() {
		var detail storage.GetOrderDetails
		var parentAssembly sql.NullString

		if err := rows.Scan(&detail.ID, &detail.Name, &detail.Count, &detail.TotalTime,
			&detail.CreatedAT, &detail.UpdatedAT, &detail.Type, &detail.PartType,
			&parentAssembly, &detail.ParentProductID, &detail.Position); err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования: %w", op, err)
		}

		if parentAssembly.Valid {
			detail.ParentAssembly = parentAssembly.String
		}
		results = append(results, &detail)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка итерации: %w", op, err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%s: не найдено нарядов для order_num=%s и position=%d: %w", op, orderNum, position, sql.ErrNoRows)
	}

	return results, nil
}

func (s *Storage) GetNormOrders(ctx context.Context, orderNum, orderType string) ([]storage.GetOrderDetails, error) {
	const op = "storage.mysql.GetNormOrders"

	stmt := `SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position FROM dem_product_instances_al 
        	WHERE 1=1 AND (?='' OR order_num LIKE CONCAT('%', ?, '%')) AND (? = '' OR type = ?) AND part_type='main' ORDER BY created_at DESC LIMIT 25`

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
			&item.Position,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: сканирование: %w", op, err)
		}
		items = append(items, item)
	}

	if orderType == "" || orderType == "mosquito" {
		//COLLATE utf8mb4_0900_ai_ci
		stmtMoskit := `
	   SELECT idorders, numorders, 'Москитная сетка', ordername, 1, 0, FROM_UNIXTIME(date),
       'mosquito', 'main', 0, '', 'in_production', 0
		FROM dem_orders
		WHERE class_id = '4'
  		AND (?= '' OR numorders LIKE CONCAT('%', ?, '%'))
  		AND NOT EXISTS (
      	SELECT 1 FROM dem_product_instances_al 
      	WHERE dem_product_instances_al.order_num = dem_orders.numorders
        	AND dem_product_instances_al.type = 'mosquito'
        	AND dem_product_instances_al.part_type = 'main')
		ORDER BY date DESC
		LIMIT 25`

		rowsMoskit, err := s.db.QueryContext(ctx, stmtMoskit, orderNum, orderNum)

		if err != nil {
			// ⚠️ Не падаем, если легаси недоступен — просто вернём внутренние заказы
			slog.Warn("failed to fetch mosquito orders", "op", op, "err", err)
		} else {
			defer rowsMoskit.Close()

			for rowsMoskit.Next() {
				var item storage.GetOrderDetails
				if err = rowsMoskit.Scan(
					&item.ID, &item.OrderNum, &item.Name, &item.Customer, &item.Count, &item.TotalTime, &item.CreatedAT,
					&item.Type, &item.PartType, &item.ParentProductID, &item.ParentAssembly, &item.Status, &item.Position,
				); err != nil {
					slog.Warn("failed to scan mosquito order", "op", op, "err", err)
					continue // Пропускаем битую строку, не ломаем весь список
				}
				//item.Type = "mosquito" // ✅ Помечаем источник
				items = append(items, item)
			}
			if err = rowsMoskit.Err(); err != nil {
				slog.Warn("error iterating mosquito orders", "op", op, "err", err)
			}
		}
	}

	if orderType == "" || orderType == "vodootliv" {
		//COLLATE utf8mb4_0900_ai_ci
		stmtVodootliv := `
	  SELECT idorders, numorders, 'Водоотлив или оцинковка', ordername, 1, 0, FROM_UNIXTIME(date),
	  'vodootliv', 'main', 0, '', 'in_production', 0
		FROM dem_orders
		WHERE class_id = '6'
		AND (?= '' OR numorders LIKE CONCAT('%', ?, '%'))
		AND NOT EXISTS (
	 	SELECT 1 FROM dem_product_instances_al
	 	WHERE dem_product_instances_al.order_num = dem_orders.numorders
	   	AND dem_product_instances_al.type = 'vodootliv'
	   	AND dem_product_instances_al.part_type = 'main')
		ORDER BY date DESC
		LIMIT 25`

		rowsVodootliv, err := s.db.QueryContext(ctx, stmtVodootliv, orderNum, orderNum)
		if err != nil {
			// ⚠️ Не падаем, если легаси недоступен — просто вернём внутренние заказы
			slog.Warn("failed to fetch mosquito orders", "op", op, "err", err)
		} else {
			defer rowsVodootliv.Close()

			for rowsVodootliv.Next() {
				var item storage.GetOrderDetails
				if err = rowsVodootliv.Scan(
					&item.ID, &item.OrderNum, &item.Name, &item.Customer, &item.Count, &item.TotalTime, &item.CreatedAT,
					&item.Type, &item.PartType, &item.ParentProductID, &item.ParentAssembly, &item.Status, &item.Position,
				); err != nil {
					slog.Warn("failed to scan mosquito order", "op", op, "err", err)
					continue // Пропускаем битую строку, не ломаем весь список
				}
				//item.Type = "vodootliv" // ✅ Помечаем источник
				items = append(items, item)
			}
			if err = rowsVodootliv.Err(); err != nil {
				slog.Warn("error iterating vodootliv orders", "op", op, "err", err)
			}
		}
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
			defer execRows.Close()

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

//func (s *Storage) GetMosquitoOrderDetails(ctx context.Context, orderID int64) ([]*storage.GetOrderDetails, error) {
//	const op = "storage.mysql.GetMosquitoOrderDetails"
//	slog.Info("Requested ID", "id", orderID)
//
//	// 1. Сначала пробуем найти как внутренний ID (если фронт уже подменил)
//	var legacyID int64
//	var orderNum, orderName string
//	var orderDate time.Time
//
//	var detail storage.GetOrderDetails
//
//	// Проверяем, есть ли это ID в dem_product_instances_al как москитка
//	err := s.db.QueryRowContext(ctx, `
//        SELECT id, order_num, name, created_at
//        FROM dem_product_instances_al
//        WHERE id = ? AND type = 'mosquito'
//    `, orderID).Scan(&legacyID, &orderNum, &orderName, &orderDate)
//
//	if err == nil {
//		// ✅ УСПЕХ: Мы нашли якорь. Используем legacy_id для запроса к старой CRM.
//		slog.Debug("Found internal anchor", "internal_id", orderID, "legacy_id", legacyID)
//	} else if err == sql.ErrNoRows {
//		// ⚠️ НЕ НАШЛИ во внутренней базе. Значит, requestedID — это скорее всего Legacy ID.
//		// Проверяем, существует ли он в старой CRM напрямую.
//		slog.Debug("No internal anchor found, assuming Legacy ID")
//
//		var checkID int64
//		errCheck := s.db.QueryRowContext(ctx, `SELECT idorders FROM dem_orders WHERE idorders = ? AND ms = '1'`, orderID).Scan(&checkID)
//
//		if errCheck == nil {
//			// Это валидный Legacy ID
//			legacyID = orderID
//			// Нам все равно придется сходить за деталями (name, date) чуть ниже,
//			// но пока просто продолжаем.
//		} else {
//			// ❌ ПРОВАЛ: Нет ни во внутренней базе, ни в старой.
//			return nil, fmt.Errorf("%s: mosquito order %d not found anywhere", op, orderID)
//		}
//	} else {
//		return nil, fmt.Errorf("%s: db error checking anchor: %w", op, err)
//	}
//
//	// ============================================================
//	// 1. Базовая информация (из dem_orders)
//	// ============================================================
//	//stmtOrder := `
//	//	SELECT
//	//		idorders, numorders, 'Москитная сетка' as ordername, FROM_UNIXTIME(date),
//	//		'mosquito' as type, 'main' as part_type, 'in_production' as status,
//	//		1 as count, 0 as total_time, 0 as position,
//	//		NOW() as created_at, NOW() as updated_at,
//	//		'' as parent_assembly, 0 as parent_product_id,
//	//		'' as template_code, '' as head_name, '' as type_izd,
//	//		NULL as ready_date
//	//	FROM dem_orders
//	//	WHERE idorders = ? AND ms = '1'
//	//`
//	//
//	//var detail storage.GetOrderDetails
//	//var readyDate sql.NullTime
//	//
//	//err := s.db.QueryRowContext(ctx, stmtOrder, orderID).Scan(
//	//	&detail.ID, &detail.OrderNum, &detail.Name, &detail.CreatedAT,
//	//	&detail.Type, &detail.PartType, &detail.Status,
//	//	&detail.Count, &detail.TotalTime, &detail.Position,
//	//	&detail.CreatedAT, &detail.UpdatedAT,
//	//	&detail.ParentAssembly, &detail.ParentProductID,
//	//	&detail.TemplateCode, &detail.HeadName, &detail.TypeIzd,
//	//	&readyDate,
//	//)
//	//if err == sql.ErrNoRows {
//	//	return nil, fmt.Errorf("%s: mosquito order %d not found", op, orderID)
//	//}
//	//if err != nil {
//	//	return nil, fmt.Errorf("%s: query order: %w", op, err)
//	//}
//
//	stmtParamMs := `SELECT SUM(kol_vo) as kol_vo,SUM(kol_vo * weight * hight) sqr FROM dem_param_moskit WHERE orderid = ?`
//
//	var count, sqr float64
//	err = s.db.QueryRowContext(ctx, stmtParamMs, orderID).Scan(&count, &sqr)
//	if err != nil {
//		// Если таблица пуста или ошибка, не ломаем весь процесс, просто логируем
//		slog.Warn("failed to get mosquito params", "op", op, "err", err)
//		count = 0
//		sqr = 0
//	}
//
//	val := sqr / 1000000.0
//	allSqr := math.Round(val*1000) / 1000
//	detail.Count = count
//	detail.Sqr = allSqr
//
//	// ✅ КРИТИЧНО: инициализируем слайсы, чтобы в JSON улетело [], а не null
//	detail.Operations = make([]storage.NormOperation, 0)
//
//	// ============================================================
//	// 2. Операции из dem_orderdetails (type_m_id = 30 → trud)
//	// ============================================================
//	const TRUD_TYPE_ID = 30
//
//	stmtOps := `
//    SELECT
//        d.name_mat,
//        SUM(d.allowances) as total_value,
//        SUM(d.kol_vo) as total_count
//    FROM dem_orderdetails d
//    WHERE d.orderid = ?
//      AND d.type_m_id = ?                      -- Только операции (trud)
//    GROUP BY d.articul_mat, d.name_mat, d.messure
//    ORDER BY FIELD(d.name_mat,
//        'Напиловка',
//        'Сборка, опрессовка',
//        'Сборка',
//        'Скатка',
//        'Установка крепежа',
//        'Изготовление',
//        'Установка защиты (вилатерм)'
//    ), d.name_mat ASC
//`
//
//	opsRows, err := s.db.QueryContext(ctx, stmtOps, orderID, TRUD_TYPE_ID)
//	if err != nil {
//		return nil, fmt.Errorf("%s: query operations: %w", op, err)
//	}
//	defer opsRows.Close()
//
//	for opsRows.Next() {
//		var oper storage.NormOperation
//		var rawHours float64
//		var name string
//
//		// ⚠️ Порядок Scan должен точно совпадать с порядком SELECT выше!
//		err := opsRows.Scan(
//			&name,       // name_mat → operation_label
//			&rawHours,   // SUM(value) → value (суммарные часы)
//			&oper.Count, // SUM(kol_vo) → count (общее количество)
//		)
//		if err != nil {
//			return nil, fmt.Errorf("%s: scan operation: %w", op, err)
//		}
//
//		oper.Label = name
//		oper.Name = name
//		oper.Value = rawHours          // Норма в часах
//		oper.Minutes = rawHours * 60.0 // Норма в минутах
//
//		// ✅ Инициализируем слайс исполнителей
//		oper.AssignedWorkers = make([]storage.AssignedWorker, 0)
//
//		// 🔍 Загрузка исполнителей: теперь ищем по имени операции (после агрегации)
//		stmtExec := `
//        SELECT employee_id, actual_minutes, actual_value
//        FROM dem_operation_executors_al
//        WHERE product_id = ? AND operation_name = ?
//    `
//		execRows, err := s.db.QueryContext(ctx, stmtExec, orderID, oper.Label)
//		if err != nil {
//			slog.Warn("failed to query executors", "op", op, "operation_name", oper.Label, "err", err)
//		} else {
//			defer execRows.Close()
//			for execRows.Next() {
//				var ex storage.AssignedWorker
//				if err := execRows.Scan(&ex.EmployeeID, &ex.ActualMinutes, &ex.ActualValue); err != nil {
//					slog.Warn("failed to scan executor", "op", op, "err", err)
//					continue
//				}
//				oper.AssignedWorkers = append(oper.AssignedWorkers, ex)
//			}
//		}
//
//		// Заглушка для пустых исполнителей
//		if len(oper.AssignedWorkers) == 0 {
//			oper.AssignedWorkers = []storage.AssignedWorker{
//				{
//					EmployeeID:    0,
//					ActualMinutes: oper.Minutes,
//					ActualValue:   oper.Value,
//				},
//			}
//		}
//
//		detail.Operations = append(detail.Operations, oper)
//	}
//
//	var totalTime float64
//	for _, oper := range detail.Operations {
//		totalTime += oper.Value
//	}
//	detail.TotalTime = totalTime
//
//	if err := opsRows.Err(); err != nil {
//		return nil, fmt.Errorf("%s: iteration error: %w", op, err)
//	}
//
//	tx, err := s.db.BeginTx(ctx, nil)
//	if err != nil {
//		return nil, fmt.Errorf("%s: begin tx: %w", op, err)
//	}
//
//	// Получаем или создаем запись в dem_product_instances_al
//	internalID, err := s.GetOrCreateMosquitoInstance(ctx, tx, detail.OrderNum, detail.Name, totalTime, detail.Count, detail.Sqr)
//	if err != nil {
//		tx.Rollback()
//		return nil, fmt.Errorf("%s: get/create anchor: %w", op, err)
//	}
//
//	// Коммитим создание якоря сразу, чтобы он стал виден другим транзакциям
//	if err := tx.Commit(); err != nil {
//		return nil, fmt.Errorf("%s: commit anchor: %w", op, err)
//	}
//
//	detail.ID = internalID
//
//	fmt.Println(detail)
//
//	return []*storage.GetOrderDetails{&detail}, nil
//}
