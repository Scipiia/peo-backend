package mysql

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"vue-golang/internal/storage"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestSaveNormOrder_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var parentProductID int64 = 100
	req := storage.OrderNormDetails{
		OrderNum:        "order_num",
		TemplateCode:    "template_code",
		Name:            "name",
		Count:           float64(2),
		TotalTime:       float64(1.5),
		Type:            "door",
		PartType:        "main",
		ParentAssembly:  "",
		ParentProductID: &parentProductID,
		Customer:        "Ivanov",
		Position:        1,
		Status:          "assigned",
		Systema:         "x",
		TypeIzd:         "2p",
		Profile:         "alutech",
		Sqr:             float64(100),
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO dem_product_instances_al (order_num, template_code, name, count, total_time, type, part_type, 
            parent_assembly, parent_product_id, customer, position, status, systema, type_izd, profile, sqr) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?,?,?)`)).
		WithArgs(req.OrderNum, req.TemplateCode, req.Name, req.Count, req.TotalTime, req.Type, req.PartType, req.ParentAssembly, req.ParentProductID, req.Customer, req.Position, req.Status, req.Systema, req.TypeIzd, req.Profile, req.Sqr).
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := stor.SaveNormOrder(context.Background(), req)
	require.NoError(t, err)
	require.NotZero(t, result)
	require.Equal(t, int64(1), result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveNormOrder_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var parentProductID int64 = 100
	req := storage.OrderNormDetails{
		OrderNum:        "order_num",
		TemplateCode:    "template_code",
		Name:            "name",
		Count:           float64(2),
		TotalTime:       float64(1.5),
		Type:            "door",
		PartType:        "main",
		ParentAssembly:  "",
		ParentProductID: &parentProductID,
		Customer:        "Ivanov",
		Position:        1,
		Status:          "assigned",
		Systema:         "x",
		TypeIzd:         "2p",
		Profile:         "alutech",
		Sqr:             float64(100),
	}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO dem_product_instances_al (order_num, template_code, name, count, total_time, type, part_type, 
            parent_assembly, parent_product_id, customer, position, status, systema, type_izd, profile, sqr) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?,?,?)`)).
		WithArgs(req.OrderNum, req.TemplateCode, req.Name, req.Count, req.TotalTime, req.Type, req.PartType, req.ParentAssembly, req.ParentProductID, req.Customer, req.Position, req.Status, req.Systema, req.TypeIzd, req.Profile, req.Sqr).
		WillReturnError(errors.New("Ошибка сохранения нормировки в базу"))

	result, err := stor.SaveNormOrder(context.Background(), req)
	require.Error(t, err)
	require.Zero(t, result)
	require.Contains(t, err.Error(), "Ошибка сохранения нормировки в базу")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveNormOperation_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var OrderID int64 = 100

	req := []storage.NormOperation{
		{
			Name:    "name",
			Label:   "label",
			Count:   float64(2),
			Value:   float64(1.5),
			Minutes: float64(2.5),
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO dem_operation_values_al 
			(product_id, operation_name, operation_label, count, value, minutes, sort_operation)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    operation_name = VALUES(operation_name),
			count = VALUES(count),
			value = VALUES(value),
			sort_operation = VALUES(sort_operation)`)).
		ExpectExec().
		WithArgs(OrderID, "name", "label", float64(2), float64(1.5), float64(2.5), int64(0)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = stor.SaveNormOperation(context.Background(), OrderID, req)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveNormOperation_PrepareError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var OrderID int64 = 100

	req := []storage.NormOperation{
		{
			Name:    "name",
			Label:   "label",
			Count:   float64(2),
			Value:   float64(1.5),
			Minutes: float64(2.5),
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO dem_operation_values_al 
			(product_id, operation_name, operation_label, count, value, minutes, sort_operation)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    operation_name = VALUES(operation_name),
			count = VALUES(count),
			value = VALUES(value),
			sort_operation = VALUES(sort_operation)`)).
		ExpectExec().
		WithArgs(OrderID, "name", "label", float64(2), float64(1.5), float64(2.5), int64(0)).
		WillReturnError(errors.New("prepare statement"))

	mock.ExpectRollback()

	err = stor.SaveNormOperation(context.Background(), OrderID, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "prepare statement")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveNormOperation_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var OrderID int64 = 100

	req := []storage.NormOperation{
		{
			Name:    "name",
			Label:   "label",
			Count:   float64(2),
			Value:   float64(1.5),
			Minutes: float64(2.5),
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO dem_operation_values_al 
			(product_id, operation_name, operation_label, count, value, minutes, sort_operation)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    operation_name = VALUES(operation_name),
			count = VALUES(count),
			value = VALUES(value),
			sort_operation = VALUES(sort_operation)`)).
		ExpectExec().
		WithArgs(OrderID, "name", "label", float64(2), float64(1.5), float64(2.5), int64(0)).
		WillReturnError(errors.New("Ошибка сохранения нормированных операции"))

	mock.ExpectRollback()

	err = stor.SaveNormOperation(context.Background(), OrderID, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Ошибка сохранения нормированных операции")

	require.NoError(t, mock.ExpectationsWereMet())
}
