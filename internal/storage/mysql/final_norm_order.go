package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"vue-golang/internal/storage"
)

func (s *Storage) GetSimpleOrderReport(ctx context.Context, orderNum string) (*storage.OrderFinalReport, error) {
	const op = "storage.mysql.GetSimpleOrderReport"

	query := `
		SELECT
			pi.id,
			pi.order_num,
			pi.name,
			t.name AS template_name,
			ov.operation_name,
			ov.operation_label,
			ov.minutes AS norm_minutes,
			ov.value AS norm_value,
			e.name AS employee_name,
			oe.actual_minutes,
			oe.actual_value
		FROM dem_product_instances_al pi
		JOIN dem_templates_al t ON pi.template_code = t.code
		JOIN dem_operation_values_al ov ON pi.id = ov.product_id
		LEFT JOIN dem_operation_executors_al oe ON ov.product_id = oe.product_id AND ov.operation_name = oe.operation_name
		LEFT JOIN dem_employees_al e ON oe.employee_id = e.id
		WHERE pi.order_num = ?
		ORDER BY pi.id, ov.operation_name;
	`

	rows, err := s.db.QueryContext(ctx, query, orderNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: заказы не найдены: %w", op, err)
		}
		return nil, fmt.Errorf("%s: ошибка выполнения запроса: %w", op, err)
	}
	defer rows.Close()

	report := &storage.OrderFinalReport{
		OrderNum: orderNum,
		Izdelie:  []storage.IzdelieInfo{},
	}

	productMap := make(map[int64]*storage.IzdelieInfo)

	for rows.Next() {
		var (
			productID      int64
			productName    string
			templateName   string
			operationName  string
			operationLabel string
			normMinutes    float64
			normValue      float64
			employeeName   sql.NullString
			actualMinutes  sql.NullFloat64
			actualValue    sql.NullFloat64
		)

		err := rows.Scan(
			&productID,
			&orderNum,
			&productName,
			&templateName,
			&operationName,
			&operationLabel,
			&normMinutes,
			&normValue,
			&employeeName,
			&actualMinutes,
			&actualValue,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования строки: %w", op, err)
		}

		izd, exists := productMap[productID]
		if !exists {
			izd = &storage.IzdelieInfo{
				ID:           productID,
				Name:         productName,
				TemplateName: templateName,
				Operations:   []storage.OperationsNorm{},
			}
			productMap[productID] = izd
		}

		var opNorm *storage.OperationsNorm
		for i := range izd.Operations {
			if izd.Operations[i].OperationName == operationName {
				opNorm = &izd.Operations[i]
				break
			}
		}

		if opNorm == nil {
			opNorm = &storage.OperationsNorm{
				OperationName:  operationName,
				OperationLabel: operationLabel,
				NormMinutes:    normMinutes,
				NormValue:      normValue,
				Executors:      []storage.Workers{},
			}
			izd.Operations = append(izd.Operations, *opNorm)
			opNorm = &izd.Operations[len(izd.Operations)-1]
		}

		if employeeName.Valid {
			worker := storage.Workers{
				WorkerName:    employeeName.String,
				ActualMinutes: actualMinutes.Float64,
				ActualValue:   actualValue.Float64,
			}
			opNorm.Executors = append(opNorm.Executors, worker)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка при чтении строк: %w", op, err)
	}

	for _, izd := range productMap {
		report.Izdelie = append(report.Izdelie, *izd)
	}

	return report, nil
}

//const (
//	StatusAssigned = "assigned"
//	StatusFinal    = "final"
//)

