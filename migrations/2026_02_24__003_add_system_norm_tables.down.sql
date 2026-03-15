
-- 1. Сначала удаляем таблицу с внешними ключами (dem_employee_teams_al)
-- Иначе удаление dem_teams_al упадет с ошибкой FOREIGN KEY
DROP TABLE IF EXISTS `dem_employee_teams_al`;

-- 2. Удаляем таблицу команд
DROP TABLE IF EXISTS `dem_teams_al`;

-- 3. Удаляем добавленный столбец sqr_stv из dem_product_instances_al
ALTER TABLE `dem_product_instances_al` DROP COLUMN `sqr_stv`;