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

func TestGetOrdersMonth_ByMonth_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	stor := &Storage{db: db}

	start := time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC).Unix()
	end := time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC).Unix()

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, order_num, creator, customer, dop_info, ms_note 
			FROM dem_ready 
			WHERE CAST(creation_date AS UNSIGNED) >= ? 
			AND CAST(creation_date AS UNSIGNED) < ?
			AND (order_num LIKE '%Q6%' OR order_num LIKE '%R6-%')
		`)).
		WithArgs(start, end).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_num", "creator", "customer", "dop_info", "ms_note"}).AddRow(100, "Q6-123", 12, "customer", "dop_info", "ms_note"))

	result, err := stor.GetOrdersMonth(context.Background(), 2023, 10, "")
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "Q6-123", result[0].OrderNum)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrdersMonth_BySearch_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	stor := &Storage{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, order_num, creator, customer, dop_info, ms_note 
			FROM dem_ready 
			WHERE order_num LIKE ?
			AND (order_num LIKE '%Q6%' OR order_num LIKE '%R6-%')
		`)).
		WithArgs("%Q6-123%").
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_num", "creator", "customer", "dop_info", "ms_note"}).AddRow(100, "Q6-123", 12, "customer", "dop_info", "ms_note"))

	result, err := stor.GetOrdersMonth(context.Background(), 2023, 10, "Q6-123")
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "Q6-123", result[0].OrderNum)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrdersMonth_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	stor := &Storage{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, order_num, creator, customer, dop_info, ms_note 
			FROM dem_ready 
			WHERE order_num LIKE ?
			AND (order_num LIKE '%Q6%' OR order_num LIKE '%R6-%')
		`)).
		WithArgs("%Q6-123%").
		WillReturnError(errors.New("query error"))

	result, err := stor.GetOrdersMonth(context.Background(), 2023, 10, "Q6-123")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "query error")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderDetails_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	stor := &Storage{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT ANY_VALUE(p.id_) AS id, ANY_VALUE(t.text_type) AS text_type, p.x, 
                    ANY_VALUE(r.order_num) AS order_num, SUM(p.sqr) AS sqr, ANY_VALUE(p.note) AS note,
                    SUM(p.icount) AS icount, ANY_VALUE(p.color) AS color, ANY_VALUE(i.im_image) AS im_image, ANY_VALUE(r.customer) AS customer 
             FROM dem_plan p 
             LEFT JOIN dem_ready r ON r.id = p.idorder 
             LEFT JOIN dem_images i ON i.im_ordername = r.order_num AND i.im_orderpos = p.x
             LEFT JOIN dem_types t ON p.type = t.id_
             WHERE r.order_num = ? AND p.type NOT IN (17, 18) 
             GROUP BY p.x, p.type`)).
		WithArgs("Q6-123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "text_type", "x", "order_num", "sqr", "note", "icount", "color", "im_image", "customer"}).
			AddRow(100, "text_type", 1, "Q6-123", 0.555, "note", 2, "color", "im_image", "customer"))

	result, err := stor.GetOrderDetails(context.Background(), "Q6-123")
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "Q6-123", result[0].OrderNum)
	require.Equal(t, 0.555, result[0].Sqr, 0.000001)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderDetails_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	stor := &Storage{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT ANY_VALUE(p.id_) AS id, ANY_VALUE(t.text_type) AS text_type, p.x, 
                    ANY_VALUE(r.order_num) AS order_num, SUM(p.sqr) AS sqr, ANY_VALUE(p.note) AS note,
                    SUM(p.icount) AS icount, ANY_VALUE(p.color) AS color, ANY_VALUE(i.im_image) AS im_image, ANY_VALUE(r.customer) AS customer 
             FROM dem_plan p 
             LEFT JOIN dem_ready r ON r.id = p.idorder 
             LEFT JOIN dem_images i ON i.im_ordername = r.order_num AND i.im_orderpos = p.x
             LEFT JOIN dem_types t ON p.type = t.id_
             WHERE r.order_num = ? AND p.type NOT IN (17, 18) 
             GROUP BY p.x, p.type`)).
		WithArgs("Q6-123").
		WillReturnError(errors.New("ошибка сканирования строк для получения  деталей заказа"))

	result, err := stor.GetOrderDetails(context.Background(), "Q6-123")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "ошибка сканирования строк для получения  деталей заказа")

	require.NoError(t, mock.ExpectationsWereMet())
}
