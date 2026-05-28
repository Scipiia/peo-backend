package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"log/slog"
	"math"
	"strings"
	"time"
	"vue-golang/internal/storage"
)

// GetGutterOrderDetails универсальный вход для водоотливов/нащельников
func (s *Storage) GetGutterOrderDetails(ctx context.Context, requestedID int64) (*storage.GetOrderDetails, error) {
	const op = "storage.mysql.GetGutterOrderDetails"

	// 🔥 ШАГ 1: Пробуем найти в НОВОЙ базе по ID (это может быть уже импортированный заказ)
	var item storage.GetOrderDetails
	err := s.db.QueryRowContext(ctx, `
		SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position
		FROM dem_product_instances_al 
		WHERE id = ? AND type = 'vodootliv' AND part_type = 'main'
	`, requestedID).Scan(&item.ID, &item.OrderNum, &item.Name, &item.Count, &item.TotalTime, &item.CreatedAT, &item.Type, &item.PartType, &item.ParentProductID, &item.ParentAssembly, &item.Status, &item.Position)

	if err == nil {
		// ✅ Нашли по новому ID — загружаем операции и возвращаем
		ops, _ := s.loadOperationsForProduct(ctx, item.ID)
		item.Operations = ops
		return &item, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("%s: search by new ID error: %w", op, err)
	}

	// 🔥 ШАГ 2: Не нашли по новому ID → значит, requestedID — это legacy idorders
	// Получаем номер заказа из старой базы
	var orderNum string
	err = s.db.QueryRowContext(ctx, `SELECT numorders FROM dem_orders WHERE idorders = ? AND class_id = 6`, requestedID).Scan(&orderNum)
	if err != nil {
		// Если и здесь не нашли — возможно, заказ уже импортирован и у него новый ID
		// Попробуем найти по order_num в новой базе (на всякий случай)
		// Но сначала вернем понятную ошибку
		return nil, fmt.Errorf("%s: legacy order not found (id=%d): %w", op, requestedID, err)
	}

	// 🔥 ШАГ 3: Проверяем наличие нащельников (блокируем импорт, если они есть)
	var nashchelnikCount int
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dem_param_vo WHERE orderid = ? AND mat IN (3,4,7,9)`, requestedID).Scan(&nashchelnikCount)
	if nashchelnikCount > 0 {
		return nil, fmt.Errorf("REQUIRES_CALCULATOR")
	}

	// 🔥 ШАГ 4: Ищем запись в новой базе по order_num (вдруг уже импортирован, но по другому ID)
	err = s.db.QueryRowContext(ctx, `
		SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position
		FROM dem_product_instances_al 
		WHERE order_num = ? AND type = 'vodootliv' AND part_type = 'main'
		ORDER BY id DESC LIMIT 1
	`, orderNum).Scan(&item.ID, &item.OrderNum, &item.Name, &item.Count, &item.TotalTime, &item.CreatedAT, &item.Type, &item.PartType, &item.ParentProductID, &item.ParentAssembly, &item.Status, &item.Position)

	if err == nil {
		// ✅ Нашли по order_num — загружаем операции и возвращаем
		ops, _ := s.loadOperationsForProduct(ctx, item.ID)
		item.Operations = ops
		return &item, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("%s: search by order_num error: %w", op, err)
	}

	// 🔥 ШАГ 5: Не нашли нигде → запускаем импорт из легаси
	newItem, err := s.importGutterFromLegacy(ctx, requestedID, orderNum)
	if err != nil {
		// Если ошибка дубля — рекурсивно пробуем найти созданную запись
		if err.Error() == "DUPLICATE_ENTRY" {
			time.Sleep(100 * time.Millisecond)
			return s.GetGutterOrderDetails(ctx, requestedID)
		}
		return nil, err
	}

	// Загружаем операции для нового заказа
	ops, _ := s.loadOperationsForProduct(ctx, newItem.ID)
	newItem.Operations = ops
	return newItem, nil
}

func (s *Storage) importGutterFromLegacy(ctx context.Context, legacyID int64, orderNum string) (*storage.GetOrderDetails, error) {
	const op = "storage.mysql.importGutterFromLegacy"

	// 1. Читаем шапку из dem_orders
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

	// 2. ПРОВЕРКА НА НАЩЕЛЬНИКИ (еще раз для надежности)
	var nashchelnikCount int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM dem_param_vo 
		WHERE orderid = ? AND mat IN (3, 4, 7, 9)
	`, legacyID).Scan(&nashchelnikCount)

	if err != nil {
		return nil, fmt.Errorf("%s: check nashchelnik: %w", op, err)
	}

	// Если есть нащельники, возвращаем специальную ошибку
	// (хотя эта проверка уже была в GetGutterOrderDetails, но лучше перестраховаться)
	if nashchelnikCount > 0 {
		return nil, fmt.Errorf("HAS_NASHCHELNIK")
	}

	// 3. СЧИТЫВАЕМ ПАРАМЕТРЫ ВОДОOTЛИВОВ/ОЦИНКОВКИ
	// mat IN (1, 5, 8) - как ты указал ранее (1=Водоотлив, 5,8=Оцинковка)
	// mat NOT IN (10, 11) - исключаем мусор
	stmtParamVo := `SELECT SUM(kol_vo) as kol_vo, SUM(kol_vo * h * b) as sqr, SUM(kol_vo * h) as pgm 
					FROM dem_param_vo 
					WHERE orderid = ? AND mat IN (1, 5, 8) AND mat NOT IN (10, 11)`

	var count, sqr, pgm float64
	err = s.db.QueryRowContext(ctx, stmtParamVo, legacyID).Scan(&count, &sqr, &pgm)
	if err != nil {
		slog.Warn("failed to get vo params", "op", op, "err", err)
		// Не падаем, пусть будут нули, если таблица пуста
	}

	// Конвертируем площадь из мм² в м²
	val := sqr / 1000000.0
	allSqr := math.Round(val*1000) / 1000
	detail.Count = count
	detail.Sqr = allSqr

	// 4. ОПРЕДЕЛЯЕМ ТИП (ВО или ОЦ) для отображения
	// (можно оптимизировать, объединив с предыдущим запросом, но оставим как есть для наглядности)
	stmtTypeVo := `SELECT mat, SUM(kol_vo) FROM dem_param_vo WHERE orderid = ? AND mat IN (1, 5, 8) GROUP BY mat`
	typeCounts := map[string]int{"vo": 0, "ocn": 0}
	tRows, err := s.db.QueryContext(ctx, stmtTypeVo, legacyID)
	if err == nil {
		defer tRows.Close()
		for tRows.Next() {
			var mat, cnt int
			if err := tRows.Scan(&mat, &cnt); err == nil {
				if mat == 5 || mat == 8 { // Оцинковка
					typeCounts["ocn"] += cnt
				} else if mat == 1 { // Водоотлив
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

	// 5. ЗАГРУЖАЕМ ОПЕРАЦИИ ИЗ ЛЕГАСИ
	// Это твой код, который тянет операции из dem_orderdetails
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
		slog.Warn("failed to query ops", "op", op, "err", err)
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
		Type:         "mosquito",
		PartType:     "main",
		Status:       &status,
		Position:     0,
		TypeIzd:      typeIzdVo,
		Profile:      "",
		Systema:      "",
	}

	// 6. СОХРАНЕНИЕ (Транзакция)
	// 🔥 Начинаем транзакцию
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 🔥 ПРОВЕРКА НА ДУБЛЬ ВНУТРИ ТРАНЗАКЦИИ
	// Если кто-то параллельно уже создал запись, мы это увидим и не будем создавать вторую
	var existingID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM dem_product_instances_al 
		WHERE order_num = ? AND type = 'vodootliv' LIMIT 1
	`, orderNum).Scan(&existingID)

	if err == nil {
		// Запись уже есть! Возвращаем её, не создавая дубль.
		// Операции загрузит вызывающая функция (GetGutterOrderDetails)
		return &storage.GetOrderDetails{ID: existingID, OrderNum: orderNum}, nil
	}
	// Если err == sql.ErrNoRows — значит записи нет, продолжаем импорт.

	// Вставляем шапку заказа
	// Добавляем profile и systema в INSERT, чтобы соответствовать структуре
	res, err := tx.ExecContext(ctx, `
		INSERT INTO dem_product_instances_al (
			order_num, template_code, name, customer, count, total_time, 
			type, part_type, parent_assembly, status, position, sqr, type_izd, profile, systema
		) VALUES (?, '0', ?, ?, ?, ?, 'vodootliv', 'main', '', ?, 0, ?, ?, '', '')
	`, newItem.OrderNum, newItem.Name, newItem.Customer, newItem.Count, newItem.TotalTime, newItem.Status, newItem.Sqr, newItem.TypeIzd)

	if err != nil {
		// 🔥 Обработка ошибки дубля (1062) от MySQL
		// Это случится, если проверка выше не сработала, но уникальный индекс сработал
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, fmt.Errorf("DUPLICATE_ENTRY")
		}
		return nil, fmt.Errorf("%s: insert failed: %w", op, err)
	}

	newID, _ := res.LastInsertId()

	// Вставляем операции
	for i, dop := range detail.Operations {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO dem_operation_values_al (product_id, operation_name, operation_label, count, value, minutes, sort_operation)
			VALUES (?, ?, ?, ?, ?, ?, ?)`, newID, dop.Name, dop.Label, dop.Count, dop.Value, dop.Minutes, i)

		if err != nil {
			slog.Warn("failed to insert op", "op", op, "err", err)
			// Можно продолжить, если не критично, или вернуть ошибку
		}
	}

	// 🔥 Коммитим транзакцию
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: commit failed: %w", op, err)
	}

	// Возвращаем созданный элемент
	// Важно: вернем новый ID, а не legacyID
	newItem.ID = newID
	return newItem, nil
}

