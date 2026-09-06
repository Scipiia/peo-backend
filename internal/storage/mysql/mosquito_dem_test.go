package mysql

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetMosquitoOrderDetails_ByNewID_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var reqID int64 = 1

	productRows := sqlmock.NewRows([]string{"id", "order_num", "name", "count", "total_time", "created_at", "type", "part_type", "parent_product_id", "parent_assembly", "status", "position", "ready_date"}).
		AddRow(reqID, "12345", "москитка", 2, 1.5, time.Now(), "mosquito", "main", 0, "", "in_production", 0, time.Now())

	operationRows := sqlmock.NewRows([]string{"id", "operation_name", "operation_label", "count", "value", "minutes", "sort_operation"}).AddRow(1, "cut", "Резка", 2, 1.5, 90, 0)

	executorRows := sqlmock.NewRows([]string{"employee_id", "actual_minutes", "actual_value"}).AddRow(1, 90, 1.5)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position, ready_date
		FROM dem_product_instances_al 
		WHERE id = ? AND type = 'mosquito' AND part_type = 'main'`)).
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

	result, err := stor.GetMosquitoOrderDetails(context.Background(), reqID)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, reqID, result.ID)
	require.Equal(t, "mosquito", result.Type)
	require.Equal(t, "12345", result.OrderNum)
	require.Equal(t, "Резка", result.Operations[0].Label)

	require.Len(t, result.Operations, 1)

	require.Equal(t, "cut", result.Operations[0].Name)
	require.Len(t, result.Operations[0].AssignedWorkers, 1)
	require.Equal(t, int64(1), result.Operations[0].AssignedWorkers[0].EmployeeID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMosquitoOrderDetails_FoundByOrderNum_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var reqID int64 = 1
	var reqOrderNum = "12345"

	productRows := sqlmock.NewRows([]string{"id", "order_num", "name", "count", "total_time", "created_at", "type", "part_type", "parent_product_id", "parent_assembly", "status", "position"}).
		AddRow(reqID, reqOrderNum, "москитка", 2, 1.5, time.Now(), "mosquito", "main", 0, "", "in_production", 0)

	operationRows := sqlmock.NewRows([]string{"id", "operation_name", "operation_label", "count", "value", "minutes", "sort_operation"}).AddRow(1, "cut", "Резка", 2, 1.5, 90, 0)

	executorRows := sqlmock.NewRows([]string{"employee_id", "actual_minutes", "actual_value"}).AddRow(1, 90, 1.5)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position, ready_date
		FROM dem_product_instances_al 
		WHERE id = ? AND type = 'mosquito' AND part_type = 'main'`)).
		WithArgs(reqID).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT numorders FROM dem_orders WHERE idorders = ? AND class_id = 4
		`)).
		WithArgs(reqID).
		WillReturnRows(sqlmock.NewRows([]string{"numorders"}).AddRow(reqOrderNum))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position
		FROM dem_product_instances_al 
		WHERE order_num = ? AND type = 'mosquito' AND part_type = 'main' 
		ORDER BY id DESC LIMIT 1`)).
		WithArgs(reqOrderNum).
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

	result, err := stor.GetMosquitoOrderDetails(context.Background(), reqID)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, reqID, result.ID)
	require.Equal(t, "mosquito", result.Type)
	require.Equal(t, "12345", result.OrderNum)
	require.Equal(t, "Резка", result.Operations[0].Label)

	require.Len(t, result.Operations, 1)

	require.Equal(t, "cut", result.Operations[0].Name)
	require.Len(t, result.Operations[0].AssignedWorkers, 1)
	require.Equal(t, int64(1), result.Operations[0].AssignedWorkers[0].EmployeeID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMosquitoOrderDetails_LegacyNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const reqID int64 = 1

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, order_num, name, count, total_time, created_at, type, part_type, parent_product_id, parent_assembly, status, position, ready_date
		FROM dem_product_instances_al 
		WHERE id = ? AND type = 'mosquito' AND part_type = 'main'`)).
		WithArgs(reqID).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT numorders FROM dem_orders WHERE idorders = ? AND class_id = 4
	`)).
		WithArgs(reqID).
		WillReturnError(sql.ErrNoRows)

	result, err := stor.GetMosquitoOrderDetails(context.Background(), reqID)

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "не удалось получить номер заказа из архива")
	require.Contains(t, err.Error(), "legacy_id=1")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportMosquitoFromLegacy_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const reqID int64 = 1

	const (
		legacyID  int64 = 100
		orderNum        = "123"
		orderName       = "ms_order"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Москитная сетка', 'mosquito' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = '4'
		`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"idorders", "numorders", "ordername", "name", "type", "part_type", "status", "template_code"}).
			AddRow(legacyID, orderNum, orderName, "Москитная сетка", "mosquito", "main", "in_production", ""))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT SUM(kol_vo) as kol_vo,SUM(kol_vo * weight * hight) sqr FROM dem_param_moskit WHERE orderid = ?
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"kol_vo", "sqr"}).AddRow(2, 1000000))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT vid, SUM(kol_vo) FROM dem_param_moskit WHERE orderid = ? GROUP BY vid
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"vid", "kol_vo"}).AddRow(5, 10))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
    ), d.name_mat ASC
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"name_mat", "total_value", "total_count"}).AddRow("Разметка", 1.5, 2))

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO dem_product_instances_al (
            order_num, template_code, name, customer, count, total_time, 
            type, part_type, parent_assembly, status, position, sqr, type_izd, profile, systema
        ) VALUES (?, '0', ?, ?, ?, ?, ?, ?, '', ?, ?, ?, ?, '', '')
	`)).
		WithArgs(orderNum, "Москитная сетка", orderName, 2.0, 1.5, "mosquito", "main", "in_production", int64(0), float64(1), "vsn").
		WillReturnResult(sqlmock.NewResult(100, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
    INSERT INTO dem_operation_values_al (
                product_id, operation_name, operation_label, count, value, minutes, sort_operation
            ) VALUES (?, ?, ?, ?, ?, ?, ?)
	`)).
		WithArgs(100, "Разметка", "Разметка", float64(2), float64(1.5), float64(90), int64(0)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	result, err := stor.importMosquitoFromLegacy(context.Background(), legacyID, orderNum)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, int64(100), result.ID)
	require.Equal(t, "mosquito", result.Type)
	require.Equal(t, orderNum, result.OrderNum)

	require.Empty(t, result.Operations)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportMosquitoFromLegacy_OrderNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const reqID int64 = 1

	const (
		legacyID  int64 = 100
		orderNum        = "123"
		orderName       = "ms_order"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Москитная сетка', 'mosquito' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = '4'
		`)).
		WithArgs(legacyID).
		WillReturnError(sql.ErrNoRows)

	result, err := stor.importMosquitoFromLegacy(context.Background(), legacyID, orderNum)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "not found")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTestImportMosquitoFromLegacy_TypeError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const reqID int64 = 1

	const (
		legacyID  int64 = 100
		orderNum        = "123"
		orderName       = "ms_order"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Москитная сетка', 'mosquito' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = '4'
		`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"idorders", "numorders", "ordername", "name", "type", "part_type", "status", "template_code"}).
			AddRow(legacyID, orderNum, orderName, "Москитная сетка", "mosquito", "main", "in_production", ""))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT SUM(kol_vo) as kol_vo,SUM(kol_vo * weight * hight) sqr FROM dem_param_moskit WHERE orderid = ?
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"kol_vo", "sqr"}).AddRow(2, 1000000))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT vid, SUM(kol_vo) FROM dem_param_moskit WHERE orderid = ? GROUP BY vid
	`)).
		WithArgs(legacyID).
		WillReturnError(errors.New("query type ms"))

	result, err := stor.importMosquitoFromLegacy(context.Background(), legacyID, orderNum)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "query type ms")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTestImportMosquitoFromLegacy_NoOperations(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const reqID int64 = 1

	const (
		legacyID  int64 = 100
		orderNum        = "123"
		orderName       = "ms_order"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Москитная сетка', 'mosquito' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = '4'
		`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"idorders", "numorders", "ordername", "name", "type", "part_type", "status", "template_code"}).
			AddRow(legacyID, orderNum, orderName, "Москитная сетка", "mosquito", "main", "in_production", ""))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT SUM(kol_vo) as kol_vo,SUM(kol_vo * weight * hight) sqr FROM dem_param_moskit WHERE orderid = ?
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"kol_vo", "sqr"}).AddRow(2, 1000000))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT vid, SUM(kol_vo) FROM dem_param_moskit WHERE orderid = ? GROUP BY vid
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"vid", "kol_vo"}).AddRow(5, 10))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
    ), d.name_mat ASC
	`)).
		WithArgs(legacyID).
		WillReturnError(errors.New("query operations"))

	result, err := stor.importMosquitoFromLegacy(context.Background(), legacyID, orderNum)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "query operations")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestImportMosquitoFromLegacy_InsertOperationsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const reqID int64 = 1

	const (
		legacyID  int64 = 100
		orderNum        = "123"
		orderName       = "ms_order"
	)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT idorders, numorders, ordername, 'Москитная сетка', 'mosquito' as type, 'main' as part_type, 'in_production' as status, '' as template_code
				FROM dem_orders
				WHERE idorders = ? AND class_id = '4'
		`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"idorders", "numorders", "ordername", "name", "type", "part_type", "status", "template_code"}).
			AddRow(legacyID, orderNum, orderName, "Москитная сетка", "mosquito", "main", "in_production", ""))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT SUM(kol_vo) as kol_vo,SUM(kol_vo * weight * hight) sqr FROM dem_param_moskit WHERE orderid = ?
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"kol_vo", "sqr"}).AddRow(2, 1000000))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT vid, SUM(kol_vo) FROM dem_param_moskit WHERE orderid = ? GROUP BY vid
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"vid", "kol_vo"}).AddRow(5, 10))

	mock.ExpectQuery(regexp.QuoteMeta(`
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
    ), d.name_mat ASC
	`)).
		WithArgs(legacyID).
		WillReturnRows(sqlmock.NewRows([]string{"name_mat", "total_value", "total_count"}).AddRow("Разметка", 1.5, 2))

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO dem_product_instances_al (
            order_num, template_code, name, customer, count, total_time, 
            type, part_type, parent_assembly, status, position, sqr, type_izd, profile, systema
        ) VALUES (?, '0', ?, ?, ?, ?, ?, ?, '', ?, ?, ?, ?, '', '')
	`)).
		WithArgs(orderNum, "Москитная сетка", orderName, 2.0, 1.5, "mosquito", "main", "in_production", int64(0), float64(1), "vsn").
		WillReturnResult(sqlmock.NewResult(100, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
    INSERT INTO dem_operation_values_al (
                product_id, operation_name, operation_label, count, value, minutes, sort_operation
            ) VALUES (?, ?, ?, ?, ?, ?, ?)
	`)).
		WithArgs(100, "Разметка", "Разметка", float64(2), float64(1.5), float64(90), int64(0)).
		WillReturnError(errors.New("вставка операции"))

	mock.ExpectRollback()

	result, err := stor.importMosquitoFromLegacy(context.Background(), legacyID, orderNum)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "вставка операции")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadOperationsForProduct_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const productID int64 = 1

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, operation_name, operation_label, count, value, minutes, sort_operation 
                FROM dem_operation_values_al 
                WHERE product_id = ? 
                ORDER BY sort_operation ASC, id ASC`)).
		WithArgs(productID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "operation_name", "operation_label", "count", "value", "minutes", "sort_operation"}).
			AddRow(1, "Разметка", "Разметка", 2, 1.5, 90, 0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT employee_id, actual_minutes, actual_value 
                     FROM dem_operation_executors_al 
                     WHERE product_id = ? AND operation_name = ?`)).
		WithArgs(productID, "Разметка").
		WillReturnRows(sqlmock.NewRows([]string{"employee_id", "actual_minutes", "actual_value"}).
			AddRow(1, 90, 1.5))

	result, err := stor.loadOperationsForProduct(context.Background(), productID)
	require.NoError(t, err)
	require.Len(t, result, 1)

	op := result[0]

	require.Equal(t, "Разметка", op.Name)
	require.Equal(t, "Разметка", op.Label)
	require.Equal(t, float64(2), op.Count)
	require.Equal(t, float64(1.5), op.Value)
	require.Equal(t, float64(90), op.Minutes)

	require.Len(t, op.AssignedWorkers, 1)
	require.Equal(t, int64(1), op.AssignedWorkers[0].EmployeeID)
	require.Equal(t, float64(90), op.AssignedWorkers[0].ActualMinutes)
	require.Equal(t, float64(1.5), op.AssignedWorkers[0].ActualValue)

	require.NoError(t, mock.ExpectationsWereMet())

}

func TestLoadOperationsForProduct_OperationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const productID int64 = 1

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, operation_name, operation_label, count, value, minutes, sort_operation 
                FROM dem_operation_values_al 
                WHERE product_id = ? 
                ORDER BY sort_operation ASC, id ASC`)).
		WithArgs(productID).
		WillReturnError(errors.New("scan operation"))

	result, err := stor.loadOperationsForProduct(context.Background(), productID)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "scan operation")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLoadOperationsForProduct_AssignedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const productID int64 = 1

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, operation_name, operation_label, count, value, minutes, sort_operation 
                FROM dem_operation_values_al 
                WHERE product_id = ? 
                ORDER BY sort_operation ASC, id ASC`)).
		WithArgs(productID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "operation_name", "operation_label", "count", "value", "minutes", "sort_operation"}).
			AddRow(1, "Разметка", "Разметка", 2, 1.5, 90, 0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT employee_id, actual_minutes, actual_value 
                     FROM dem_operation_executors_al 
                     WHERE product_id = ? AND operation_name = ?`)).
		WithArgs(productID, "Разметка").
		WillReturnError(errors.New("error iterating executors"))

	result, err := stor.loadOperationsForProduct(context.Background(), productID)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Empty(t, result[0].AssignedWorkers)

	require.NoError(t, mock.ExpectationsWereMet())
}
