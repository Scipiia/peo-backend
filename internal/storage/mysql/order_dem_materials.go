package mysql

import (
	"context"
	"fmt"
	"vue-golang/internal/storage"
)

func (s *Storage) GetOrderMaterials(ctx context.Context, orderNum string, pos int) ([]*storage.KlaesMaterials, error) {
	const op = "storage.order-dem-materials.GetOrderMaterials.sql"

	//orderNum := "Q6-327732"

	stmtOrderID := `SELECT idorders FROM dem_orders WHERE numorders = ? AND class_id = ?`

	var id int

	err := s.db.QueryRowContext(ctx, stmtOrderID, orderNum, 10).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка выполнения запроса для получения id который нужен для материалов %w", op, err)
	}

	stmt := `SELECT idorders, articul_mat, name_mat, width, height, count, pole, position FROM dem_klaes_materials 
            	WHERE idorders=? AND position=? AND TRIM(name_mat) IN ('импост', 'стойка-импост', 'профиль импостный',
            	'импост в дверь', 'Накладка на цилиндр Stublina', 'Створка Т-образная', 'Створка-коробка', 'Створка Т - образ.',
            	'Петля роликовая RDRH', 'Многозапорный замок Stublina с управлением от ручки', 'Петля роликовая для КП45',
            	'Петля Фурал дверная 2-част. с подшипником', 'Петля дверная трехсекционная с удлиненной базой', 'Притвор КП40',
            	'Петля двухсекционная 67мм', 'Накладка на цилиндр Stublina (под покраску)', 'Замок Elementis 1155 (D30) (для бугельных ручек)',
            	'Замок Elementis 1153 (D30) (под нажимной гарнитур)', 'Штульп', 'Створка оконная', 'Створка оконная усиленная прямоугольная',
            	'Фурнитурная тяга')`

	rows, err := s.db.QueryContext(ctx, stmt, id, pos)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка выполнения запроса для получения материалов %w", op, err)
	}

	defer rows.Close()

	var materials []*storage.KlaesMaterials

	for rows.Next() {
		var material storage.KlaesMaterials

		err = rows.Scan(&material.OrderID, &material.ArticulMat, &material.NameMat, &material.Width, &material.Height, &material.Count, &material.Pole, &material.Position)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования строк материалов %w", op, err)
		}

		materials = append(materials, &material)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка сканирования строк для получения материалов %w", op, err)
	}

	return materials, nil
}

// TODO добавить в параметры позицию для более точного поиска материалов
func (s *Storage) GetDopInfoFromDemPrice(ctx context.Context, orderNum string) ([]*storage.DopInfoDemPrice, error) {
	const op = "storage.order-dem-materials.GetDopInfoFromDemPrice.sql"

	stmt := `SELECT name_position, vendor, pos_k, kol_vo FROM dem_price WHERE numorders LIKE ?`

	rows, err := s.db.QueryContext(ctx, stmt, orderNum)
	if err != nil {
		return nil, fmt.Errorf("%s: ошибка выполнения запроса для получения доп инфы из dem_price %w", op, err)
	}

	defer rows.Close()

	var prices []*storage.DopInfoDemPrice

	for rows.Next() {
		var price storage.DopInfoDemPrice

		err := rows.Scan(&price.NamePosition, &price.ArticulMat, &price.Position, &price.Count)
		if err != nil {
			return nil, fmt.Errorf("%s: ошибка сканирования строк материалов %w", op, err)
		}

		prices = append(prices, &price)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: ошибка сканирования строк для получения доп инфы из dem_price %w", op, err)
	}

	return prices, nil
}
