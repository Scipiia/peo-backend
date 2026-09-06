package mysql

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetOrderMaterials_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	stor := &Storage{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT idorders FROM dem_orders WHERE numorders = ? AND class_id = ?`)).
		WithArgs("Q6-123", 10).
		WillReturnRows(sqlmock.NewRows([]string{"idorders"}).AddRow(100))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT idorders, articul_mat, name_mat, width, height, count, pole, position FROM dem_klaes_materials 
            	WHERE idorders=? AND position=? AND TRIM(name_mat) IN ('импост', 'стойка-импост', 'профиль импостный',
            	'импост в дверь', 'Накладка на цилиндр Stublina', 'Створка Т-образная', 'Створка-коробка', 'Створка Т - образ.',
            	'Петля роликовая RDRH', 'Многозапорный замок Stublina с управлением от ручки', 'Петля роликовая для КП45',
            	'Петля Фурал дверная 2-част. с подшипником', 'Петля дверная трехсекционная с удлиненной базой', 'Притвор КП40',
            	'Петля двухсекционная 67мм', 'Накладка на цилиндр Stublina (под покраску)', 'Замок Elementis 1155 (D30) (для бугельных ручек)',
            	'Замок Elementis 1153 (D30) (под нажимной гарнитур)', 'Штульп', 'Створка оконная', 'Створка оконная усиленная прямоугольная',
            	'Фурнитурная тяга', 'Многозапорный замок KFV AS4350 с управлением от ручки', 'Многозапорный замок KFV AS2750 с управлением от ключа',
            	'Рама нижняя', 'Створка оконная усиленная', 'Створка верх/низ', 'Ригель облег. двухпол. КП40', 'Соединитель /сл.60-сл.60/', 'Набор вставок',
            	'Петля Фурал дверная 3-част. с подшипником', 'Створка', 'Притвор для ручки с защёлкой', 'Стойка ригель. глухарей', 'Стойка-импост 64мм',
    			'03524590N Фурнитурная тяга', 'Многозапорный замок Fuhr D30 с управлением от ключа', 'Многозапорный замок KFV AS2300 с управлением от ключа',
            	'Створка оконная прямоугольная')`)).
		WithArgs(100, int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"idorders", "articul_mat", "name_mat", "width", "height", "count", "pole", "position"}).AddRow(100, "03524590N", "Фурнитурная тяга", 0.5, 0.5, 1, 0, 1))

	result, err := stor.GetOrderMaterials(context.Background(), "Q6-123", 1)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "03524590N", result[0].ArticulMat)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderMaterials_SetIdError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	stor := &Storage{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT idorders FROM dem_orders WHERE numorders = ? AND class_id = ?`)).
		WithArgs("Q6-123", 10).
		WillReturnError(errors.New("ошибка выполнения запроса для получения id который нужен для материалов"))

	result, err := stor.GetOrderMaterials(context.Background(), "Q6-123", 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ошибка выполнения запроса для получения id который нужен для материалов")
	require.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderMaterials_SetMaterialsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	stor := &Storage{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT idorders FROM dem_orders WHERE numorders = ? AND class_id = ?`)).
		WithArgs("Q6-123", 10).
		WillReturnRows(sqlmock.NewRows([]string{"idorders"}).AddRow(100))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT idorders, articul_mat, name_mat, width, height, count, pole, position FROM dem_klaes_materials 
            	WHERE idorders=? AND position=? AND TRIM(name_mat) IN ('импост', 'стойка-импост', 'профиль импостный',
            	'импост в дверь', 'Накладка на цилиндр Stublina', 'Створка Т-образная', 'Створка-коробка', 'Створка Т - образ.',
            	'Петля роликовая RDRH', 'Многозапорный замок Stublina с управлением от ручки', 'Петля роликовая для КП45',
            	'Петля Фурал дверная 2-част. с подшипником', 'Петля дверная трехсекционная с удлиненной базой', 'Притвор КП40',
            	'Петля двухсекционная 67мм', 'Накладка на цилиндр Stublina (под покраску)', 'Замок Elementis 1155 (D30) (для бугельных ручек)',
            	'Замок Elementis 1153 (D30) (под нажимной гарнитур)', 'Штульп', 'Створка оконная', 'Створка оконная усиленная прямоугольная',
            	'Фурнитурная тяга', 'Многозапорный замок KFV AS4350 с управлением от ручки', 'Многозапорный замок KFV AS2750 с управлением от ключа',
            	'Рама нижняя', 'Створка оконная усиленная', 'Створка верх/низ', 'Ригель облег. двухпол. КП40', 'Соединитель /сл.60-сл.60/', 'Набор вставок',
            	'Петля Фурал дверная 3-част. с подшипником', 'Створка', 'Притвор для ручки с защёлкой', 'Стойка ригель. глухарей', 'Стойка-импост 64мм',
    			'03524590N Фурнитурная тяга', 'Многозапорный замок Fuhr D30 с управлением от ключа', 'Многозапорный замок KFV AS2300 с управлением от ключа',
            	'Створка оконная прямоугольная')`)).
		WithArgs(100, int64(1)).
		WillReturnError(errors.New("ошибка сканирования строк для получения материалов"))

	result, err := stor.GetOrderMaterials(context.Background(), "Q6-123", 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ошибка сканирования строк для получения материалов")
	require.Nil(t, result)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetDopInfoFromDemPrice_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	stor := &Storage{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT name_position, vendor, pos_k, kol_vo FROM dem_price WHERE numorders LIKE ?`)).
		WithArgs("Q6-123").
		WillReturnRows(sqlmock.NewRows([]string{"name_position", "vendor", "pos_k", "kol_vo"}).AddRow("03524590N", "03524590N", 100, 100))

	result, err := stor.GetDopInfoFromDemPrice(context.Background(), "Q6-123")
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "03524590N", result[0].ArticulMat)
	require.Equal(t, float64(100), result[0].Count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetDopInfoFromDemPrice_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	stor := &Storage{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT name_position, vendor, pos_k, kol_vo FROM dem_price WHERE numorders LIKE ?`)).
		WithArgs("Q6-123").
		WillReturnError(errors.New("ошибка сканирования строк для получения доп инфы из dem_price"))

	result, err := stor.GetDopInfoFromDemPrice(context.Background(), "Q6-123")
	require.Error(t, err)
	require.Contains(t, err.Error(), "ошибка сканирования строк для получения доп инфы из dem_price")
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}
