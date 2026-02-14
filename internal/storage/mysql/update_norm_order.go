package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"vue-golang/internal/storage"
)

func (s *Storage) UpdateNormOrder(ctx context.Context, ID int64, update storage.UpdateOrderDetails) error {
	const op = "storage.mysql.UpdateNormOrder"

	stmtUpdate := `UPDATE dem_product_instances_al SET total_time = ?, type = ?, status = ? WHERE id = ?`
	stmtDelete := `DELETE FROM dem_operation_values_al WHERE product_id = ?`
	stmtInsert := `INSERT INTO dem_operation_values_al (product_id, operation_name, operation_label, count, value, minutes, sort_operation) VALUES (?, ?, ?, ?, ?, ?, ?)`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: старт транзакции: %w", op, err)
	}
	defer tx.Rollback()

	//Обновляем основное изделие
	_, err = tx.ExecContext(ctx, stmtUpdate, update.TotalTime, update.Type, update.Status, ID)
	if err != nil {
		return fmt.Errorf("%s: ошибка обновление основной информации об изделии: %w", op, err)
	}

	// Удаляем старые операции
	_, err = tx.ExecContext(ctx, stmtDelete, ID)
	if err != nil {
		return fmt.Errorf("%s: ошибка удаления старых операции: %w", op, err)
	}

	// Вставляем новые операции
	prepareInsert, err := tx.PrepareContext(ctx, stmtInsert)
	if err != nil {
		return fmt.Errorf("%s: ошибка при подготовке вставки новых операции: %w", op, err)
	}
	defer prepareInsert.Close()

	for i, operation := range update.Operations {
		opName := operation.Name
		if opName == "" {
			opName = fmt.Sprintf("extra_%d", rand.Intn(10000))
		}

		_, err := prepareInsert.ExecContext(ctx, ID, opName, operation.Label, operation.Count, operation.Value, operation.Minutes, i)
		if err != nil {
			return fmt.Errorf("%s: ошибка вставки новых операции %s: %w", op, opName, err)
		}
	}

	// Коммит
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: ошибка завершения транзакции: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateFinalOrder(ctx context.Context, ID int64, update storage.UpdateFinalOrderDetails) error {
	const op = "storage.mysql.UpdateFinalOrder"

	stmt := `UPDATE dem_product_instances_al SET customer_type = ?, norm_money = ?, profile = ?, sqr = ?, systema = ?, 
            parent_assembly = ?, brigade = ?, type_izd = ?, status = 'final', coefficient = ? WHERE id = ?`

	_, err := s.db.ExecContext(ctx, stmt, update.CustomerType, update.NormMoney, update.Profile, update.Sqr, update.Systema, update.ParentAssembly,
		update.Brigade, update.TypeIzd, update.Coefficient, ID)
	if err != nil {
		return fmt.Errorf("%s: ошибка обновления  %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateStatus(ctx context.Context, rootProductID int64, status string) error {
	const op = "storage.mysql.UpdateStatus"

	stmtUpdateStatus := `UPDATE dem_product_instances_al SET status = ? WHERE id = ? OR parent_product_id = ?`
	stmtDeleteExecutors := `DELETE FROM dem_operation_executors_al WHERE product_id IN (SELECT id FROM dem_product_instances_al WHERE id = ? OR parent_product_id = ?)`

	_, err := s.db.ExecContext(ctx, stmtUpdateStatus, status, rootProductID, rootProductID)
	if err != nil {
		return fmt.Errorf("%s: ошибка обновления статуса root ID %d: %w", op, rootProductID, err)
	}

	_, err = s.db.ExecContext(ctx, stmtDeleteExecutors, rootProductID, rootProductID)
	if err != nil {
		return fmt.Errorf("%s: ошибка удаления назначенных сотрудников заказа с ID %d: %w", op, rootProductID, err)
	}

	return nil
}

func (s *Storage) UpdateStatusTx(ctx context.Context, tx *sql.Tx, rootProductID int64, status string) error {
	const op = "storage.mysql.UpdateStatusTx"

	stmtUpdateStatus := `UPDATE dem_product_instances_al SET status = ? WHERE id = ? OR parent_product_id = ?`

	_, err := tx.ExecContext(ctx, stmtUpdateStatus, status, rootProductID, rootProductID)
	if err != nil {
		return fmt.Errorf("%s: failed to update status in tx for root ID %d: %w", op, rootProductID, err)
	}

	return nil
}
