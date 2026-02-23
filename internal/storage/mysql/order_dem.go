package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"vue-golang/internal/storage"
)

func (s *Storage) GetOrdersMonth(ctx context.Context, year int, month int, search string) ([]*storage.Order, error) {
	const op = "storage.order-dem-details.GetOrdersMonth.sql"

	var stmt string
	var args []interface{}

	if search != "" {
		stmt = `
			SELECT id, order_num, creator, customer, dop_info, ms_note 
			FROM dem_ready 
			WHERE order_num LIKE ?
		`
		args = append(args, "%"+search+"%")
	} else {
		// Иначе фильтруем по месяцу
		startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := startOfMonth.AddDate(0, 1, 0)
		startUnix := startOfMonth.Unix()
		endUnix := endOfMonth.Unix()

		stmt = `
			SELECT id, order_num, creator, customer, dop_info, ms_note 
			FROM dem_ready 
			WHERE CAST(creation_date AS UNSIGNED) >= ? 
			  AND CAST(creation_date AS UNSIGNED) < ?
		`
		args = []interface{}{startUnix, endUnix}
	}

	// Дополнительно вытягивать только АЛ заказы
	stmt += " AND (order_num LIKE '%Q6%' OR order_num LIKE '%R6-%')"
	//order_num LIKE '%Q6%' OR order_num LIKE '%R6%')

	rows, err := s.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка получения заказов из дема по месяцам %w", op, err)
	}
	defer rows.Close()

	var orders []*storage.Order
	for rows.Next() {
		var order storage.Order
		var msNote sql.NullString

		err := rows.Scan(&order.ID, &order.OrderNum, &order.Creator, &order.Customer, &order.DopInfo, &msNote)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if msNote.Valid {
			order.MsNote = msNote.String
		} else {
			order.MsNote = ""
		}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка сканирования строк %w", op, err)
	}

	return orders, nil
}

//func (s *Storage) GetOrderDetails1(ctx context.Context, id int) (*storage.ResultOrderDetails1, error) {
//	const op = "storage.order-dem-details.GetOrderDetails.sql"
//
//	details := &storage.ResultOrderDetails1{}
//
//	// основной заказ
//	stmtDemOrders := `
//		SELECT id, order_num, creator, customer, dop_info, ms_note
//		FROM dem_ready
//		WHERE id = ?`
//
//	var msNote sql.NullString
//	details.Order = &storage.Order{}
//	err := s.db.QueryRowContext(ctx, stmtDemOrders, id).Scan(
//		&details.Order.ID,
//		&details.Order.OrderNum,
//		&details.Order.Creator,
//		&details.Order.Customer,
//		&details.Order.DopInfo,
//		&msNote,
//	)
//	if err != nil {
//		if errors.Is(err, sql.ErrNoRows) {
//			return nil, fmt.Errorf("%s: no rows found for order id=%d: %w", op, id, err)
//		}
//		return nil, fmt.Errorf("%s: query failed: %w", op, err)
//	}
//
//	if msNote.Valid {
//		details.Order.MsNote = msNote.String
//	} else {
//		details.Order.MsNote = ""
//	}
//
//	// позиции заказа
//	stmtDemPrice := `
//		SELECT
//			CAST(p.position AS UNSIGNED),
//			p.creator,
//			p.name_position,
//			p.kol_vo,
//			i.im_image,
//			pl.color,
//			COALESCE(SUM(pl.sqr), 0) AS sqr
//		FROM dem_price p
//		LEFT JOIN dem_images i ON i.im_ordername = p.numorders AND i.im_orderpos = p.position
//		LEFT JOIN dem_plan pl ON pl.idorder = ? AND CAST(pl.x AS UNSIGNED) = p.position
//		WHERE p.numorders LIKE ?
//		GROUP BY p.position, p.creator, p.name_position, p.kol_vo, i.im_image, pl.color
//		ORDER BY 1`
//
//	rows, err := s.db.QueryContext(ctx, stmtDemPrice, id, details.Order.OrderNum)
//	if err != nil {
//		return nil, fmt.Errorf("%s: failed to query dem_price: %w", op, err)
//	}
//	defer rows.Close()
//
//	details.OrderDemPrice = []*storage.OrderDemPrice{}
//	for rows.Next() {
//		price := &storage.OrderDemPrice{}
//		err := rows.Scan(
//			&price.Position,
//			&price.Creator,
//			&price.NamePosition,
//			&price.Count,
//			&price.Image,
//			&price.Color,
//			&price.Sqr,
//		)
//		if err != nil {
//			return nil, fmt.Errorf("%s: failed to scan dem_price row: %w", op, err)
//		}
//		details.OrderDemPrice = append(details.OrderDemPrice, price)
//	}
//	if err = rows.Err(); err != nil {
//		return nil, fmt.Errorf("%s: rows iteration error: %w", op, err)
//	}
//
//	return details, nil
//}

func (s *Storage) GetOrderDetails(ctx context.Context, orderNum string) ([]*storage.ResultOrderDetails, error) {
	const op = "storage.order-dem-details.GetOrderDetails.sql"

	stmt := `SELECT ANY_VALUE(p.id_) AS id, ANY_VALUE(t.text_type) AS text_type, p.x, 
                    ANY_VALUE(r.order_num) AS order_num, SUM(p.sqr) AS sqr, ANY_VALUE(p.note) AS note,
                    SUM(p.icount) AS icount, ANY_VALUE(p.color) AS color, ANY_VALUE(i.im_image) AS im_image, ANY_VALUE(r.customer) AS customer 
             FROM dem_plan p 
             LEFT JOIN dem_ready r ON r.id = p.idorder 
             LEFT JOIN dem_images i ON i.im_ordername = r.order_num AND i.im_orderpos = p.x
             LEFT JOIN dem_types t ON p.type = t.id_
             WHERE r.order_num LIKE ? AND p.type NOT IN (17, 18) 
             GROUP BY p.x, p.type`

	rows, err := s.db.QueryContext(ctx, stmt, orderNum)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to query dem_price: %w", op, err)
	}

	defer rows.Close()

	var details []*storage.ResultOrderDetails

	for rows.Next() {
		var detail storage.ResultOrderDetails

		err := rows.Scan(&detail.ID, &detail.NamePosition, &detail.Position, &detail.OrderNum, &detail.Sqr, &detail.Note, &detail.Count, &detail.Color, &detail.Image, &detail.Customer)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования строк для получения деталей заказа %w", op, err)
		}

		details = append(details, &detail)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка сканирования строк для получения  деталей заказа %w", op, err)
	}

	return details, nil
}
