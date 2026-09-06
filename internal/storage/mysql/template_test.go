package mysql

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"vue-golang/internal/storage"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetTemplateByCode_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	var Code string = "2"

	operationsJSON := `[{"name":"razmetka","label":"Разметка","value":1.5}]`
	rulesJSON := `[{"field":"a","operator":">","value":100}]`

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, name, category, operations, systema, izd, profile, rules
		FROM dem_templates_al 
		WHERE code = ? AND is_active = TRUE`)).
		WithArgs(Code).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"code",
				"name",
				"category",
				"operations",
				"systema",
				"izd",
				"profile",
				"rules",
			}).AddRow(
				1,
				"2",
				"Дврь 1п кп45",
				"cut",
				operationsJSON,
				"",
				"1p",
				"",
				rulesJSON,
			),
		)

	result, err := stor.GetTemplateByCode(context.Background(), Code)
	require.NoError(t, err)
	require.NotNil(t, result)

	require.Equal(t, "2", result.Code)
	require.Len(t, result.Operations, 1)
	require.Equal(t, "Разметка", result.Operations[0].Label)

	require.Len(t, result.Rules, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTemplateByCode_Error(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock, code string)
		wantError string
	}{
		{
			name: "Template not found sqlNotFound",
			setupMock: func(mock sqlmock.Sqlmock, code string) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, name, category, operations, systema, izd, profile, rules
					FROM dem_templates_al 
					WHERE code = ? AND is_active = TRUE`)).
					WithArgs(code).
					WillReturnError(sql.ErrNoRows)
			},
			wantError: "шаблон с code",
		},
		{
			name: "Template not found",
			setupMock: func(mock sqlmock.Sqlmock, code string) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, name, category, operations, systema, izd, profile, rules
					FROM dem_templates_al 
					WHERE code = ? AND is_active = TRUE`)).
					WithArgs(code).
					WillReturnError(errors.New("выполнение запроса завершилось ошибкой"))
			},
			wantError: "выполнение запроса завершилось ошибкой",
		},
		{
			name: "Template JSONOperations error",
			setupMock: func(mock sqlmock.Sqlmock, code string) {
				operationsJSON := `[{]`
				rulesJSON := `[{"field":"a","operator":">","value":100}]`

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, name, category, operations, systema, izd, profile, rules
						FROM dem_templates_al 
						WHERE code = ? AND is_active = TRUE`)).
					WithArgs("2").
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "code", "name", "category", "operations", "systema", "izd", "profile", "rules",
					}).AddRow(1, "2", "Дврь 1п кп45", "cut", operationsJSON, "", "1p", "", rulesJSON))
			},
			wantError: "ошибка парсинга JSON операций",
		},
		{
			name: "Template JSONRules error",
			setupMock: func(mock sqlmock.Sqlmock, code string) {
				operationsJSON := `[{"name":"razmetka","label":"Разметка","value":1.5}]`
				rulesJSON := `[{]`

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, name, category, operations, systema, izd, profile, rules
						FROM dem_templates_al 
						WHERE code = ? AND is_active = TRUE`)).
					WithArgs("2").
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "code", "name", "category", "operations", "systema", "izd", "profile", "rules",
					}).AddRow(1, "2", "Дврь 1п кп45", "cut", operationsJSON, "", "1p", "", rulesJSON))
			},
			wantError: "ошибка парсинга JSON правил",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			stor := &Storage{db: db}

			var Code string = "2"

			tt.setupMock(mock, Code)

			result, err := stor.GetTemplateByCode(context.Background(), Code)
			require.Error(t, err)
			require.Nil(t, result)
			require.Contains(t, err.Error(), tt.wantError)

			require.NoError(t, mock.ExpectationsWereMet())

		})
	}
}

func TestGetAllTemplates_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, name, category, systema, izd, profile FROM dem_templates_al WHERE is_active = TRUE`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "code", "name", "category", "systema", "izd", "profile",
		}).AddRow(1, "2", "Дврь 1п кп45", "cut", "", "1p", ""))

	result, err := stor.GetAllTemplates(context.Background())
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "2", result[0].Code)
	require.Equal(t, "Дврь 1п кп45", result[0].Name)
	require.Equal(t, "cut", result[0].Category)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllTemplates_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, code, name, category, systema, izd, profile FROM dem_templates_al WHERE is_active = TRUE`)).
		WillReturnError(errors.New("выполнение запроса завершилось ошибкой"))

	result, err := stor.GetAllTemplates(context.Background())
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "выполнение запроса завершилось ошибкой")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTemplateAdmin_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE dem_templates_al SET code=?, category=?, is_active=?, name=?, profile=?, systema=?, izd=?, operations=?, head_name=? WHERE id=?`)).
		WithArgs("2", "category", true, "name Дверь 2п", "profile", "systema", "izd", "[{}]", "head_name", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	var id int = 1
	req := storage.TemplateAdmin{
		Code:      "2",
		Category:  "category",
		IsActive:  true,
		Name:      "name Дверь 2п",
		Profile:   "profile",
		Systema:   "systema",
		TypeIzd:   "izd",
		Operation: `[{}]`,
		HeadName:  "head_name",
	}

	err = stor.UpdateTemplateAdmin(context.Background(), id, req)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTemplateAdmin_UpdateTemplateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE dem_templates_al SET code=?, category=?, is_active=?, name=?, profile=?, systema=?, izd=?, operations=?, head_name=? WHERE id=?`)).
		WithArgs("2", "category", true, "name Дверь 2п", "profile", "systema", "izd", "[{}]", "head_name", 1).
		WillReturnError(errors.New("ошибка обновления шаблона нормирования"))

	var id int = 1
	req := storage.TemplateAdmin{
		Code:      "2",
		Category:  "category",
		IsActive:  true,
		Name:      "name Дверь 2п",
		Profile:   "profile",
		Systema:   "systema",
		TypeIzd:   "izd",
		Operation: `[{}]`,
		HeadName:  "head_name",
	}

	err = stor.UpdateTemplateAdmin(context.Background(), id, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ошибка обновления шаблона нормирования")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTemplateAdmin_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO dem_templates_al (code, name, category, operations, is_active, systema, izd, profile, head_name, rules) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs("2", "name", "category", "[{}]", true, "systema", "izd", "profile", "head_name", "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := storage.TemplateAdmin{
		Code:      "2",
		Name:      "name",
		Category:  "category",
		Operation: `[{}]`,
		IsActive:  true,
		Systema:   "systema",
		TypeIzd:   "izd",
		Profile:   "profile",
		HeadName:  "head_name",
		Rules:     "",
	}

	err = stor.CreateTemplateAdmin(context.Background(), req)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTemplateAdmin_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	stor := &Storage{db: db}

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO dem_templates_al (code, name, category, operations, is_active, systema, izd, profile, head_name, rules) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs("2", "name", "category", "[{}]", true, "systema", "izd", "profile", "head_name", "").
		WillReturnError(errors.New("Ошибка сохранения шаблона в базу"))

	req := storage.TemplateAdmin{
		Code:      "2",
		Name:      "name",
		Category:  "category",
		Operation: `[{}]`,
		IsActive:  true,
		Systema:   "systema",
		TypeIzd:   "izd",
		Profile:   "profile",
		HeadName:  "head_name",
		Rules:     "",
	}

	err = stor.CreateTemplateAdmin(context.Background(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Ошибка сохранения шаблона в базу")

	require.NoError(t, mock.ExpectationsWereMet())
}
