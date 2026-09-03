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

func TestGetAllWorkers(t *testing.T) {
	tests := []struct {
		name      string
		typeIzd   string
		mockError error
		scanError bool
		wantError bool
	}{
		{
			name:    "OK window",
			typeIzd: "window",
		},
		{
			name:    "Empty type",
			typeIzd: "",
		},
		{
			name:    "Unknown type",
			typeIzd: "unknown",
		},
		{
			name:      "Database error",
			typeIzd:   "window",
			mockError: errors.New("database error"),
			wantError: true,
		},
		{
			name:      "Scan error",
			typeIzd:   "window",
			scanError: true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			storage := &Storage{
				db: db,
			}
			var rows *sqlmock.Rows

			if tt.scanError {
				rows = sqlmock.NewRows([]string{"id"}).AddRow(1)
			} else {
				rows = sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Ivan").AddRow(2, "Petr")
			}

			if tt.mockError != nil {
				mock.ExpectQuery("SELECT DISTINCT").
					WithArgs("windows").
					WillReturnError(tt.mockError)
			} else if tt.typeIzd == "window" {
				mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT DISTINCT e.id, e.name FROM dem_employees_al e
                JOIN dem_employee_teams_al et ON e.id = et.employee_id
                JOIN dem_teams_al t ON et.team_id = t.id
                WHERE e.is_active = TRUE AND t.slug = ?
                ORDER BY e.name ASC`)).
					WithArgs("windows").
					WillReturnRows(rows)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`
        		SELECT DISTINCT e.id, e.name FROM dem_employees_al e
                WHERE e.is_active = TRUE
                ORDER BY e.name ASC`)).
					WillReturnRows(rows)
			}

			workers, err := storage.GetAllWorkers(context.Background(), tt.typeIzd)

			if tt.wantError {
				require.Error(t, err)
				require.NoError(t, mock.ExpectationsWereMet())
				return
			}

			require.NoError(t, err)
			require.Len(t, workers, 2)
			require.NoError(t, mock.ExpectationsWereMet())
			require.Equal(t, int64(1), workers[0].ID)
			require.Equal(t, "Ivan", workers[0].Name)
		})
	}
}

func TestSaveOperationWorkers_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	req := storage.SaveWorkers{
		RootProductID: 100,
		Assignments: []storage.OperationWorkers{
			{
				ProductID:     101,
				OperationName: "Сборка",
				EmployeeID:    5,
				ActualMinutes: 60,
				Notes:         "test",
				ActualValue:   10,
			},
		},
	}

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(
		"DELETE FROM dem_operation_executors_al",
	)).
		WithArgs(100, 100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectPrepare(regexp.QuoteMeta(
		"INSERT INTO dem_operation_executors_al",
	)).
		ExpectExec().
		WithArgs(
			int64(101),
			"Сборка",
			int64(5),
			float64(60),
			"test",
			float64(10),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = stor.SaveOperationWorkers(context.Background(), req)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveOperationWorkers_Errors(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock, req storage.SaveWorkers)
		wantError string
	}{
		{
			name: "Delete error",
			setupMock: func(mock sqlmock.Sqlmock, req storage.SaveWorkers) {
				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta(
					"DELETE FROM dem_operation_executors_al",
				)).
					WithArgs(
						req.RootProductID,
						req.RootProductID,
					).
					WillReturnError(errors.New("delete error"))

				mock.ExpectRollback()
			},
			wantError: "ошибка удаления старых назначении",
		},
		{
			name: "Insert error",
			setupMock: func(mock sqlmock.Sqlmock, req storage.SaveWorkers) {
				mock.ExpectBegin()

				mock.ExpectExec(
					"DELETE",
				).
					WithArgs(req.RootProductID, req.RootProductID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectPrepare("INSERT INTO dem_operation_executors_al").
					ExpectExec().
					WithArgs(
						int64(101),
						"Сборка",
						int64(5),
						float64(60),
						"",
						float64(10),
					).
					WillReturnError(errors.New("insert error"))

				mock.ExpectRollback()
			},
			wantError: "ошибка вставки новых назначенных сотрудников",
		},
		{
			name: "Prepare error",
			setupMock: func(mock sqlmock.Sqlmock, req storage.SaveWorkers) {
				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM dem_operation_executors_al")).
					WithArgs(req.RootProductID, req.RootProductID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO dem_operation_executors_al")).
					WillReturnError(errors.New("prepare error"))

				mock.ExpectRollback()
			},
			wantError: "ошибка подготовки запроса",
		},
		{
			name: "Update status error",
			setupMock: func(mock sqlmock.Sqlmock, req storage.SaveWorkers) {
				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta("DELETE FROM dem_operation_executors_al")).
					WithArgs(req.RootProductID, req.RootProductID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO dem_operation_executors_al")).
					ExpectExec().
					WithArgs(
						int64(101),
						"Сборка",
						int64(5),
						float64(60),
						"",
						float64(10),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(regexp.QuoteMeta("UPDATE dem_product_instances_al SET status = ? WHERE id = ? OR parent_product_id = ?")).
					WithArgs(
						req.UpdateStatus,
						req.RootProductID,
						req.RootProductID,
					).
					WillReturnError(errors.New("update status error"))

				mock.ExpectRollback()
			},
			wantError: "ошибка обновления статуса",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			stor := Storage{db: db}

			req := storage.SaveWorkers{
				RootProductID: 100,
				UpdateStatus:  "assigned",
				Assignments: []storage.OperationWorkers{
					{
						ProductID:     101,
						OperationName: "Сборка",
						EmployeeID:    5,
						ActualMinutes: 60,
						ActualValue:   10,
					},
				},
			}

			tt.setupMock(mock, req)

			err = stor.SaveOperationWorkers(context.Background(), req)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantError)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateStatusTx_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		rootID int64 = 100
		status       = "assigned"
	)

	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE dem_product_instances_al SET status = ? WHERE id = ? OR parent_product_id = ?")).
		WithArgs(
			status,
			rootID,
			rootID,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectRollback()

	err = stor.UpdateStatusTx(context.Background(), tx, rootID, status)
	require.NoError(t, err)

	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateStatusTx_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		rootID int64 = 100
		status       = "assigned"
	)

	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE dem_product_instances_al SET status = ? WHERE id = ? OR parent_product_id = ?")).
		WithArgs(
			status,
			rootID,
			rootID,
		).
		WillReturnError(errors.New("failed to update status"))

	mock.ExpectRollback()

	err = stor.UpdateStatusTx(context.Background(), tx, rootID, status)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to update status")

	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveReadyDate_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		rootID    int64 = 100
		readyDate       = "2023-01-01"
	)

	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE dem_product_instances_al SET ready_date = ? WHERE id = ? OR parent_product_id = ?")).
		WithArgs(readyDate, rootID, rootID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectRollback()

	err = stor.SaveReadyDate(context.Background(), tx, rootID, readyDate)
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveReadyDate_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	const (
		rootID    int64 = 100
		readyDate       = "2023-01-01"
	)

	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE dem_product_instances_al SET ready_date = ? WHERE id = ? OR parent_product_id = ?")).
		WithArgs(readyDate, rootID, rootID).
		WillReturnError(errors.New("ошибка обновления даты готовности"))

	mock.ExpectRollback()

	err = stor.SaveReadyDate(context.Background(), tx, rootID, readyDate)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ошибка обновления даты готовности")

	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}
