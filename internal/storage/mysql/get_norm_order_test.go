package mysql

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetSimpleOrderReport_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var orderNum = "123"

	mock.ExpectQuery(regexp.QuoteMeta(`
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
		ORDER BY pi.id, ov.operation_name`)).
		WithArgs(orderNum).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_num", "name", "template_name", "operation_name", "operation_label",
			"norm_minutes", "norm_value", "employee_name", "actual_minutes", "actual_value"}).
			AddRow(1, "123", "nameWindow", "template_name1", "operationCut", "Резка", 6, 0.1, "Ivan", 6, 0.1).
			AddRow(2, "123", "nameWindow", "template_name1", "operationObr", "Обработка", 12, 0.2, "Ivan", 12, 0.2).
			AddRow(3, "123", "nameWindow", "template_name1", "operationYpac", "Упаковка", 24, 0.4, "Petr", 24, 0.4))

	result, err := stor.GetSimpleOrderReport(context.Background(), orderNum)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Len(t, result.Izdelie, 3)
	require.Equal(t, "123", result.OrderNum)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetSimpleOrderReport_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var orderNum = "123"

	mock.ExpectQuery(regexp.QuoteMeta(`
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
		ORDER BY pi.id, ov.operation_name`)).
		WithArgs(orderNum).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_num", "name", "template_name", "operation_name", "operation_label",
			"norm_minutes", "norm_value", "employee_name", "actual_minutes", "actual_value"})).
		WillReturnError(errors.New("ошибка выполнения запроса"))

	result, err := stor.GetSimpleOrderReport(context.Background(), orderNum)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "ошибка выполнения запроса")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetSimpleOrderReport_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var orderNum = "123"

	mock.ExpectQuery(regexp.QuoteMeta(`
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
		ORDER BY pi.id, ov.operation_name`)).
		WithArgs(orderNum).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_num", "name", "template_name", "operation_name", "operation_label",
			"norm_minutes", "norm_value", "employee_name", "actual_minutes", "actual_value"}).
			AddRow("abc", "123", "nameWindow", "template_name1", "operationCut", "Резка", 6, 0.1, "Ivan", 6, 0.1))

	result, err := stor.GetSimpleOrderReport(context.Background(), orderNum)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "ошибка сканирования строки")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPEOProductsByCategory_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}
	now := time.Now()

	filter := ProductFilter{}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
			p.id, p.order_num, p.customer, p.total_time, p.created_at, p.status,
			p.part_type, p.type, p.parent_product_id, p.parent_assembly,
			COALESCE(c.short_name_customer, p.customer_type) AS customer_type,
			p.systema, p.type_izd, p.profile, p.count, p.sqr, p.brigade, 
			p.norm_money, p.position, p.ready_date,
			COALESCE(p.coefficient, dc.coefficient) AS coefficient, p.name, p.sqr_stv
		FROM dem_product_instances_al p
		LEFT JOIN dem_customer_al c ON p.customer = c.name
		LEFT JOIN dem_coefficient_al dc ON dc.type = p.type
		WHERE p.status IN (?, ?)
		ORDER BY p.ready_date DESC, p.id DESC`)).
		WithArgs("assigned", "final").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "order_num", "customer", "total_time", "created_at", "status",
			"part_type", "type", "parent_product_id", "parent_assembly",
			"customer_type", "systema", "type_izd", "profile", "count", "sqr",
			"brigade", "norm_money", "position", "ready_date",
			"coefficient", "name", "sqr_stv",
		}).AddRow(
			1, "123", "Иванов", 120.0, now, "assigned",
			"основное", "window", nil, "", "Частное лицо",
			"KBE", "Окно", "70мм", 1, 2.5, "Бригада 1",
			1500.0, 1, now, 1.0, "Окно двухстворчатое", 1.2,
		))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT e.id, e.name
		FROM dem_employees_al e
		INNER JOIN dem_operation_executors_al oe ON e.id = oe.employee_id
		WHERE e.is_active = TRUE AND oe.product_id IN (?)
		ORDER BY e.name ASC`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Ivan"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT product_id, employee_id, actual_minutes, actual_value
		FROM dem_operation_executors_al
		WHERE product_id IN (?) AND employee_id IN (?)`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"product_id", "employee_id", "actual_minutes", "actual_value"}).
			AddRow(1, 1, 120, 0.8))

	products, workers, err := stor.GetPEOProductsByCategory(context.Background(), filter)
	require.NoError(t, err)
	require.Len(t, products, 1)
	require.Len(t, workers, 1)

	require.Equal(t, int64(1), products[0].ID)
	require.Equal(t, "123", products[0].OrderNum)

	require.Equal(t, int64(1), workers[0].ID)
	require.Equal(t, "Ivan", workers[0].Name)

	require.Equal(t, 120.0, products[0].EmployeeMinutes[1])
	require.Equal(t, 0.8, products[0].EmployeeValue[1])

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPEOProductsByCategory_EmptyProducts(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	filter := ProductFilter{}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
			p.id, p.order_num, p.customer, p.total_time, p.created_at, p.status,
			p.part_type, p.type, p.parent_product_id, p.parent_assembly,
			COALESCE(c.short_name_customer, p.customer_type) AS customer_type,
			p.systema, p.type_izd, p.profile, p.count, p.sqr, p.brigade, 
			p.norm_money, p.position, p.ready_date,
			COALESCE(p.coefficient, dc.coefficient) AS coefficient, p.name, p.sqr_stv
		FROM dem_product_instances_al p
		LEFT JOIN dem_customer_al c ON p.customer = c.name
		LEFT JOIN dem_coefficient_al dc ON dc.type = p.type
		WHERE p.status IN (?, ?)
		ORDER BY p.ready_date DESC, p.id DESC`)).
		WithArgs("assigned", "final").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "order_num", "customer", "total_time", "created_at", "status",
			"part_type", "type", "parent_product_id", "parent_assembly",
			"customer_type", "systema", "type_izd", "profile", "count", "sqr",
			"brigade", "norm_money", "position", "ready_date",
			"coefficient", "name", "sqr_stv",
		}))

	products, workers, err := stor.GetPEOProductsByCategory(context.Background(), filter)
	require.NoError(t, err)
	require.Empty(t, products)
	require.Empty(t, workers)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPEOProductsByCategory_ProductsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	filter := ProductFilter{}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
			p.id, p.order_num, p.customer, p.total_time, p.created_at, p.status,
			p.part_type, p.type, p.parent_product_id, p.parent_assembly,
			COALESCE(c.short_name_customer, p.customer_type) AS customer_type,
			p.systema, p.type_izd, p.profile, p.count, p.sqr, p.brigade, 
			p.norm_money, p.position, p.ready_date,
			COALESCE(p.coefficient, dc.coefficient) AS coefficient, p.name, p.sqr_stv
		FROM dem_product_instances_al p
		LEFT JOIN dem_customer_al c ON p.customer = c.name
		LEFT JOIN dem_coefficient_al dc ON dc.type = p.type
		WHERE p.status IN (?, ?)
		ORDER BY p.ready_date DESC, p.id DESC`)).
		WithArgs("assigned", "final").
		WillReturnError(errors.New("database error"))

	products, workers, err := stor.GetPEOProductsByCategory(context.Background(), filter)
	require.Error(t, err)
	require.Nil(t, products)
	require.Nil(t, workers)
	require.Contains(t, err.Error(), "database error")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPEOProductsByCategory_EmptyWorkers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}
	now := time.Now()

	filter := ProductFilter{}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
			p.id, p.order_num, p.customer, p.total_time, p.created_at, p.status,
			p.part_type, p.type, p.parent_product_id, p.parent_assembly,
			COALESCE(c.short_name_customer, p.customer_type) AS customer_type,
			p.systema, p.type_izd, p.profile, p.count, p.sqr, p.brigade, 
			p.norm_money, p.position, p.ready_date,
			COALESCE(p.coefficient, dc.coefficient) AS coefficient, p.name, p.sqr_stv
		FROM dem_product_instances_al p
		LEFT JOIN dem_customer_al c ON p.customer = c.name
		LEFT JOIN dem_coefficient_al dc ON dc.type = p.type
		WHERE p.status IN (?, ?)
		ORDER BY p.ready_date DESC, p.id DESC`)).
		WithArgs("assigned", "final").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "order_num", "customer", "total_time", "created_at", "status",
			"part_type", "type", "parent_product_id", "parent_assembly",
			"customer_type", "systema", "type_izd", "profile", "count", "sqr",
			"brigade", "norm_money", "position", "ready_date",
			"coefficient", "name", "sqr_stv",
		}).AddRow(
			1, "123", "Иванов", 120.0, now, "assigned",
			"основное", "window", nil, "", "Частное лицо",
			"KBE", "Окно", "70мм", 1, 2.5, "Бригада 1",
			1500.0, 1, now, 1.0, "Окно двухстворчатое", 1.2,
		))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT e.id, e.name
		FROM dem_employees_al e
		INNER JOIN dem_operation_executors_al oe ON e.id = oe.employee_id
		WHERE e.is_active = TRUE AND oe.product_id IN (?)
		ORDER BY e.name ASC`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	products, workers, err := stor.GetPEOProductsByCategory(context.Background(), filter)
	require.NoError(t, err)
	require.Len(t, products, 1)
	require.Empty(t, workers)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPEOProductsByCategory_ErrorWorkers(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}
	now := time.Now()

	filter := ProductFilter{}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
			p.id, p.order_num, p.customer, p.total_time, p.created_at, p.status,
			p.part_type, p.type, p.parent_product_id, p.parent_assembly,
			COALESCE(c.short_name_customer, p.customer_type) AS customer_type,
			p.systema, p.type_izd, p.profile, p.count, p.sqr, p.brigade, 
			p.norm_money, p.position, p.ready_date,
			COALESCE(p.coefficient, dc.coefficient) AS coefficient, p.name, p.sqr_stv
		FROM dem_product_instances_al p
		LEFT JOIN dem_customer_al c ON p.customer = c.name
		LEFT JOIN dem_coefficient_al dc ON dc.type = p.type
		WHERE p.status IN (?, ?)
		ORDER BY p.ready_date DESC, p.id DESC`)).
		WithArgs("assigned", "final").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "order_num", "customer", "total_time", "created_at", "status",
			"part_type", "type", "parent_product_id", "parent_assembly",
			"customer_type", "systema", "type_izd", "profile", "count", "sqr",
			"brigade", "norm_money", "position", "ready_date",
			"coefficient", "name", "sqr_stv",
		}).AddRow(
			1, "123", "Иванов", 120.0, now, "assigned",
			"основное", "window", nil, "", "Частное лицо",
			"KBE", "Окно", "70мм", 1, 2.5, "Бригада 1",
			1500.0, 1, now, 1.0, "Окно двухстворчатое", 1.2,
		))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT e.id, e.name
		FROM dem_employees_al e
		INNER JOIN dem_operation_executors_al oe ON e.id = oe.employee_id
		WHERE e.is_active = TRUE AND oe.product_id IN (?)
		ORDER BY e.name ASC`)).
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	products, workers, err := stor.GetPEOProductsByCategory(context.Background(), filter)
	require.Error(t, err)
	require.Nil(t, products)
	require.Nil(t, workers)
	require.Contains(t, err.Error(), "database error")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPEOProductsByCategory_EnrichExecutorsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}
	now := time.Now()

	filter := ProductFilter{}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
			p.id, p.order_num, p.customer, p.total_time, p.created_at, p.status,
			p.part_type, p.type, p.parent_product_id, p.parent_assembly,
			COALESCE(c.short_name_customer, p.customer_type) AS customer_type,
			p.systema, p.type_izd, p.profile, p.count, p.sqr, p.brigade, 
			p.norm_money, p.position, p.ready_date,
			COALESCE(p.coefficient, dc.coefficient) AS coefficient, p.name, p.sqr_stv
		FROM dem_product_instances_al p
		LEFT JOIN dem_customer_al c ON p.customer = c.name
		LEFT JOIN dem_coefficient_al dc ON dc.type = p.type
		WHERE p.status IN (?, ?)
		ORDER BY p.ready_date DESC, p.id DESC`)).
		WithArgs("assigned", "final").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "order_num", "customer", "total_time", "created_at", "status",
			"part_type", "type", "parent_product_id", "parent_assembly",
			"customer_type", "systema", "type_izd", "profile", "count", "sqr",
			"brigade", "norm_money", "position", "ready_date",
			"coefficient", "name", "sqr_stv",
		}).AddRow(
			1, "123", "Иванов", 120.0, now, "assigned",
			"основное", "window", nil, "", "Частное лицо",
			"KBE", "Окно", "70мм", 1, 2.5, "Бригада 1",
			1500.0, 1, now, 1.0, "Окно двухстворчатое", 1.2,
		))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT e.id, e.name
		FROM dem_employees_al e
		INNER JOIN dem_operation_executors_al oe ON e.id = oe.employee_id
		WHERE e.is_active = TRUE AND oe.product_id IN (?)
		ORDER BY e.name ASC`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Ivan"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT product_id, employee_id, actual_minutes, actual_value
		FROM dem_operation_executors_al
		WHERE product_id IN (?) AND employee_id IN (?)`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("database error"))

	products, workers, err := stor.GetPEOProductsByCategory(context.Background(), filter)
	require.Error(t, err)
	require.Nil(t, products)
	require.Nil(t, workers)
	require.Contains(t, err.Error(), "database error")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBuildProductFilters_Default(t *testing.T) {
	filter := ProductFilter{}

	where, args := buildProductFilters(filter)

	require.Equal(t, "WHERE p.status IN (?, ?)", where)
	require.Equal(t, []interface{}{"assigned", "final"}, args)
}