//
//// buildProductFilters возвращает WHERE-часть и аргументы для запроса
//func buildProductFilters(f ProductFilter) (string, []interface{}) {
//	var conditions []string
//	var args []interface{}
//
//	// 1. Статусы: показываем только назначенные и готовые
//	conditions = append(conditions, "p.status IN (?, ?)")
//	args = append(args, StatusAssigned, StatusFinal)
//
//	// 2. Диапазон дат (ready_date)
//	if !f.From.IsZero() {
//		conditions = append(conditions, "p.ready_date >= ?")
//		args = append(args, f.From.Format("2006-01-02"))
//	}
//	if !f.To.IsZero() {
//		// To включает весь день, поэтому берём начало следующего дня
//		conditions = append(conditions, "p.ready_date < ?")
//		args = append(args, f.To.AddDate(0, 0, 1).Format("2006-01-02"))
//	}
//
//	// 3. Поиск по номеру заказа
//	if f.OrderNum != "" {
//		conditions = append(conditions, "p.order_num LIKE ?")
//		args = append(args, "%"+f.OrderNum+"%")
//	}
//
//	// 4. Фильтр по типам изделий
//	var types []string
//	for _, t := range f.Type {
//		if t != "" {
//			types = append(types, t)
//		}
//	}
//	if len(types) > 0 {
//		placeholders := make([]string, len(types))
//		for i := range placeholders {
//			placeholders[i] = "?"
//		}
//		conditions = append(conditions, fmt.Sprintf("p.type IN (%s)", strings.Join(placeholders, ",")))
//		for _, t := range types {
//			args = append(args, t)
//		}
//	}
//
//	// Собираем итоговую строку WHERE
//	where := ""
//	if len(conditions) > 0 {
//		where = "WHERE " + strings.Join(conditions, " AND ")
//	}
//	return where, args
//}
//
//func (s *Storage) GetPEOProductsByCategory(ctx context.Context, filter ProductFilter) ([]storage.PEOProduct, []storage.GetWorkers, error) {
//	const op = "storage.mysql.GetPEOProductsByCategory"
//
//	// 1. Строим фильтр
//	whereClause, args := buildProductFilters(filter)
//
//	// 2. Запрос продуктов
//	// 👇 Порядок полей в SELECT должен ТОЧНО совпадать с порядком в Scan() ниже!
//	productQuery := fmt.Sprintf(`
//        SELECT
//            p.id, p.order_num, p.customer, p.total_time, p.created_at, p.status,
//            p.part_type, p.type, p.parent_product_id, p.parent_assembly,
//            COALESCE(c.short_name_customer, p.customer_type) AS customer_type,
//            p.systema, p.type_izd, p.profile, p.count, p.sqr, p.brigade,
//            p.norm_money, p.position, p.ready_date,
//            COALESCE(p.coefficient, dc.coefficient) AS coefficient
//        FROM dem_product_instances_al p
//        LEFT JOIN dem_customer_al c ON p.customer = c.name
//        LEFT JOIN dem_coefficient_al dc ON dc.type = p.type
//        %s
//        ORDER BY p.ready_date DESC, p.order_num
//    `, whereClause)
//
//	rows, err := s.db.QueryContext(ctx, productQuery, args...)
//	if err != nil {
//		return nil, nil, fmt.Errorf("%s: query products: %w", op, err)
//	}
//	defer rows.Close()
//
//	products := make(map[int64]*storage.PEOProduct)
//	var productIDs []int64
//
//	for rows.Next() {
//		var p storage.PEOProduct
//		var parentID sql.NullInt64
//		var readyDate sql.NullTime
//		var coef sql.NullFloat64 // 👇 Читаем в временную переменную!
//
//		// 👇 ВАЖНО: Порядок аргументов должен ТОЧНО совпадать с SELECT выше!
//		// Считай колонки в SELECT: 1.id, 2.order_num, ... 21.coefficient
//		err := rows.Scan(
//			&p.ID, &p.OrderNum, &p.Customer, &p.TotalTime, &p.CreatedAt, &p.Status, // 1-6
//			&p.PartType, &p.Type, &parentID, &p.ParentAssembly, // 7-10
//			&p.CustomerType, &p.Systema, &p.TypeIzd, &p.Profile, // 11-14
//			&p.Count, &p.Sqr, &p.Brigade, &p.NormMoney, &p.Position, // 15-19
//			&readyDate, &coef, // 20-21
//		)
//		if err != nil {
//			return nil, nil, fmt.Errorf("%s: scan product: %w", op, err)
//		}
//
//		// Обработка NULL
//		if parentID.Valid {
//			p.ParentProductID = &parentID.Int64
//		}
//		if readyDate.Valid {
//			t := readyDate.Time
//			p.ReadyDate = &t
//		}
//		if coef.Valid {
//			v := coef.Float64
//			p.Coefficient = &v
//		} else {
//			// Fallback: если коэффициента нет ни в продукте, ни в справочнике — ставим 1.0
//			defaultCoef := 1.0
//			p.Coefficient = &defaultCoef
//		}
//
//		p.EmployeeMinutes = make(map[int64]float64)
//		p.EmployeeValue = make(map[int64]float64)
//
//		products[p.ID] = &p
//		productIDs = append(productIDs, p.ID)
//	}
//
//	if len(products) == 0 {
//		return []storage.PEOProduct{}, []storage.GetWorkers{}, nil
//	}
//
//	// 3. Запрос сотрудников, которые работали в этих продуктах
//	empQuery := fmt.Sprintf(`
//        SELECT DISTINCT e.id, e.name
//        FROM dem_employees_al e
//        INNER JOIN dem_operation_executors_al oe ON e.id = oe.employee_id
//        WHERE e.is_active = TRUE
//        AND oe.product_id IN (%s)
//        ORDER BY e.name ASC
//    `, placeholders(len(productIDs)))
//
//	empArgs := toInterfaceSlice(productIDs)
//
//	rows, err = s.db.QueryContext(ctx, empQuery, empArgs...)
//	if err != nil {
//		return nil, nil, fmt.Errorf("%s: query employees: %w", op, err)
//	}
//	defer rows.Close()
//
//	var employees []storage.GetWorkers
//	var employeeIDs []int64
//
//	for rows.Next() {
//		var emp storage.GetWorkers
//		if err := rows.Scan(&emp.ID, &emp.Name); err != nil {
//			return nil, nil, fmt.Errorf("%s: scan employee: %w", op, err)
//		}
//		employees = append(employees, emp)
//		employeeIDs = append(employeeIDs, emp.ID)
//	}
//
//	// 👇 Если сотрудников нет — возвращаем продукты с пустыми полями исполнителей
//	// Это нормальная ситуация: заказ создан, но ещё никто не работал
//	if len(employees) == 0 {
//		return toProductList(products), []storage.GetWorkers{}, nil
//	}
//
//	// 4. Запрос фактических данных (минуты/значения)
//	// 👇 Защита от пустых списков через placeholders() который возвращает "NULL"
//	execQuery := fmt.Sprintf(`
//        SELECT product_id, employee_id, actual_minutes, actual_value
//        FROM dem_operation_executors_al
//        WHERE product_id IN (%s)
//        AND employee_id IN (%s)
//    `, placeholders(len(productIDs)), placeholders(len(employeeIDs)))
//
//	execArgs := make([]interface{}, 0, len(productIDs)+len(employeeIDs))
//	for _, id := range productIDs {
//		execArgs = append(execArgs, id)
//	}
//	for _, id := range employeeIDs {
//		execArgs = append(execArgs, id)
//	}
//
//	rows, err = s.db.QueryContext(ctx, execQuery, execArgs...)
//	if err != nil {
//		return nil, nil, fmt.Errorf("%s: query executors: %w", op, err)
//	}
//	defer rows.Close()
//
//	for rows.Next() {
//		var prodID, empID int64
//		var minutes, value float64
//		if err := rows.Scan(&prodID, &empID, &minutes, &value); err != nil {
//			return nil, nil, fmt.Errorf("%s: scan executor: %w", op, err)
//		}
//		if prod, ok := products[prodID]; ok {
//			prod.EmployeeMinutes[empID] += minutes
//			prod.EmployeeValue[empID] += value
//		}
//	}
//
//	return toProductList(products), employees, nil
//}
//
//// Вспомогательная функция: map[*PEOProduct] → []PEOProduct
//func toOrderedProductList(m map[int64]*storage.PEOProduct, ids []int64) []storage.PEOProduct {
//	res := make([]storage.PEOProduct, 0, len(ids))
//	for _, id := range ids {
//		if p, ok := m[id]; ok {
//			res = append(res, *p)
//		}
//	}
//	return res
//}
//
//// toInterfaceSlice конвертирует []int64 в []interface{} для аргументов SQL
////func toInterfaceSlice(ids []int64) []interface{} {
////	res := make([]interface{}, len(ids))
////	for i, id := range ids {
////		res[i] = id
////	}
////	return res
////}
////
////// toProductList конвертирует map в слайс
////func toProductList(m map[int64]*storage.PEOProduct) []storage.PEOProduct {
////	res := make([]storage.PEOProduct, 0, len(m))
////	for _, v := range m {
////		res = append(res, *v)
////	}
////	return res
////}
////
////func placeholders(n int) string {
////	if n <= 0 {
////		return ""
////	}
////	return strings.Repeat(",?", n)[1:] // результат: "?,?,?"
////}
//
//// TODO старая версия
//func (s *Storage) GetPEOProductsByCategory1(ctx context.Context, filter ProductFilter) ([]storage.PEOProduct, []storage.GetWorkers, error) {
//	const op = "storage.mysql.GetPEOProductsByCategory"
//
//	//TODO поправить для ПЭО
//	employees, err := s.GetAllWorkers(ctx, "")
//	if err != nil {
//		return nil, nil, fmt.Errorf("%s: %w", op, err)
//	}
//
//	if len(employees) == 0 {
//		return []storage.PEOProduct{}, []storage.GetWorkers{}, nil
//	}
//
//	// Получаем ID сотрудников
//	employeeIDs := make([]int64, len(employees))
//	for i, emp := range employees {
//		employeeIDs[i] = emp.ID
//	}
//
//	// Подготавливаем условия фильтрации
//	var conditions []string
//	var args []interface{}
//
//	// Статусы
//	conditions = append(conditions, "p.status IN (?, ?)")
//	args = append(args, StatusAssigned, StatusFinal)
//
//	if !filter.From.IsZero() {
//		conditions = append(conditions, "p.ready_date >= ?")
//		args = append(args, filter.From.Format("2006-01-02"))
//	}
//
//	if !filter.To.IsZero() {
//		nextDay := filter.To.AddDate(0, 0, 1)
//		conditions = append(conditions, "p.ready_date < ?")
//		args = append(args, nextDay.Format("2006-01-02"))
//	}
//
//	// Номер заказа
//	if filter.OrderNum != "" {
//		conditions = append(conditions, "p.order_num LIKE ?")
//		args = append(args, "%"+filter.OrderNum+"%")
//	}
//
//	// Типы: фильтруем пустые значения
//	var nonEmptyTypes []string
//	for _, t := range filter.Type {
//		if t != "" {
//			nonEmptyTypes = append(nonEmptyTypes, t)
//		}
//	}
//	if len(nonEmptyTypes) > 0 {
//		conditions = append(conditions, fmt.Sprintf("p.type IN (%s)", placeholders(len(nonEmptyTypes))))
//		for _, t := range nonEmptyTypes {
//			args = append(args, t)
//		}
//	}
//
//	// Формируем WHERE
//	whereClause := ""
//	if len(conditions) > 0 {
//		whereClause = "WHERE " + strings.Join(conditions, " AND ")
//	}
//
//	// Запрос изделий
//	queryProducts := `
//		SELECT
//			p.id, p.order_num, p.customer, p.total_time, p.created_at, p.status,
//			p.part_type, p.type, p.parent_product_id, p.parent_assembly,
//			COALESCE(c.short_name_customer, p.customer_type) AS customer_type,
//			p.systema, p.type_izd, p.profile, p.count, p.sqr, p.brigade, p.norm_money, p.position, p.ready_date, p.coefficient
//		FROM dem_product_instances_al p
//		LEFT JOIN dem_customer_al c ON p.customer = c.name
//		` + whereClause + `
//		ORDER BY p.ready_date DESC, p.order_num
//	`
//
//	rowsProducts, err := s.db.QueryContext(ctx, queryProducts, args...)
//	if err != nil {
//		return nil, nil, fmt.Errorf("%s: ошибка получения изделий в финальной функции отчета для ПЕО: %w", op, err)
//	}
//	defer rowsProducts.Close()
//
//	// Используем map указателей для последующего обновления
//	products := make(map[int64]*storage.PEOProduct)
//
//	for rowsProducts.Next() {
//		var (
//			id              int64
//			orderNum        string
//			customer        string
//			totalTime       float64
//			createdAt       time.Time
//			status          string
//			partType        string
//			Type            string
//			parentProductID sql.NullInt64
//			parentAssembly  string
//			customerType    string
//			systema         string
//			typeIzd         string
//			profile         string
//			count           int
//			sqr             float64
//			brigade         string
//			normMoney       float64
//			position        float64
//			readyDate       sql.NullTime
//			coefficient     sql.NullFloat64
//		)
//
//		err := rowsProducts.Scan(&id, &orderNum, &customer, &totalTime, &createdAt, &status, &partType, &Type, &parentProductID, &parentAssembly,
//			&customerType, &systema, &typeIzd, &profile, &count, &sqr, &brigade, &normMoney, &position, &readyDate, &coefficient)
//		if err != nil {
//			return nil, nil, fmt.Errorf("%s: scan product: %w", op, err)
//		}
//
//		// Обработка NULL-полей (оставляем пустыми — отображение на уровне UI)
//		// Примечание: если требуется "не определено", можно заменить здесь,
//		// но лучше делегировать это уровню представления.
//
//		var parentID *int64
//		if parentProductID.Valid {
//			parentID = &parentProductID.Int64
//		}
//
//		var readyDatePtr *time.Time
//		if readyDate.Valid {
//			t := readyDate.Time
//			readyDatePtr = &t
//		}
//
//		var coef *float64
//		if coefficient.Valid {
//			coef = &coefficient.Float64
//		} else {
//			stmtCoef := `SELECT coefficient FROM dem_coefficient_al WHERE type = ?`
//
//			err := s.db.QueryRowContext(ctx, stmtCoef, Type).Scan(&coef)
//			if err != nil {
//				if errors.Is(err, sql.ErrNoRows) {
//					return nil, nil, fmt.Errorf("%s: коэффифиент для type='%s' не найден: %w", op, Type, err)
//				}
//				return nil, nil, fmt.Errorf("%s: выполнение запроса для получения коэффициента завершилось ошибкой: %w", op, err)
//			}
//		}
//
//		p := &storage.PEOProduct{
//			ID:              id,
//			OrderNum:        orderNum,
//			Customer:        customer,
//			TotalTime:       totalTime,
//			CreatedAt:       createdAt,
//			Status:          status,
//			PartType:        partType,
//			Type:            Type,
//			ParentProductID: parentID,
//			ParentAssembly:  parentAssembly,
//			CustomerType:    customerType,
//			Systema:         systema,
//			TypeIzd:         typeIzd,
//			Profile:         profile,
//			Count:           count,
//			Sqr:             sqr,
//			Brigade:         brigade,
//			NormMoney:       normMoney,
//			Position:        position,
//			ReadyDate:       readyDatePtr,
//			Coefficient:     coef,
//			EmployeeMinutes: make(map[int64]float64),
//			EmployeeValue:   make(map[int64]float64),
//		}
//
//		products[p.ID] = p
//	}
//
//	// Если нет изделий — возвращаем рано
//	if len(products) == 0 {
//		return []storage.PEOProduct{}, employees, nil
//	}
//
//	// Собираем ID изделий
//	productIDs := make([]int64, 0, len(products))
//	for id := range products {
//		productIDs = append(productIDs, id)
//	}
//
//	// Запрос executors
//	queryExecutors := `
//		SELECT product_id, employee_id, actual_minutes, actual_value
//		FROM dem_operation_executors_al
//		WHERE product_id IN (` + placeholders(len(productIDs)) + `)
//		  AND employee_id IN (` + placeholders(len(employeeIDs)) + `)
//	`
//
//	execArgs := make([]interface{}, 0, len(productIDs)+len(employeeIDs))
//	for _, id := range productIDs {
//		execArgs = append(execArgs, id)
//	}
//	for _, id := range employeeIDs {
//		execArgs = append(execArgs, id)
//	}
//
//	rowsExecutors, err := s.db.QueryContext(ctx, queryExecutors, execArgs...)
//	if err != nil {
//		return nil, nil, fmt.Errorf("%s: ошибка получения исполнителей: %w", op, err)
//	}
//	defer rowsExecutors.Close()
//
//	// Агрегируем данные по сотрудникам
//	for rowsExecutors.Next() {
//		var productID, employeeID int64
//		var minutes, value float64
//		if err := rowsExecutors.Scan(&productID, &employeeID, &minutes, &value); err != nil {
//			return nil, nil, fmt.Errorf("%s: ошибка сканирования исполнителя: %w", op, err)
//		}
//		if p, ok := products[productID]; ok {
//			p.EmployeeMinutes[employeeID] += minutes
//			p.EmployeeValue[employeeID] += value
//		}
//	}
//
//	// Формируем итоговый слайс из указателей
//	productList := make([]storage.PEOProduct, 0, len(products))
//	for _, p := range products {
//		productList = append(productList, *p)
//	}
//
//	return productList, employees, nil
//}

