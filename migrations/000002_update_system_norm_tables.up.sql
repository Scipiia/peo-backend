ALTER TABLE dem_product_instances_al ADD COLUMN coefficient decimal(10,3) DEFAULT NULL;

CREATE TABLE `dem_coefficient_al` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `type` varchar(50) DEFAULT NULL,
    `coefficient` decimal(10,3) DEFAULT NULL,
    `is_active` tinyint(1) DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `dem_coefficient_al` (`id`, `type`, `coefficient`, `is_active`) VALUES
    (1, 'window', 76.870, 1),
    (2, 'door', 76.870, 1),
    (3, 'glyhar', 76.870, 1),
    (4, 'vitrage', 77.770, 0);