// TODO нащельники
func (s *Storage) SaveNashchelnikNorm(ctx context.Context, legacyID int64, orderNum string, a, b, c, d, sqr, count float64) (*storage.GetOrderDetails, error) {
	const op = "storage.mysql.SaveNashchelnikNorm"

	// 0. ПОЛУЧАЕМ ДАННЫЕ ЗАКАЗА (Заказчик и т.д.)
	var customerName string
	err := s.db.QueryRowContext(ctx, `SELECT ordername FROM dem_orders WHERE idorders = ?`, legacyID).Scan(&customerName)
	if err != nil {
		customerName = "" // Если не нашли, пусть будет пусто
	}

	// 1. РАСЧЕТ ВРЕМЕНИ ПО ФОРМУЛАМ ИЗ ТЗ
	timeGib := (a * (38.25 / 3600.0)) + (b * (42.5 / 3600.0))
	timeEdge := (c * ((38.25 * 1.5) / 3600.0)) + (d * ((42.5 * 1.5) / 3600.0))

	// 2. БАЗОВЫЕ ОПЕРАЦИИ (ИСКЛЮЧАЯ ГИБ И ОТБОРТОВКУ, ТАК КАК МЫ ИХ СЧИТАЕМ ЗАНОВО)
	var baseOps []storage.NormOperation
	stmtBaseOps := `
		SELECT d.name_mat, SUM(d.allowances) as hours, SUM(d.kol_vo) as count
		FROM dem_orderdetails d
		LEFT JOIN dem_type_works t ON d.type_m_id = t.type_m_id
		WHERE d.orderid = ? 
		  AND t.type_code LIKE 'trud'
		  AND d.name_mat NOT IN ('Гиб', 'Отбортовка')
		GROUP BY d.name_mat
	`
	rows, err := s.db.QueryContext(ctx, stmtBaseOps, legacyID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			var hours, count float64
			if err := rows.Scan(&name, &hours, &count); err == nil {
				baseOps = append(baseOps, storage.NormOperation{
					Name:    strings.ToLower(strings.ReplaceAll(name, " ", "_")),
					Label:   name,
					Value:   math.Round(hours*1000) / 1000,
					Minutes: math.Round((hours*60)*1000) / 1000,
					Count:   count,
				})
			}
		}
	}

	// 3. ФОРМИРУЕМ СПИСОК ВСЕХ ОПЕРАЦИЙ
	var allOps []storage.NormOperation
	allOps = append(allOps, baseOps...)

	// Добавляем рассчитанный Гиб
	if timeGib > 0 || (a+b+c+d) > 0 {
		allOps = append(allOps, storage.NormOperation{
			Name:    "gib_nashchelnika",
			Label:   "Гиб",
			Count:   a + b + c + d,
			Value:   math.Round(timeGib*1000) / 1000,
			Minutes: math.Round((timeGib*60)*1000) / 1000,
		})
	}

	// Добавляем рассчитанную Отбортовку
	if timeEdge > 0 || (c+d) > 0 {
		allOps = append(allOps, storage.NormOperation{
			Name:    "otbortovka_nashchelnika",
			Label:   "Отбортовка",
			Count:   c + d,
			Value:   math.Round(timeEdge*1000) / 1000,
			Minutes: math.Round((timeEdge*60)*1000) / 1000,
		})
	}

	// Считаем общее время
	var totalTime float64
	for _, op := range allOps {
		totalTime += op.Value
	}
	totalTime = math.Round(totalTime*1000) / 1000

	// 4. СОЗДАЕМ ИЛИ ОБНОВЛЯЕМ ИНСТАНС ЗАКАЗА
	var existingID int64
	err = s.db.QueryRowContext(ctx, `
		SELECT id FROM dem_product_instances_al 
		WHERE order_num = ? AND type = 'vodootliv' LIMIT 1
	`, orderNum).Scan(&existingID)

	status := "in_production"
	var newItem storage.GetOrderDetails

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if existingID > 0 {
		// Обновляем существующий
		_, err = tx.ExecContext(ctx,
			`UPDATE dem_product_instances_al SET total_time = ?, customer = ?, count = ?, sqr = ? WHERE id = ?`,
			totalTime,
			customerName,
			count,
			sqr,
			existingID,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: update failed: %w", op, err)
		}

		// Удаляем старые операции, чтобы не было дублей
		_, err = tx.ExecContext(ctx, `DELETE FROM dem_operation_values_al WHERE product_id = ?`, existingID)
		if err != nil {
			return nil, err
		}

		newItem.ID = existingID
	} else {
		// Создаем новый
		res, err := tx.ExecContext(ctx, `
			INSERT INTO dem_product_instances_al (
				order_num, template_code, name, customer, count, total_time, 
				type, part_type, parent_assembly, status, position, sqr, type_izd, systema, profile
			) VALUES (
				?, '0', 'Водоотлив', ?, ?, ?, 
				'vodootliv', 'main', '', ?, 0, ?, 'vo', '', ''
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
	for i, dop := range allOps {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO dem_operation_values_al (product_id, operation_name, operation_label, count, value, minutes, sort_operation)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, newItem.ID, dop.Name, dop.Label, dop.Count, dop.Value, dop.Minutes, i)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	// Заполняем ответ для фронта
	newItem.OrderNum = orderNum
	newItem.TotalTime = totalTime
	newItem.Operations = allOps
	newItem.Type = "vodootliv"
	newItem.Count = count
	newItem.Sqr = sqr

	fmt.Println(newItem)
	return &newItem, nil
}

// GetNashchelnikRawData загружает "сырые" данные для калькулятора
// Возвращает базовые операции из легаси, исключая пересчитываемые (Гиб, Отбортовка)
func (s *Storage) GetNashchelnikRawData(ctx context.Context, legacyID int64) (*storage.NashchelnikRawData, error) {
	const op = "storage.mysql.GetNashchelnikRawData"

	// 1. Получаем шапку заказа
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
	data.Pgm = math.Round(pgm / 1000.0)
	data.Count = count

	// 2. Получаем операции ИЗ ЛЕГАСИ, НО БЕЗ ГИБА И ОТБОРТОВКИ
	// Эти операции мы возьмем "как есть", а Гиб/Отбортовку посчитаем по новым формулам
	// Проверь имя таблицы: dem_type_works или dem_types?
	stmtOps := `
		SELECT 
			d.name_mat,
			SUM(d.allowances) as total_hours,
			SUM(d.kol_vo) as total_count
		FROM dem_orderdetails d
		LEFT JOIN dem_type_works t ON d.type_m_id = t.type_m_id
		WHERE d.orderid = ? 
		  AND t.type_code LIKE 'trud'
		  AND d.name_mat NOT IN ('Гиб', 'Отбортовка') -- 🔥 ИСКЛЮЧАЕМ ПЕРЕСЧИТЫВАЕМЫЕ
		GROUP BY d.name_mat
		ORDER BY FIELD(d.name_mat, 'Разметка', 'Резка', 'Упаковка')
	`

	rows, err := s.db.QueryContext(ctx, stmtOps, legacyID)
	if err != nil {
		slog.Warn("failed to query legacy ops", "op", op, "err", err)
		// Не падаем, просто вернем пустой список операций
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
