package mysql

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"
	"vue-golang/internal/storage"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetGutterOrderDetails_NewRecord_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var reqID int64 = 1

	productRows := sqlmock.NewRows([]string{
		"id",
		"order_num",
		"name",
		"count",
		"total_time",
		"created_at",
		"type",
		"part_type",
		"parent_product_id",
		"parent_assembly",
		"status",
		"position",
	}).AddRow(
		reqID,
		"12345",
		"Водоотлив",
		2,
		1.5,
		time.Now(),
		"vodootliv",
		"main",
		0,
		"",
		"in_production",
		0,
	)

	operationRows := sqlmock.NewRows([]string{
		"id",
		"operation_name",
		"operation_label",
		"count",
		"value",
		"minutes",
		"sort_operation",
	}).AddRow(
		1,
		"cut",
		"Резка",
		2,
		1.5,
		90,
		0,
	)

	executorRows := sqlmock.NewRows([]string{
		"employee_id",
		"actual_minutes",
		"actual_value",
	}).AddRow(
		1,
		90,
		1.5,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position
		FROM dem_product_instances_al 
	`)).
		WithArgs(reqID).
		WillReturnRows(productRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, operation_name, operation_label, count, value, minutes, sort_operation 
            FROM dem_operation_values_al 
            WHERE product_id = ? 
            ORDER BY sort_operation ASC, id ASC
		`)).
		WithArgs(reqID).
		WillReturnRows(operationRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT employee_id, actual_minutes, actual_value 
            FROM dem_operation_executors_al 
            WHERE product_id = ? AND operation_name = ?
		`)).
		WithArgs(reqID, "cut").
		WillReturnRows(executorRows)

	result, err := stor.GetGutterOrderDetails(context.Background(), reqID)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, reqID, result.ID)
	require.Equal(t, "vodootliv", result.Type)

	require.Len(t, result.Operations, 1)

	require.Equal(t, "cut", result.Operations[0].Name)
	require.Len(t, result.Operations[0].AssignedWorkers, 1)
	require.Equal(t, int64(1), result.Operations[0].AssignedWorkers[0].EmployeeID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGutterOrderDetails_FindByOrderNum_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var reqID int64 = 1
	var LegacyID int64 = 100
	var reqOrderNum string = "123"

	productRows := sqlmock.NewRows([]string{
		"id",
		"order_num",
		"name",
		"count",
		"total_time",
		"created_at",
		"type",
		"part_type",
		"parent_product_id",
		"parent_assembly",
		"status",
		"position",
	}).AddRow(
		LegacyID,
		"123",
		"Водоотлив",
		2,
		1.5,
		time.Now(),
		"vodootliv",
		"main",
		0,
		"",
		"in_production",
		0,
	)

	operationRows := sqlmock.NewRows([]string{
		"id",
		"operation_name",
		"operation_label",
		"count",
		"value",
		"minutes",
		"sort_operation",
	}).AddRow(
		1,
		"cut",
		"Резка",
		2,
		1.5,
		90,
		0,
	)

	executorRows := sqlmock.NewRows([]string{
		"employee_id",
		"actual_minutes",
		"actual_value",
	}).AddRow(
		1,
		90,
		1.5,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position
		FROM dem_product_instances_al
	`)).
		WithArgs(reqID).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT numorders FROM dem_orders WHERE idorders = ? AND class_id = 6
		`)).
		WithArgs(reqID).
		WillReturnRows(sqlmock.NewRows([]string{"numorders"}).AddRow(reqOrderNum))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position
		FROM dem_product_instances_al 
		WHERE order_num = ? AND type = 'vodootliv' AND part_type = 'main'
		ORDER BY id DESC LIMIT 1`)).
		WithArgs(reqOrderNum).
		WillReturnRows(productRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, operation_name, operation_label, count, value, minutes, sort_operation 
            FROM dem_operation_values_al 
            WHERE product_id = ? 
            ORDER BY sort_operation ASC, id ASC
		`)).
		WithArgs(LegacyID).
		WillReturnRows(operationRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT employee_id, actual_minutes, actual_value 
            FROM dem_operation_executors_al 
            WHERE product_id = ? AND operation_name = ?
		`)).
		WithArgs(LegacyID, "cut").
		WillReturnRows(executorRows)

	result, err := stor.GetGutterOrderDetails(context.Background(), reqID)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, LegacyID, result.ID)
	require.Equal(t, "123", result.OrderNum)
	require.Equal(t, "vodootliv", result.Type)

	require.Len(t, result.Operations, 1)

	require.Equal(t, "cut", result.Operations[0].Name)
	require.Len(t, result.Operations[0].AssignedWorkers, 1)
	require.Equal(t, int64(1), result.Operations[0].AssignedWorkers[0].EmployeeID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGutterOrderDetails_RequiresCalculator(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		reqID       int64 = 1
		reqOrderNum       = "123"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position
		FROM dem_product_instances_al
	`)).
		WithArgs(reqID).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT numorders FROM dem_orders WHERE idorders = ? AND class_id = 6
		`)).
		WithArgs(reqID).
		WillReturnRows(sqlmock.NewRows([]string{"numorders"}).AddRow(reqOrderNum))

	// TODO тут првоерку для нащельников
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM dem_param_vo WHERE orderid = ? AND mat IN (3,4,7,9)
	`)).
		WithArgs(reqID).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1))

	result, err := stor.GetGutterOrderDetails(context.Background(), reqID)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "REQUIRES_CALCULATOR")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetGutterOrderDetails_LegacyNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const reqID int64 = 1

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position
		FROM dem_product_instances_al
	`)).
		WithArgs(reqID).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT numorders FROM dem_orders WHERE idorders = ? AND class_id = 6
	`)).
		WithArgs(reqID).
		WillReturnError(sql.ErrNoRows)

	result, err := stor.GetGutterOrderDetails(context.Background(), reqID)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "legacy order not found")
	require.Contains(t, err.Error(), "id=1")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportGutterFromLegacy_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		legacyID  int64 = 100
		orderNum        = "123"
		orderName       = "vo_order"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Водоотлив/Оцинковка' as name, 'vodootliv' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = 6
		`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"idorders", "numorders", "ordername", "name", "type", "part_type", "status", "template_code"}).
			AddRow(legacyID, orderNum, orderName, "Водоотлив/Оцинковка", "vodootliv", "main", "in_production", ""))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM dem_param_vo 
		WHERE orderid = ? AND mat IN (3, 4, 7, 9)
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT SUM(kol_vo) as kol_vo, SUM(kol_vo * h * b) as sqr, SUM(kol_vo * h) as pgm 
		FROM dem_param_vo 
		WHERE orderid = ? AND mat IN (1, 5, 8) AND mat NOT IN (10, 11)
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"kol_vo", "sqr", "pgm"}).AddRow(2, 1000000, 100))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT mat, SUM(kol_vo) FROM dem_param_vo WHERE orderid = ? AND mat IN (1, 5, 8) GROUP BY mat
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"mat", "kol_vo"}).AddRow(1, 2))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"name_mat", "total_hours", "total_count"}).AddRow("Разметка", 1.5, 2))

		//TODO транзакция началась
	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id FROM dem_product_instances_al 
		WHERE order_num = ? AND type = 'vodootliv' LIMIT 1
	`)).
		WithArgs(orderNum).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(regexp.QuoteMeta(`
	INSERT INTO dem_product_instances_al (
		order_num, template_code, name, customer, count, total_time, 
		type, part_type, parent_assembly, status, position, sqr, type_izd, profile, systema
	) VALUES (?, '0', ?, ?, ?, ?, 'vodootliv', 'main', '', ?, 0, ?, ?, '', '')
	`)).
		WithArgs(orderNum, "Водоотлив", orderName, float64(2), float64(1.5), "in_production", float64(1), "vo").
		WillReturnResult(sqlmock.NewResult(500, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
    INSERT INTO dem_operation_values_al (product_id, operation_name, operation_label, count, value, minutes, sort_operation)
			VALUES (?, ?, ?, ?, ?, ?, ?)	
	`)).
		WithArgs(
			500, // id из NewResult предыдущего INSERT
			"разметка",
			"Разметка",
			float64(2),
			float64(1.5),
			float64(90.0),
			0,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// TODO транзакция закоммичена
	mock.ExpectCommit()

	result, err := stor.importGutterFromLegacy(context.Background(), legacyID, orderNum)

	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, int64(500), result.ID)
	require.Equal(t, "123", result.OrderNum)
	require.Equal(t, "Водоотлив", result.Name)
	require.Equal(t, "vodootliv", result.Type)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportGutterFromLegacy_LegacyOrderNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		legacyID int64 = 100
		orderNum       = "123"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Водоотлив/Оцинковка' as name, 'vodootliv' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = 6
		`)).
		WithArgs(legacyID).
		WillReturnError(sql.ErrNoRows)

	result, err := stor.importGutterFromLegacy(context.Background(), legacyID, orderNum)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "not found")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportGutterFromLegacy_HasNashchelnik(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		legacyID  int64 = 100
		orderNum        = "123"
		orderName       = "321"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Водоотлив/Оцинковка' as name, 'vodootliv' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = 6
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"idorders",
			"numorders",
			"ordername",
			"name",
			"type",
			"part_type",
			"status",
			"template_code",
		}).AddRow(
			legacyID,
			orderNum,
			orderName,
			"Водоотлив/Оцинковка",
			"vodootliv",
			"main",
			"in_production",
			"",
		))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM dem_param_vo 
		WHERE orderid = ? AND mat IN (3, 4, 7, 9)
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1),
		)

	result, err := stor.importGutterFromLegacy(context.Background(), legacyID, orderNum)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "HAS_NASHCHELNIK")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportGutterFromLegacy_InsertProductError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		legacyID  int64 = 100
		orderNum        = "123"
		orderName       = "321"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Водоотлив/Оцинковка' as name, 'vodootliv' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = 6
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"idorders",
			"numorders",
			"ordername",
			"name",
			"type",
			"part_type",
			"status",
			"template_code",
		}).AddRow(
			legacyID,
			orderNum,
			orderName,
			"Водоотлив/Оцинковка",
			"vodootliv",
			"main",
			"in_production",
			"",
		))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM dem_param_vo 
		WHERE orderid = ? AND mat IN (3, 4, 7, 9)
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT SUM(kol_vo) as kol_vo, SUM(kol_vo * h * b) as sqr, SUM(kol_vo * h) as pgm 
	FROM dem_param_vo 
	WHERE orderid = ? AND mat IN (1, 5, 8) AND mat NOT IN (10, 11)
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"kol_vo", "sqr", "pgm"}).
				AddRow(2, 1000000, 100),
		)

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT mat, SUM(kol_vo) FROM dem_param_vo WHERE orderid = ? AND mat IN (1, 5, 8) GROUP BY mat
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"mat", "kol_vo"}).
				AddRow(1, 2),
		)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"name_mat", "total_hours", "total_count"}).
				AddRow("Разметка", 1.5, 2),
		)

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT id FROM dem_product_instances_al 
	WHERE order_num = ? AND type = 'vodootliv' LIMIT 1
	`)).
		WithArgs(orderNum).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(regexp.QuoteMeta(`
	INSERT INTO dem_product_instances_al (
		order_num, template_code, name, customer, count, total_time, 
		type, part_type, parent_assembly, status, position, sqr, type_izd, profile, systema
	) VALUES (?, '0', ?, ?, ?, ?, 'vodootliv', 'main', '', ?, 0, ?, ?, '', '')
	`)).
		WithArgs(
			orderNum,
			"Водоотлив",
			orderName,
			float64(2),
			float64(1.5),
			"in_production",
			float64(1),
			"vo",
		).
		WillReturnError(errors.New("database insert error"))

	mock.ExpectRollback()

	result, err := stor.importGutterFromLegacy(context.Background(), legacyID, orderNum)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "database insert error")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportGutterFromLegacy_InsertOperationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		legacyID  int64 = 100
		orderNum        = "123"
		orderName       = "321"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Водоотлив/Оцинковка' as name, 'vodootliv' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = 6
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"idorders",
			"numorders",
			"ordername",
			"name",
			"type",
			"part_type",
			"status",
			"template_code",
		}).AddRow(
			legacyID,
			orderNum,
			orderName,
			"Водоотлив/Оцинковка",
			"vodootliv",
			"main",
			"in_production",
			"",
		))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM dem_param_vo 
		WHERE orderid = ? AND mat IN (3, 4, 7, 9)
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT SUM(kol_vo) as kol_vo, SUM(kol_vo * h * b) as sqr, SUM(kol_vo * h) as pgm 
	FROM dem_param_vo 
	WHERE orderid = ? AND mat IN (1, 5, 8) AND mat NOT IN (10, 11)
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"kol_vo", "sqr", "pgm"}).
				AddRow(2, 1000000, 100),
		)

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT mat, SUM(kol_vo) FROM dem_param_vo WHERE orderid = ? AND mat IN (1, 5, 8) GROUP BY mat
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"mat", "kol_vo"}).
				AddRow(1, 2),
		)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"name_mat", "total_hours", "total_count"}).
				AddRow("Разметка", 1.5, 2),
		)

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT id FROM dem_product_instances_al 
	WHERE order_num = ? AND type = 'vodootliv' LIMIT 1
	`)).
		WithArgs(orderNum).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(regexp.QuoteMeta(`
	INSERT INTO dem_product_instances_al (
		order_num, template_code, name, customer, count, total_time, 
		type, part_type, parent_assembly, status, position, sqr, type_izd, profile, systema
	) VALUES (?, '0', ?, ?, ?, ?, 'vodootliv', 'main', '', ?, 0, ?, ?, '', '')
	`)).
		WithArgs(
			orderNum,
			"Водоотлив",
			orderName,
			float64(2),
			float64(1.5),
			"in_production",
			float64(1),
			"vo",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
    INSERT INTO dem_operation_values_al (product_id, operation_name, operation_label, count, value, minutes, sort_operation)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`)).
		WithArgs(
			int64(1),
			"разметка",
			"Разметка",
			float64(2),
			float64(1.5),
			float64(90),
			0,
		).
		WillReturnError(errors.New("operation insert failed"))

	mock.ExpectRollback()

	result, err := stor.importGutterFromLegacy(context.Background(), legacyID, orderNum)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "insert operation")
	require.Contains(t, err.Error(), "operation insert failed")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportGutterFromLegacy_AlreadyExists_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		legacyID  int64 = 100
		orderNum        = "123"
		orderName       = "321"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Водоотлив/Оцинковка' as name, 'vodootliv' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = 6
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"idorders",
			"numorders",
			"ordername",
			"name",
			"type",
			"part_type",
			"status",
			"template_code",
		}).AddRow(
			legacyID,
			orderNum,
			orderName,
			"Водоотлив/Оцинковка",
			"vodootliv",
			"main",
			"in_production",
			"",
		))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM dem_param_vo 
		WHERE orderid = ? AND mat IN (3, 4, 7, 9)
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT SUM(kol_vo) as kol_vo, SUM(kol_vo * h * b) as sqr, SUM(kol_vo * h) as pgm 
	FROM dem_param_vo 
	WHERE orderid = ? AND mat IN (1, 5, 8) AND mat NOT IN (10, 11)
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"kol_vo", "sqr", "pgm"}).
				AddRow(2, 1000000, 100),
		)

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT mat, SUM(kol_vo) FROM dem_param_vo WHERE orderid = ? AND mat IN (1, 5, 8) GROUP BY mat
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"mat", "kol_vo"}).
				AddRow(1, 2),
		)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"name_mat", "total_hours", "total_count"}).
				AddRow("Разметка", 1.5, 2),
		)

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT id FROM dem_product_instances_al 
	WHERE order_num = ? AND type = 'vodootliv' LIMIT 1
	`)).
		WithArgs(orderNum).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(500))

	mock.ExpectRollback()

	result, err := stor.importGutterFromLegacy(context.Background(), legacyID, orderNum)

	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, int64(500), result.ID)
	require.Equal(t, orderNum, result.OrderNum)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportGutterFromLegacy_NoOperations(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		legacyID  int64 = 100
		orderNum        = "123"
		orderName       = "321"
	)

	// 1. dem_orders
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Водоотлив/Оцинковка' as name, 'vodootliv' as type, 'main' as part_type, 'in_production' as status, '' as template_code
		FROM dem_orders
		WHERE idorders = ? AND class_id = 6
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"idorders",
			"numorders",
			"ordername",
			"name",
			"type",
			"part_type",
			"status",
			"template_code",
		}).AddRow(
			legacyID,
			orderNum,
			orderName,
			"Водоотлив/Оцинковка",
			"vodootliv",
			"main",
			"in_production",
			"",
		))

	// 2. Проверка нащельника
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FROM dem_param_vo 
		WHERE orderid = ? AND mat IN (3, 4, 7, 9)
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"COUNT(*)"}).
				AddRow(0),
		)

	// 3. Параметры изделия
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT SUM(kol_vo) as kol_vo, SUM(kol_vo * h * b) as sqr, SUM(kol_vo * h) as pgm 
		FROM dem_param_vo 
		WHERE orderid = ? AND mat IN (1, 5, 8) AND mat NOT IN (10, 11)
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"kol_vo", "sqr", "pgm"}).
				AddRow(2, 1000000, 100),
		)

	// 4. Тип материала
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT mat, SUM(kol_vo) FROM dem_param_vo WHERE orderid = ? AND mat IN (1, 5, 8) GROUP BY mat
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"mat", "kol_vo"}).
				AddRow(1, 2),
		)

	// 5. Самое важное - операции пустые
	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"name_mat",
				"total_hours",
				"total_count",
			}),
		)

	result, err := stor.importGutterFromLegacy(context.Background(), legacyID, orderNum)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "no operations")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNashchelnikRawData_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var legacyID int64 = 100

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT numorders, ordername 
	FROM dem_orders 
	WHERE idorders = ? AND class_id = 6
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"numorders", "ordername"}).
			AddRow("123", "321"))

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT SUM(kol_vo) as kol_vo, SUM(kol_vo * h * b) as sqr, SUM(kol_vo * CASE WHEN h > b THEN h ELSE b END) as pgm FROM dem_param_vo WHERE orderid = ? AND mat NOT IN (10,11)
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"kol_vo", "sqr", "pgm"}).
			AddRow(2, 200, 20))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"name_mat", "total_hours", "total_count"}).
			AddRow("Разметка", 2, 200))

	result, err := stor.GetNashchelnikRawData(context.Background(), legacyID)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, int64(legacyID), result.LegacyID)
	require.Equal(t, "123", result.OrderNum)

	require.Len(t, result.ExistingOps, 1)
	require.Equal(t, "разметка", result.ExistingOps[0].Name)

	require.Equal(t, float64(200), result.ExistingOps[0].Count)
	require.Equal(t, float64(2), result.ExistingOps[0].Value)
	require.Equal(t, float64(120), result.ExistingOps[0].Minutes)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNashchelnikRawData_OrderNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var legacyID int64 = 100

	mock.ExpectQuery(regexp.QuoteMeta(`
	SELECT numorders, ordername 
	FROM dem_orders 
	WHERE idorders = ? AND class_id = 6
	`)).
		WithArgs(legacyID).
		WillReturnError(sql.ErrNoRows)

	result, err := stor.GetNashchelnikRawData(context.Background(), legacyID)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "order not found")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveNashchelnikNorm_CreateNew_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	legacyID := int64(100)
	orderNum := "123"

	operations := []storage.NormOperation{
		{
			Name:    "разметка",
			Label:   "Разметка",
			Count:   2,
			Value:   1.5,
			Minutes: 90,
		},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT ordername FROM dem_orders WHERE idorders = ?
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"ordername"}).
				AddRow("Иванов"),
		)

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id FROM dem_product_instances_al WHERE order_num = ? AND type = 'vodootliv' LIMIT 1 FOR UPDATE
	`)).
		WithArgs(orderNum).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO dem_product_instances_al (
			order_num, template_code, name, customer, count, total_time,
			type, part_type, parent_assembly, status, position, sqr, type_izd, profile
		) VALUES (
			?, '0', 'Водоотлив', ?, ?, ?,
			'vodootliv', 'main', '', ?, 0, ?, 'vo', ''
		)
	`)).
		WithArgs(
			orderNum,
			"Иванов",
			float64(2),
			float64(1.5),
			sqlmock.AnyArg(),
			float64(10),
		).
		WillReturnResult(sqlmock.NewResult(500, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO dem_operation_values_al
	`)).
		WithArgs(
			int64(500),
			"разметка",
			"Разметка",
			float64(2),
			float64(1.5),
			float64(90),
			0,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	result, err := stor.SaveNashchelnikNorm(
		context.Background(),
		legacyID,
		orderNum,
		0, 0, 0, 0,
		10,
		2,
		operations,
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, int64(500), result.ID)
	require.Equal(t, "vodootliv", result.Type)
	require.Len(t, result.Operations, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveNashchelnikNorm_Update_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	legacyID := int64(100)
	orderNum := "123"

	operations := []storage.NormOperation{
		{
			Name:    "разметка",
			Label:   "Разметка",
			Count:   2,
			Value:   1.5,
			Minutes: 90,
		},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT ordername FROM dem_orders WHERE idorders = ?
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"ordername"}).
				AddRow("Иванов"),
		)

	mock.ExpectBegin()

	// проверяем существующий продукт
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id FROM dem_product_instances_al WHERE order_num = ? AND type = 'vodootliv' LIMIT 1 FOR UPDATE
	`)).
		WithArgs(orderNum).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(100),
		)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE dem_product_instances_al SET total_time = ?, customer = ?, count = ?, sqr = ? WHERE id = ?`)).
		WithArgs(
			float64(1.5),
			"Иванов",
			float64(2),
			float64(10),
			int64(100),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM dem_operation_values_al WHERE product_id = ?`)).
		WithArgs(100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO dem_operation_values_al
	`)).
		WithArgs(
			int64(100),
			"разметка",
			"Разметка",
			float64(2),
			float64(1.5),
			float64(90),
			0,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	result, err := stor.SaveNashchelnikNorm(
		context.Background(),
		legacyID,
		orderNum,
		0, 0, 0, 0,
		10,
		2,
		operations,
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, int64(100), result.ID)
	require.Equal(t, "vodootliv", result.Type)
	require.Len(t, result.Operations, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveNashchelnikNorm_InsertOperationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	legacyID := int64(100)
	orderNum := "123"

	operations := []storage.NormOperation{
		{
			Name:    "разметка",
			Label:   "Разметка",
			Count:   2,
			Value:   1.5,
			Minutes: 90,
		},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT ordername FROM dem_orders WHERE idorders = ?
	`)).
		WithArgs(legacyID).
		WillReturnRows(
			sqlmock.NewRows([]string{"ordername"}).
				AddRow("Иванов"),
		)

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id FROM dem_product_instances_al WHERE order_num = ? AND type = 'vodootliv' LIMIT 1 FOR UPDATE
	`)).
		WithArgs(orderNum).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).
				AddRow(100),
		)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE dem_product_instances_al SET total_time = ?, customer = ?, count = ?, sqr = ? WHERE id = ?`)).
		WithArgs(
			float64(1.5),
			"Иванов",
			float64(2),
			float64(10),
			int64(100),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM dem_operation_values_al WHERE product_id = ?`)).
		WithArgs(100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO dem_operation_values_al (product_id, operation_name, operation_label, count, value, minutes, sort_operation)
			VALUES (?, ?, ?, ?, ?, ?, ?)
	`)).
		WithArgs(
			int64(100),
			"разметка",
			"Разметка",
			float64(2),
			float64(1.5),
			float64(90),
			0,
		).
		WillReturnError(errors.New("insert failed"))

	mock.ExpectRollback()

	result, err := stor.SaveNashchelnikNorm(
		context.Background(),
		legacyID,
		orderNum,
		0, 0, 0, 0,
		10,
		2,
		operations,
	)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "duplicate op")
	require.Contains(t, err.Error(), "Разметка")
	require.Contains(t, err.Error(), "insert failed")

	require.NoError(t, mock.ExpectationsWereMet())
}
