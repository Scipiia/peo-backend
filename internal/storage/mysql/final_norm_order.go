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

const (
	StatusAssigned = "assigned"
	StatusFinal    = "final"
)

type ProductFilter struct {
	From     time.Time
	To       time.Time
	OrderNum string
	Type     []string
}

func (s *Storage) GetPEOProductsByCategory(ctx context.Context, filter ProductFilter) ([]storage.PEOProduct, []storage.GetWorkers, error) {
	const op = "storage.mysql.GetPEOProductsByCategory"

	employees, err := s.GetAllWorkers(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(employees) == 0 {
		return []storage.PEOProduct{}, []storage.GetWorkers{}, nil
	}

	// Получаем ID сотрудников
	employeeIDs := make([]int64, len(employees))
	for i, emp := range employees {
		employeeIDs[i] = emp.ID
	}

	// Подготавливаем условия фильтрации
	var conditions []string
	var args []interface{}

	// Статусы
	conditions = append(conditions, "p.status IN (?, ?)")
	args = append(args, StatusAssigned, StatusFinal)

	if !filter.From.IsZero() {
		conditions = append(conditions, "p.ready_date >= ?")
		args = append(args, filter.From.Format("2006-01-02"))
	}

	if !filter.To.IsZero() {
		nextDay := filter.To.AddDate(0, 0, 1)
		conditions = append(conditions, "p.ready_date < ?")
		args = append(args, nextDay.Format("2006-01-02"))
	}

	// Номер заказа
	if filter.OrderNum != "" {
		conditions = append(conditions, "p.order_num LIKE ?")
		args = append(args, "%"+filter.OrderNum+"%")
	}

	// Типы: фильтруем пустые значения
	var nonEmptyTypes []string
	for _, t := range filter.Type {
		if t != "" {
			nonEmptyTypes = append(nonEmptyTypes, t)
		}
	}
	if len(nonEmptyTypes) > 0 {
		conditions = append(conditions, fmt.Sprintf("p.type IN (%s)", placeholders(len(nonEmptyTypes))))
		for _, t := range nonEmptyTypes {
			args = append(args, t)
		}
	}

	// Формируем WHERE
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Запрос изделий
	queryProducts := `
		SELECT 
			p.id, p.order_num, p.customer, p.total_time, p.created_at, p.status,
			p.part_type, p.type, p.parent_product_id, p.parent_assembly,
			COALESCE(c.short_name_customer, p.customer_type) AS customer_type,
			p.systema, p.type_izd, p.profile, p.count, p.sqr, p.brigade, p.norm_money, p.position, p.ready_date, p.coefficient
		FROM dem_product_instances_al p
		LEFT JOIN dem_customer_al c ON p.customer = c.name
		` + whereClause + `
		ORDER BY p.ready_date DESC, p.order_num
	`

	rowsProducts, err := s.db.QueryContext(ctx, queryProducts, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: ошибка получения изделий в финальной функции отчета для ПЕО: %w", op, err)
	}
	defer rowsProducts.Close()

	// Используем map указателей для последующего обновления
	products := make(map[int64]*storage.PEOProduct)

	for rowsProducts.Next() {
		var (
			id              int64
			orderNum        string
			customer        string
			totalTime       float64
			createdAt       time.Time
			status          string
			partType        string
			Type            string
			parentProductID sql.NullInt64
			parentAssembly  string
			customerType    string
			systema         string
			typeIzd         string
			profile         string
			count           int
			sqr             float64
			brigade         string
			normMoney       float64
			position        float64
			readyDate       sql.NullTime
			coefficient     sql.NullFloat64
		)

		err := rowsProducts.Scan(&id, &orderNum, &customer, &totalTime, &createdAt, &status, &partType, &Type, &parentProductID, &parentAssembly,
			&customerType, &systema, &typeIzd, &profile, &count, &sqr, &brigade, &normMoney, &position, &readyDate, &coefficient)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: scan product: %w", op, err)
		}

		// Обработка NULL-полей (оставляем пустыми — отображение на уровне UI)
		// Примечание: если требуется "не определено", можно заменить здесь,
		// но лучше делегировать это уровню представления.

		var parentID *int64
		if parentProductID.Valid {
			parentID = &parentProductID.Int64
		}

		var readyDatePtr *time.Time
		if readyDate.Valid {
			t := readyDate.Time
			readyDatePtr = &t
		}

		var coef *float64
		if coefficient.Valid {
			coef = &coefficient.Float64
		} else {
			stmtCoef := `SELECT coefficient FROM dem_coefficient_al WHERE type = ?`

			err := s.db.QueryRowContext(ctx, stmtCoef, Type).Scan(&coef)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, nil, fmt.Errorf("%s: коэффифиент для type='%s' не найден: %w", op, Type, err)
				}
				return nil, nil, fmt.Errorf("%s: выполнение запроса для получения коэффициента завершилось ошибкой: %w", op, err)
			}
		}

		p := &storage.PEOProduct{
			ID:              id,
			OrderNum:        orderNum,
			Customer:        customer,
			TotalTime:       totalTime,
			CreatedAt:       createdAt,
			Status:          status,
			PartType:        partType,
			Type:            Type,
			ParentProductID: parentID,
			ParentAssembly:  parentAssembly,
			CustomerType:    customerType,
			Systema:         systema,
			TypeIzd:         typeIzd,
			Profile:         profile,
			Count:           count,
			Sqr:             sqr,
			Brigade:         brigade,
			NormMoney:       normMoney,
			Position:        position,
			ReadyDate:       readyDatePtr,
			Coefficient:     coef,
			EmployeeMinutes: make(map[int64]float64),
			EmployeeValue:   make(map[int64]float64),
		}

		products[p.ID] = p
	}

	// Если нет изделий — возвращаем рано
	if len(products) == 0 {
		return []storage.PEOProduct{}, employees, nil
	}

	// Собираем ID изделий
	productIDs := make([]int64, 0, len(products))
	for id := range products {
		productIDs = append(productIDs, id)
	}

	// Запрос executors
	queryExecutors := `
		SELECT product_id, employee_id, actual_minutes, actual_value
		FROM dem_operation_executors_al
		WHERE product_id IN (` + placeholders(len(productIDs)) + `)
		  AND employee_id IN (` + placeholders(len(employeeIDs)) + `)
	`

	execArgs := make([]interface{}, 0, len(productIDs)+len(employeeIDs))
	for _, id := range productIDs {
		execArgs = append(execArgs, id)
	}
	for _, id := range employeeIDs {
		execArgs = append(execArgs, id)
	}

	rowsExecutors, err := s.db.QueryContext(ctx, queryExecutors, execArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: ошибка получения исполнителей: %w", op, err)
	}
	defer rowsExecutors.Close()

	// Агрегируем данные по сотрудникам
	for rowsExecutors.Next() {
		var productID, employeeID int64
		var minutes, value float64
		if err := rowsExecutors.Scan(&productID, &employeeID, &minutes, &value); err != nil {
			return nil, nil, fmt.Errorf("%s: ошибка сканирования исполнителя: %w", op, err)
		}
		if p, ok := products[productID]; ok {
			p.EmployeeMinutes[employeeID] += minutes
			p.EmployeeValue[employeeID] += value
		}
	}

	// Формируем итоговый слайс из указателей
	productList := make([]storage.PEOProduct, 0, len(products))
	for _, p := range products {
		productList = append(productList, *p)
	}

	return productList, employees, nil
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	items := make([]string, n)
	for i := range items {
		items[i] = "?"
	}
	return strings.Join(items, ",")
}
