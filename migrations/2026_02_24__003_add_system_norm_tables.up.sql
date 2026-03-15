-- 1. Добавление столбца sqr_stv
-- ВНИМАНИЕ: IF NOT EXISTS удалено, так как MySQL это не поддерживает.
-- При повторном запуске на базе, где столбец уже есть, будет ошибка.
ALTER TABLE `dem_product_instances_al`
    ADD COLUMN `sqr_stv` DOUBLE DEFAULT NULL COMMENT 'Площадь (float64)';

-- 2. Создание таблицы dem_teams_al
CREATE TABLE IF NOT EXISTS `dem_teams_al` (
                                              `id` int NOT NULL AUTO_INCREMENT,
                                              `name` varchar(100) NOT NULL,
    `slug` varchar(50) NOT NULL,
    `is_active` tinyint(1) DEFAULT '1',
    PRIMARY KEY (`id`),
    UNIQUE KEY `slug` (`slug`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 3. Создание таблицы dem_employee_teams_al
CREATE TABLE IF NOT EXISTS `dem_employee_teams_al` (
                                                       `id` bigint NOT NULL AUTO_INCREMENT,
                                                       `employee_id` bigint NOT NULL,
                                                       `team_id` int NOT NULL,
                                                       `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                                                       PRIMARY KEY (`id`),
    UNIQUE KEY `unique_assignment` (`employee_id`,`team_id`),
    KEY `idx_employee` (`employee_id`),
    KEY `idx_team` (`team_id`),
    CONSTRAINT `fk_et_employee` FOREIGN KEY (`employee_id`) REFERENCES `dem_employees_al` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_et_team` FOREIGN KEY (`team_id`) REFERENCES `dem_teams_al` (`id`) ON DELETE CASCADE
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 4. Заполнение данными
INSERT INTO `dem_teams_al` (`id`, `name`, `slug`, `is_active`) VALUES
                                                                   (1, 'Окна и двери', 'windows', 1),
                                                                   (2, 'Витражи и лоджии', 'vitrages', 1)
    ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

ALTER TABLE `dem_teams_al` AUTO_INCREMENT = 3;

INSERT INTO `dem_employee_teams_al` (`id`, `employee_id`, `team_id`, `created_at`) VALUES
                                                                                       (1, 1, 1, '2026-02-22 12:49:06'),
                                                                                       (2, 2, 1, '2026-02-22 12:49:06'),
                                                                                       (3, 3, 1, '2026-02-22 12:49:06'),
                                                                                       (4, 4, 1, '2026-02-22 12:49:06'),
                                                                                       (5, 5, 1, '2026-02-22 12:49:06'),
                                                                                       (6, 6, 1, '2026-02-22 12:49:06'),
                                                                                       (7, 7, 1, '2026-02-22 12:49:06'),
                                                                                       (8, 8, 1, '2026-02-22 12:49:06'),
                                                                                       (10, 10, 1, '2026-02-22 12:49:06'),
                                                                                       (11, 11, 2, '2026-02-22 12:49:06'),
                                                                                       (12, 12, 2, '2026-02-22 12:49:06'),
                                                                                       (14, 14, 1, '2026-02-22 12:49:06'),
                                                                                       (15, 15, 1, '2026-02-22 12:49:06'),
                                                                                       (16, 16, 1, '2026-02-22 12:49:06'),
                                                                                       (17, 17, 1, '2026-02-22 12:49:06'),
                                                                                       (18, 18, 1, '2026-02-22 12:49:06'),
                                                                                       (19, 19, 2, '2026-02-22 12:49:06'),
                                                                                       (20, 20, 1, '2026-02-22 12:49:06'),
                                                                                       (21, 21, 1, '2026-02-22 12:49:06'),
                                                                                       (22, 22, 1, '2026-02-22 12:49:06'),
                                                                                       (23, 23, 1, '2026-02-22 12:49:06'),
                                                                                       (44, 8, 2, '2026-02-22 12:49:06'),
                                                                                       (45, 4, 2, '2026-02-22 12:49:06'),
                                                                                       (46, 2, 2, '2026-02-22 12:49:06'),
                                                                                       (47, 1, 2, '2026-02-22 12:49:06'),
                                                                                       (48, 3, 2, '2026-02-22 12:49:06')
    ON DUPLICATE KEY UPDATE `created_at` = VALUES(`created_at`);

ALTER TABLE `dem_employee_teams_al` AUTO_INCREMENT = 49;