// TODO рещение от гугл ИИ

type ProductFilter struct {
	From     time.Time
	To       time.Time
	OrderNum string
	Type     []string
}

// Константы статусов
const (
	StatusAssigned = "assigned"
	StatusFinal    = "final"
)

// --- Вспомогательные структуры и конструкторы запросов ---

// buildProductFilters формирует SQL условия и аргументы
func buildProductFilters(f ProductFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// Стандартные условия (всегда присутствуют)
	conditions = append(conditions, "p.status IN (?, ?)")
	args = append(args, StatusAssigned, StatusFinal)

	if !f.From.IsZero() {
		conditions = append(conditions, "p.ready_date >= ?")
		args = append(args, f.From.Format("2006-01-02"))
	}
	if !f.To.IsZero() {
		conditions = append(conditions, "p.ready_date < ?")
		args = append(args, f.To.AddDate(0, 0, 1).Format("2006-01-02"))
	}
	if f.OrderNum != "" {
		conditions = append(conditions, "p.order_num LIKE ?")
		args = append(args, "%"+f.OrderNum+"%")
	}

	// Фильтр по типам
	var validTypes []string
	for _, t := range f.Type {
		if t != "" {
			validTypes = append(validTypes, t)
		}
	}
	if len(validTypes) > 0 {
		conditions = append(conditions, fmt.Sprintf("p.type IN (%s)", placeholders(len(validTypes))))
		for _, t := range validTypes {
			args = append(args, t)
		}
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	return where, args
}

func placeholders(n int) string {
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}

func toInterfaceSlice(ids []int64) []interface{} {
	res := make([]interface{}, len(ids))
	for i, id := range ids {
		res[i] = id
	}
	return res
}

// --- Основные методы хранилища ---

func (s *Storage) GetPEOProductsByCategory(ctx context.Context, filter ProductFilter) ([]storage.PEOProduct, []storage.GetWorkers, error) {
	const op = "storage.mysql.GetPEOProductsByCategory"

	// 1. Загружаем основные данные продуктов
	productsMap, productIDs, err := s.fetchProducts(ctx, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	if len(productIDs) == 0 {
		return []storage.PEOProduct{}, []storage.GetWorkers{}, nil
	}

	// 2. Получаем список уникальных сотрудников для этих продуктов
	employees, err := s.fetchEmployeesByProducts(ctx, productIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	if len(employees) == 0 {
		return s.mapToOrderedSlice(productsMap, productIDs), []storage.GetWorkers{}, nil
	}

	// 3. Обогащаем продукты данными о затраченном времени
	if err := s.enrichWithExecutors(ctx, productsMap, productIDs, employees); err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	return s.mapToOrderedSlice(productsMap, productIDs), employees, nil
}

// fetchProducts загружает продукты и сохраняет порядок ID
func (s *Storage) fetchProducts(ctx context.Context, f ProductFilter) (map[int64]*storage.PEOProduct, []int64, error) {
	whereClause, args := buildProductFilters(f)

	query := fmt.Sprintf(`
		SELECT 
			p.id, p.order_num, p.customer, p.total_time, p.created_at, p.status,
			p.part_type, p.type, p.parent_product_id, p.parent_assembly,
			COALESCE(c.short_name_customer, p.customer_type) AS customer_type,
			p.systema, p.type_izd, p.profile, p.count, p.sqr, p.brigade, 
			p.norm_money, p.position, p.ready_date,
			COALESCE(p.coefficient, dc.coefficient) AS coefficient
		FROM dem_product_instances_al p
		LEFT JOIN dem_customer_al c ON p.customer = c.name
		LEFT JOIN dem_coefficient_al dc ON dc.type = p.type
		%s
		ORDER BY p.ready_date DESC, p.order_num`, whereClause)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	products := make(map[int64]*storage.PEOProduct)
	var order []int64

	for rows.Next() {
		var p storage.PEOProduct
		var parentID sql.NullInt64
		var readyDate sql.NullTime
		var coef sql.NullFloat64

		err := rows.Scan(
			&p.ID, &p.OrderNum, &p.Customer, &p.TotalTime, &p.CreatedAt, &p.Status,
			&p.PartType, &p.Type, &parentID, &p.ParentAssembly,
			&p.CustomerType, &p.Systema, &p.TypeIzd, &p.Profile,
			&p.Count, &p.Sqr, &p.Brigade, &p.NormMoney, &p.Position,
			&readyDate, &coef,
		)
		if err != nil {
			return nil, nil, err
		}

		// Маппинг Null-типов
		if parentID.Valid {
			p.ParentProductID = &parentID.Int64
		}
		if readyDate.Valid {
			t := readyDate.Time
			p.ReadyDate = &t
		}
		if coef.Valid {
			v := coef.Float64
			p.Coefficient = &v
		}

		p.EmployeeMinutes = make(map[int64]float64)
		p.EmployeeValue = make(map[int64]float64)

		products[p.ID] = &p
		order = append(order, p.ID)
	}
	return products, order, nil
}

// fetchEmployeesByProducts получает список активных сотрудников для набора продуктов
func (s *Storage) fetchEmployeesByProducts(ctx context.Context, productIDs []int64) ([]storage.GetWorkers, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT e.id, e.name
		FROM dem_employees_al e
		INNER JOIN dem_operation_executors_al oe ON e.id = oe.employee_id
		WHERE e.is_active = TRUE AND oe.product_id IN (%s)
		ORDER BY e.name ASC`, placeholders(len(productIDs)))

	rows, err := s.db.QueryContext(ctx, query, toInterfaceSlice(productIDs)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []storage.GetWorkers
	for rows.Next() {
		var emp storage.GetWorkers
		if err := rows.Scan(&emp.ID, &emp.Name); err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}
	return employees, nil
}

// enrichWithExecutors загружает детальную статистику исполнителей в объекты продуктов
func (s *Storage) enrichWithExecutors(ctx context.Context, products map[int64]*storage.PEOProduct, prodIDs []int64, employees []storage.GetWorkers) error {
	empIDs := make([]int64, len(employees))
	for i, e := range employees {
		empIDs[i] = e.ID
	}

	query := fmt.Sprintf(`
		SELECT product_id, employee_id, actual_minutes, actual_value
		FROM dem_operation_executors_al
		WHERE product_id IN (%s) AND employee_id IN (%s)`,
		placeholders(len(prodIDs)), placeholders(len(empIDs)))

	args := append(toInterfaceSlice(prodIDs), toInterfaceSlice(empIDs)...)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var pID, eID int64
		var mins, val float64
		if err := rows.Scan(&pID, &eID, &mins, &val); err != nil {
			return err
		}
		if p, ok := products[pID]; ok {
			p.EmployeeMinutes[eID] += mins
			p.EmployeeValue[eID] += val
		}
	}
	return nil
}

// mapToOrderedSlice преобразует карту обратно в слайс, сохраняя порядок из БД
func (s *Storage) mapToOrderedSlice(m map[int64]*storage.PEOProduct, order []int64) []storage.PEOProduct {
	res := make([]storage.PEOProduct, 0, len(order))
	for _, id := range order {
		if p, ok := m[id]; ok {
			res = append(res, *p)
		}
	}
	return res
}
