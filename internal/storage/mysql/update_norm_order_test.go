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

func TestUpdateNormOrder_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}
	var status string = "assigned"
	var ID int64 = 100

	req := storage.UpdateOrderDetails{
		ID:        int64(100),
		Type:      "door",
		TotalTime: float64(1.5),
		Count:     float64(2),
		Status:    &status,
		Operations: []storage.NormOperation{
			{
				"cut",
				"резка",
				float64(2),
				float64(10),
				float64(1.5),
				nil,
			},
		},
	}

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE dem_product_instances_al SET total_time = ?, type = ?, status = ? WHERE id = ?`)).
		WithArgs(
			float64(1.5),
			"door",
			"assigned",
			int64(100),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM dem_operation_values_al WHERE product_id = ?`)).
		WithArgs(100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO dem_operation_values_al (product_id, operation_name, operation_label, count, value, minutes, sort_operation) VALUES (?, ?, ?, ?, ?, ?, ?)`)).
		ExpectExec().
		WithArgs(
			int64(100),
			"cut",
			"резка",
			float64(2),
			float64(10),
			float64(1.5),
			int64(0),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = stor.UpdateNormOrder(context.Background(), ID, req)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateNormOrder_Error(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock, ID int64, update storage.UpdateOrderDetails)
		wantError string
	}{
		{
			name: "Update norm order error",
			setupMock: func(mock sqlmock.Sqlmock, ID int64, update storage.UpdateOrderDetails) {
				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta(`UPDATE dem_product_instances_al SET total_time = ?, type = ?, status = ? WHERE id = ?`)).
					WithArgs(
						update.TotalTime,
						update.Type,
						update.Status,
						ID,
					).
					WillReturnError(errors.New("ошибка обновление основной информации об изделии"))

				mock.ExpectRollback()
			},
			wantError: "ошибка обновление основной информации об изделии",
		},
		{
			name: "Delete operation error",
			setupMock: func(mock sqlmock.Sqlmock, ID int64, update storage.UpdateOrderDetails) {
				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta(`UPDATE dem_product_instances_al SET total_time = ?, type = ?, status = ? WHERE id = ?`)).
					WithArgs(
						update.TotalTime,
						update.Type,
						update.Status,
						ID,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM dem_operation_values_al WHERE product_id = ?`)).
					WithArgs(ID).
					WillReturnError(errors.New("ошибка удаления старых операции"))

				mock.ExpectRollback()
			},
			wantError: "ошибка удаления старых операции",
		},
		{
			name: "Prepare error",
			setupMock: func(mock sqlmock.Sqlmock, ID int64, update storage.UpdateOrderDetails) {
				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta(`UPDATE dem_product_instances_al SET total_time = ?, type = ?, status = ? WHERE id = ?`)).
					WithArgs(
						update.TotalTime,
						update.Type,
						update.Status,
						ID,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM dem_operation_values_al WHERE product_id = ?`)).
					WithArgs(ID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO dem_operation_values_al`)).
					WillReturnError(errors.New("оошибка при подготовке вставки новых операции"))

				mock.ExpectRollback()
			},
			wantError: "ошибка при подготовке вставки новых операции",
		},
		{
			name: "Insert error",
			setupMock: func(mock sqlmock.Sqlmock, ID int64, update storage.UpdateOrderDetails) {
				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta(`UPDATE dem_product_instances_al SET total_time = ?, type = ?, status = ? WHERE id = ?`)).
					WithArgs(
						update.TotalTime,
						update.Type,
						update.Status,
						ID,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM dem_operation_values_al WHERE product_id = ?`)).
					WithArgs(ID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectPrepare(regexp.QuoteMeta(`INSERT INTO dem_operation_values_al (product_id, operation_name, operation_label, count, value, minutes, sort_operation) VALUES (?, ?, ?, ?, ?, ?, ?)`)).
					ExpectExec().
					WithArgs(
						ID,
						"cut",
						"резка",
						float64(2),
						float64(10),
						float64(1.5),
						int64(0),
					).
					WillReturnError(errors.New("оошибка вставки новых операции"))

				mock.ExpectRollback()
			},
			wantError: "ошибка вставки новых операции",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			stor := &Storage{db: db}

			var status string = "assigned"
			var ID int64 = 100

			req := storage.UpdateOrderDetails{
				ID:        ID,
				Type:      "door",
				TotalTime: float64(1.5),
				Count:     float64(2),
				Status:    &status,
				Operations: []storage.NormOperation{
					{
						"cut",
						"резка",
						float64(2),
						float64(10),
						float64(1.5),
						nil,
					},
				},
			}
			tt.setupMock(mock, ID, req)

			err = stor.UpdateNormOrder(context.Background(), ID, req)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateFinalOrder_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var ID int64 = 100
	var Brigade string = "door"
	var NormMoney float64 = 100
	var ParentAssembly string = "parent_assembly"
	var Profile string = "profile"
	var Sqr float64 = 100
	var Systema string = "x"
	var TypeIzd string = "type_izd"
	var CustomerType string = "customer_type"
	var Coefficient float64 = 100
	var SqrStv float64 = 100

	req := storage.UpdateFinalOrderDetails{
		ID:             ID,
		Brigade:        &Brigade,
		Systema:        &Systema,
		NormMoney:      &NormMoney,
		ParentAssembly: &ParentAssembly,
		Profile:        &Profile,
		Sqr:            &Sqr,
		TypeIzd:        &TypeIzd,
		CustomerType:   &CustomerType,
		Coefficient:    &Coefficient,
		SqrStv:         &SqrStv,
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE dem_product_instances_al SET customer_type = ?, norm_money = ?, profile = ?, sqr = ?, systema = ?, 
            parent_assembly = ?, brigade = ?, type_izd = ?, status = 'final', coefficient = ?, sqr_stv = ? WHERE id = ?`)).
		WithArgs(
			CustomerType, NormMoney, Profile, Sqr, Systema, ParentAssembly, Brigade, TypeIzd, Coefficient, SqrStv, ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = stor.UpdateFinalOrder(context.Background(), ID, req)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateFinalOrder_UpdateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var ID int64 = 100
	var Brigade string = "door"
	var NormMoney float64 = 100
	var ParentAssembly string = "parent_assembly"
	var Profile string = "profile"
	var Sqr float64 = 100
	var Systema string = "x"
	var TypeIzd string = "type_izd"
	var CustomerType string = "customer_type"
	var Coefficient float64 = 100
	var SqrStv float64 = 100

	req := storage.UpdateFinalOrderDetails{
		ID:             ID,
		Brigade:        &Brigade,
		Systema:        &Systema,
		NormMoney:      &NormMoney,
		ParentAssembly: &ParentAssembly,
		Profile:        &Profile,
		Sqr:            &Sqr,
		TypeIzd:        &TypeIzd,
		CustomerType:   &CustomerType,
		Coefficient:    &Coefficient,
		SqrStv:         &SqrStv,
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE dem_product_instances_al SET customer_type = ?, norm_money = ?, profile = ?, sqr = ?, systema = ?, 
            parent_assembly = ?, brigade = ?, type_izd = ?, status = 'final', coefficient = ?, sqr_stv = ? WHERE id = ?`)).
		WithArgs(
			CustomerType, NormMoney, Profile, Sqr, Systema, ParentAssembly, Brigade, TypeIzd, Coefficient, SqrStv, ID,
		).
		WillReturnError(errors.New("ошибка обновления заказа"))

	err = stor.UpdateFinalOrder(context.Background(), ID, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ошибка обновления заказа")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateStatus_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var ID int64 = 100
	var Status string = "assigned"

	mock.ExpectExec(regexp.QuoteMeta("UPDATE dem_product_instances_al SET status = ? WHERE id = ? OR parent_product_id = ?")).
		WithArgs(
			Status,
			ID,
			ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM dem_operation_executors_al WHERE product_id IN (SELECT id FROM dem_product_instances_al WHERE id = ? OR parent_product_id = ?)`)).
		WithArgs(
			ID,
			ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = stor.UpdateStatus(context.Background(), ID, Status)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateStatus_UpdateStatusError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var ID int64 = 100
	var Status string = "assigned"

	mock.ExpectExec(regexp.QuoteMeta("UPDATE dem_product_instances_al SET status = ? WHERE id = ? OR parent_product_id = ?")).
		WithArgs(
			Status,
			ID,
			ID,
		).
		WillReturnError(errors.New("ошибка обновления статуса"))

	err = stor.UpdateStatus(context.Background(), ID, Status)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ошибка обновления статуса")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateStatus_DeleteExecutorsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var ID int64 = 100
	var Status string = "assigned"

	mock.ExpectExec(regexp.QuoteMeta("UPDATE dem_product_instances_al SET status = ? WHERE id = ? OR parent_product_id = ?")).
		WithArgs(
			Status,
			ID,
			ID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM dem_operation_executors_al WHERE product_id IN (SELECT id FROM dem_product_instances_al WHERE id = ? OR parent_product_id = ?)`)).
		WithArgs(
			ID,
			ID,
		).
		WillReturnError(errors.New("ошибка удаления назначенных сотрудников заказа"))

	err = stor.UpdateStatus(context.Background(), ID, Status)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ошибка удаления назначенных сотрудников заказа")
	require.NoError(t, mock.ExpectationsWereMet())
}
