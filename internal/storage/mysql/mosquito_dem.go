package mysql

import (
	"context"
	"fmt"
	"time"
	"vue-golang/internal/storage"
)

func (s *Storage) MoscuiteDem(ctx context.Context, since time.Time) ([]storage.MosquitoOrderDTO, error) {
	const op = "storage.MoscuiteDem.sql"

	stmt1 := `SELECT o.idorders, o.numorders, o.class_id, FROM_UNIXTIME(o.date) AS order_date FROM dem_orders o WHERE o.ms = '1' AND o.date >= ? ORDER BY o.date ASC;`

	rows, err := s.db.QueryContext(ctx, stmt1, since.Unix())
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех москиток из dem_orders: %w", err)
	}
	defer rows.Close()

	var orders []storage.MosquitoOrderDTO

	for rows.Next() {
		var o storage.MosquitoOrderDTO

		err := rows.Scan(&o.OrderID, &o.OrderNum, &o.ClassID, &o.OrderDate)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строк для получения маскиток : %w", err)
		}

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации строк для получения маскиток: %w", err)
	}

	return orders, nil
}

func (s *Storage) UpsertMoskitOrder(ctx context.Context, order storage.MosquitoOrderDTO, externalID string) (int64, error) {
	const op = "storage.UpsertMoskitOrder"

	stmt := `
		INSERT INTO dem_product_instances_al (
			id_dem_orders, order_num, external_id, source_system,
			type, status, template_code, part_type, total_time,
			created_at, updated_at
		) VALUES (
			?, ?, ?, 'moskit_crm', 'moskit', 'in_production', 
			'moskit_default', 'main', 0, NOW(), NOW()
		)
		ON DUPLICATE KEY UPDATE
			id_dem_orders = VALUES(id_dem_orders),
			order_num = VALUES(order_num),
			status = VALUES(status),
			type = VALUES(type),
			updated_at = NOW()
	`

	result, err := s.db.ExecContext(ctx, stmt,
		order.OrderID, order.OrderNum, externalID,
	)
	if err != nil {
		return 0, fmt.Errorf("%s: ошибка upsert москитки: %w", op, err)
	}

	// Получаем ID записи (работает и для INSERT, и для UPDATE в MySQL)
	instanceID, _ := result.LastInsertId()
	if instanceID == 0 {
		// Если был UPDATE, LastInsertId() может вернуть 0 → делаем SELECT
		err = s.db.QueryRowContext(ctx,
			"SELECT id FROM dem_product_instances_al WHERE external_id = ? AND source_system = ?",
			externalID, "moskit_crm",
		).Scan(&instanceID)
		if err != nil {
			return 0, fmt.Errorf("%s: failed to get instance id: %w", op, err)
		}
	}

	return instanceID, nil
}