func TestBuildProductFilters_OrderNum(t *testing.T) {
	filter := ProductFilter{
		OrderNum: "123",
	}

	where, args := buildProductFilters(filter)

	require.Equal(t,
		"WHERE p.status IN (?, ?) AND p.order_num LIKE ?",
		where,
	)

	require.Equal(t,
		[]interface{}{"assigned", "final", "%123%"},
		args,
	)
}

func TestBuildProductFilters_Type(t *testing.T) {
	filter := ProductFilter{
		Type: []string{"window", "door"},
	}

	where, args := buildProductFilters(filter)

	require.Equal(t,
		"WHERE p.status IN (?, ?) AND p.type IN (?,?)",
		where,
	)

	require.Equal(t,
		[]interface{}{"assigned", "final", "window", "door"},
		args,
	)
}

func TestBuildProductFilters_Date(t *testing.T) {
	from := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)

	filter := ProductFilter{
		From: from,
		To:   to,
	}

	where, args := buildProductFilters(filter)

	require.Equal(t,
		"WHERE p.status IN (?, ?) AND p.ready_date >= ? AND p.ready_date < ?",
		where,
	)

	require.Equal(t,
		[]interface{}{
			"assigned",
			"final",
			"2025-01-10",
			"2025-01-21",
		},
		args,
	)
}

func TestBuildProductFilters_AllFilters(t *testing.T) {
	filter := ProductFilter{
		From:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		To:       time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		OrderNum: "123",
		Type:     []string{"window", "door"},
	}

	where, args := buildProductFilters(filter)

	require.Equal(t,
		"WHERE p.status IN (?, ?) AND p.ready_date >= ? AND p.ready_date < ? AND p.order_num LIKE ? AND p.type IN (?,?)",
		where,
	)

	require.Equal(t,
		[]interface{}{
			"assigned",
			"final",
			"2025-01-01",
			"2025-02-01",
			"%123%",
			"window",
			"door",
		},
		args,
	)
}
