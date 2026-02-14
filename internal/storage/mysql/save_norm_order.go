package mysql

import (
	"context"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"math/rand"
	"vue-golang/internal/storage"
)

func (s *Storage) SaveNormOrder(ctx context.Context, result storage.OrderNormDetails) (int64, error) {
	const op = "storage.mysql.sql.SaveNormOrder"
	stmt := `INSERT INTO dem_product_instances_al (order_num, template_code, name, count, total_time, type, part_type, 
            parent_assembly, parent_product_id, customer, position, status, systema, type_izd, profile, sqr) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?,?,?)`

	exec, err := s.db.ExecContext(ctx, stmt, result.OrderNum, result.TemplateCode, result.Name, result.Count, result.TotalTime,
		result.Type, result.PartType, result.ParentAssembly, result.ParentProductID, result.Customer, result.Position,
		result.Status, result.Systema, result.TypeIzd, result.Profile, result.Sqr)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1452 {
			return 0, fmt.Errorf("%s: Ошибка сохранения нормировки в базу='%s'", op, err)
		}
		return 0, fmt.Errorf("%s: Ошибка сохранения нормировки в базу='%s'", op, err)
	}

	return exec.LastInsertId()
}

func (s *Storage) SaveNormOperation(ctx context.Context, OrderID int64, operations []storage.NormOperation) error {
	const op = "storage.mysql.sql.SaveNormOperation"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: begin transaction: %w", op, err)
	}

	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO dem_operation_values_al 
			(product_id, operation_name, operation_label, count, value, minutes, sort_operation)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    operation_name = VALUES(operation_name),
			count = VALUES(count),
			value = VALUES(value),
			sort_operation = VALUES(sort_operation)
	`)
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	for i, opr := range operations {
		opName := opr.Name
		if opName == "" {
			opName = fmt.Sprintf("manual_%d", rand.Intn(10000))
		}

		_, err := stmt.ExecContext(ctx, OrderID, opName, opr.Label, opr.Count, opr.Value, opr.Minutes, i)
		if err != nil {
			return fmt.Errorf("%s: Ошибка сохранения нормированных операции в базу='%s'", op, err)
		}
	}

	return tx.Commit()
}
