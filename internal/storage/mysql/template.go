package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"vue-golang/internal/storage"
)

func (s *Storage) GetTemplateByCode(ctx context.Context, code string) (*storage.Template, error) {
	const op = "storage.mysql.sql.GetFormByCode"

	query := `
		SELECT id, code, name, category, operations, systema, izd, profile, rules
		FROM dem_templates_al 
		WHERE code = ? AND is_active = TRUE
	`

	template := &storage.Template{}

	// Сканируем JSON как строку
	var operationsJSON string
	var rulesJSON string
	err := s.db.QueryRowContext(ctx, query, code).Scan(
		&template.ID,
		&template.Code,
		&template.Name,
		&template.Category,
		&operationsJSON,
		&template.Systema,
		&template.TypeIzd,
		&template.Profile,
		&rulesJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: шаблон с code='%s' не найден: %w", op, code, err)
		}
		return nil, fmt.Errorf("%s: выполнение запроса завершилось ошибкой: %w", op, err)
	}

	// Парсим JSON операций
	if err := json.Unmarshal([]byte(operationsJSON), &template.Operations); err != nil {
		return nil, fmt.Errorf("%s: ошибка парсинга JSON операций: %w", op, err)
	}

	// парсим json правила
	if err := json.Unmarshal([]byte(rulesJSON), &template.Rules); err != nil {
		return nil, fmt.Errorf("%s: ошибка парсинга JSON правил: %w", op, err)
	}

	return template, nil
}

func (s *Storage) GetAllTemplates(ctx context.Context) ([]*storage.Template, error) {
	const op = "storage.mysql.sql.GetAllForms"

	stmt := "SELECT id, code, name, category, systema, izd, profile FROM dem_templates_al WHERE is_active = TRUE"

	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var templates []*storage.Template

	for rows.Next() {
		template := &storage.Template{}

		err := rows.Scan(&template.ID, &template.Code, &template.Name, &template.Category, &template.Systema, &template.TypeIzd, &template.Profile)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования строки: %w", op, err)
		}

		templates = append(templates, template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка при итерации по строкам: %w", op, err)
	}

	return templates, nil
}

func (s *Storage) GetTemplateByCodeAdmin(ctx context.Context, code string) (*storage.Template, error) {
	const op = "storage.mysql.sql.GetTemplateByCodeAdmin"

	query := `
		SELECT id, code, name, category, operations, systema, izd, profile, rules, is_active, head_name
		FROM dem_templates_al 
		WHERE code = ?
	`

	template := &storage.Template{}

	// Сканируем JSON как строку
	var operationsJSON string
	var rulesJSON string
	err := s.db.QueryRowContext(ctx, query, code).Scan(
		&template.ID,
		&template.Code,
		&template.Name,
		&template.Category,
		&operationsJSON,
		&template.Systema,
		&template.TypeIzd,
		&template.Profile,
		&rulesJSON,
		&template.IsActive,
		&template.HeadName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: шаблон с code='%s' не найден: %w", op, code, err)
		}
		return nil, fmt.Errorf("%s: выполнение запроса завершилось ошибкой: %w", op, err)
	}

	// Парсим JSON операций
	if err := json.Unmarshal([]byte(operationsJSON), &template.Operations); err != nil {
		return nil, fmt.Errorf("%s: ошибка парсинга JSON операций: %w", op, err)
	}

	// парсим json правила
	if err := json.Unmarshal([]byte(rulesJSON), &template.Rules); err != nil {
		return nil, fmt.Errorf("%s: ошибка парсинга JSON правил: %w", op, err)
	}

	return template, nil
}

func (s *Storage) GetAllTemplatesAdmin(ctx context.Context) ([]*storage.Template, error) {
	const op = "storage.mysql.sql.GetAllTemplatesAdmin"

	stmt := "SELECT id, code, name, category, systema, izd, profile, is_active FROM dem_templates_al"

	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var templates []*storage.Template

	for rows.Next() {
		template := &storage.Template{}

		err := rows.Scan(&template.ID, &template.Code, &template.Name, &template.Category, &template.Systema, &template.TypeIzd, &template.Profile, &template.IsActive)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования строки: %w", op, err)
		}

		templates = append(templates, template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка при итерации по строкам: %w", op, err)
	}

	return templates, nil
}

func (s *Storage) UpdateTemplateAdmin(ctx context.Context, code int, update storage.TemplateAdmin) error {
	const op = "storage.mysql.TemplateAdmin"

	stmt := `UPDATE dem_templates_al SET category=?, is_active=?, name=?, profile=?, systema=?, izd=?, operations=?, head_name=? WHERE code=?`

	_, err := s.db.ExecContext(ctx, stmt, update.Category, update.IsActive, update.Name, update.Profile,
		update.Systema, update.TypeIzd, update.Operation, update.HeadName, code)
	if err != nil {
		return fmt.Errorf("%s: ошибка обновления шаблона нормирования: %w", op, err)
	}

	return nil
}

func (s *Storage) CreateTemplateAdmin(ctx context.Context, res storage.TemplateAdmin) error {
	const op = "storage.mysql.CreateTemplateAdmin"

	stmt := `INSERT INTO dem_templates_al (code, name, category, operations, is_active, systema, 
            izd, profile, head_name, rules) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, stmt, res.Code, res.Name, res.Category, res.Operation,
		res.IsActive, res.Systema, res.TypeIzd, res.Profile, res.HeadName, res.Rules)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1452 {
			return fmt.Errorf("%s: Ошибка сохранения шаблона в базу='%s'", op, err)
		}
		return fmt.Errorf("%s: Ошибка сохранения шаблона в базу='%s'", op, err)
	}

	return nil
}
