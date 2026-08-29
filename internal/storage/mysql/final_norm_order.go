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
			COALESCE(p.coefficient, dc.coefficient) AS coefficient, p.name, p.sqr_stv
		FROM dem_product_instances_al p
		LEFT JOIN dem_customer_al c ON p.customer = c.name
		LEFT JOIN dem_coefficient_al dc ON dc.type = p.type
		%s
		ORDER BY p.ready_date DESC, p.id DESC`, whereClause)

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
		var sqrStv sql.NullFloat64

		err := rows.Scan(
			&p.ID, &p.OrderNum, &p.Customer, &p.TotalTime, &p.CreatedAt, &p.Status,
			&p.PartType, &p.Type, &parentID, &p.ParentAssembly,
			&p.CustomerType, &p.Systema, &p.TypeIzd, &p.Profile,
			&p.Count, &p.Sqr, &p.Brigade, &p.NormMoney, &p.Position,
			&readyDate, &coef, &p.Name, &sqrStv,
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
		if sqrStv.Valid {
			s := sqrStv.Float64
			p.SqrStv = &s
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
