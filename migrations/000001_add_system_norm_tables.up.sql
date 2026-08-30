CREATE TABLE `dem_product_instances_al` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `order_num` varchar(32) NOT NULL,
    `template_code` varchar(100) NOT NULL,
    `name` varchar(255) DEFAULT NULL,
    `count` int DEFAULT NULL,
    `total_time` float NOT NULL,
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `type` varchar(25) DEFAULT NULL,
    `parent_assembly` varchar(50) DEFAULT NULL,
    `part_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
    `parent_product_id` bigint DEFAULT NULL,
    `position` int DEFAULT NULL,
    `customer` varchar(50) DEFAULT NULL,
    `status` varchar(20) DEFAULT NULL,
    `systema` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `type_izd` varchar(50) DEFAULT NULL,
    `profile` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
    `sqr` float DEFAULT NULL,
    `brigade` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '',
    `norm_money` float DEFAULT '0',
    `customer_type` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '',
    `ready_date` date DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_order_num` (`order_num`),
    KEY `idx_template_code` (`template_code`),
    KEY `idx_created` (`created_at`),
    KEY `idx_parent_product` (`parent_product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `dem_operation_values_al` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `product_id` bigint NOT NULL,
    `operation_name` varchar(100) NOT NULL,
    `operation_label` varchar(255) NOT NULL,
    `count` float NOT NULL,
    `value` float NOT NULL,
    `minutes` float NOT NULL,
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
    `sort_operation` int DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_op_per_product` (`product_id`,`operation_name`),
    KEY `idx_operation_name` (`operation_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `dem_employees_al` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `name` varchar(255) NOT NULL,
    `is_active` tinyint(1) DEFAULT '1',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=24 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO `dem_employees_al` (`id`, `name`, `is_active`) VALUES
    (1, 'Доронин М.А', 1);
INSERT INTO `dem_employees_al` (`id`, `name`, `is_active`) VALUES
    (2, 'Попов А.В', 1);
INSERT INTO `dem_employees_al` (`id`, `name`, `is_active`) VALUES
    (3, 'Сувориков А.В', 1);
INSERT INTO `dem_employees_al` (`id`, `name`, `is_active`) VALUES
    (4, 'Перебоец А.И', 1),
    (5, 'Голиков А.И', 1),
    (6, 'Столяров Н.Н', 1),
    (7, 'Устюгов И.В', 1),
    (8, 'Фомиченко А.А', 1),
    (9, 'Белов П.О', 1),
    (10, 'Хамроев Б', 1),
    (11, 'Нестеренко', 1),
    (12, 'Чиркин', 1),
    (13, 'Боричев', 1),
    (14, 'Белов', 1),
    (15, 'Котов', 1),
    (16, 'Яшин', 1),
    (17, 'Клусович', 1),
    (18, 'Фролов', 1),
    (19, 'Боричев', 1),
    (20, 'Лузоненко', 1),
    (21, 'Ткаченко', 1),
    (22, 'Карташов', 1),
    (23, 'Филимонов', 1);

CREATE TABLE `dem_operation_executors_al` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `product_id` bigint NOT NULL,
    `operation_name` varchar(100) NOT NULL,
    `employee_id` bigint NOT NULL,
    `actual_minutes` float NOT NULL,
    `notes` text,
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `actual_value` float DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_assignment` (`product_id`,`operation_name`,`employee_id`),
    KEY `idx_product_op` (`product_id`,`operation_name`),
    KEY `idx_employee` (`employee_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE `dem_customer_al` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `name` varchar(50) DEFAULT NULL,
    `short_name_customer` varchar(10) DEFAULT NULL,
    UNIQUE KEY `id` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=123 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


INSERT INTO `dem_customer_al` (`id`, `name`, `short_name_customer`) VALUES
    (1, 'Бурдугуз', 'ч');
INSERT INTO `dem_customer_al` (`id`, `name`, `short_name_customer`) VALUES
    (2, 'Щеголев', 'д');
INSERT INTO `dem_customer_al` (`id`, `name`, `short_name_customer`) VALUES
    (3, 'ИрАэро', 'ч');
INSERT INTO `dem_customer_al` (`id`, `name`, `short_name_customer`) VALUES
                                                                        (4, 'Оконный сервис', 'д'),
                                                                        (6, 'МонтажТехСтройИркутск', 'д'),
                                                                        (7, 'Крайнов', 'д'),
                                                                        (8, 'Нагаев', 'д'),
                                                                        (9, 'Стимул СК', 'д'),
                                                                        (11, 'Забелин', 'ч'),
                                                                        (12, 'Демикс', 'д'),
                                                                        (13, 'Жарких', 'д'),
                                                                        (14, 'Паздников', 'д'),
                                                                        (15, 'Носоченко', 'д'),
                                                                        (16, 'Стройсистема', 'д'),
                                                                        (17, 'Колесникова', 'ч'),
                                                                        (18, 'Атмосфера СК', 'ч'),
                                                                        (21, 'внутренний', 'к'),
                                                                        (22, 'Кондратенко', 'д'),
                                                                        (23, 'Ника', 'д'),
                                                                        (24, 'Зубкова', 'д'),
                                                                        (25, 'Сотрудник Деметра', 'д'),
                                                                        (26, 'Шкурченко', 'д'),
                                                                        (27, 'Сменов', 'д'),
                                                                        (28, 'Шмигорский', 'д'),
                                                                        (30, 'Петюнов', 'д'),
                                                                        (31, 'Федоров', 'д'),
                                                                        (32, 'Балконыч', 'д'),
                                                                        (33, 'Кислицын', 'ч'),
                                                                        (34, 'Богданова', 'ч'),
                                                                        (35, 'Кожухов', 'д'),
                                                                        (36, 'СУ 38', 'к'),
                                                                        (37, 'Кичигин', 'ч'),
                                                                        (38, 'Восток', 'д'),
                                                                        (39, 'ЖК \"Авиатор\"', 'к'),
                                                                        (40, 'Москвитина', 'ч'),
                                                                        (41, 'Политова', 'д'),
                                                                        (42, 'Ляшенко', 'д'),
                                                                        (44, 'Град', 'ч'),
                                                                        (45, 'Алексеев', 'д'),
                                                                        (46, 'СИК \"Развитие\"', 'д'),
                                                                        (47, 'Макарчев', 'ч'),
                                                                        (48, 'Номоконов', 'ч'),
                                                                        (49, 'Эверест', 'д'),
                                                                        (50, 'Пронькин', 'д'),
                                                                        (51, 'РегионСтрой', 'д'),
                                                                        (52, 'СТК', 'д'),
                                                                        (53, 'ТигРА', 'д'),
                                                                        (55, 'Антар', 'д'),
                                                                        (56, 'ТехноСибБайкал', 'ч'),
                                                                        (57, 'Гогнадзе', 'ч'),
                                                                        (58, 'Тирских', 'д'),
                                                                        (59, 'Машуков', 'ч'),
                                                                        (60, 'Онысько', 'д'),
                                                                        (61, 'Ринчинова', 'д'),
                                                                        (62, 'Самарина', 'д'),
                                                                        (63, 'Алабужева', 'ч'),
                                                                        (64, 'Никифоров', 'ч'),
                                                                        (65, 'Бельская', 'ч'),
                                                                        (66, 'Скаллер', 'ч'),
                                                                        (67, 'Каменский', 'д'),
                                                                        (68, 'Остахов', 'ч'),
                                                                        (69, 'Хамелеон', 'ч'),
                                                                        (70, 'Рэдван', 'д'),
                                                                        (72, 'Афанасьев', 'ч'),
                                                                        (74, 'ЖК на Грязнова', 'к'),
                                                                        (75, 'Хабибулина', 'д'),
                                                                        (76, 'Аврора', 'д'),
                                                                        (77, 'Азимут', 'д'),
                                                                        (79, 'Деметра Ангарск', 'д'),
                                                                        (80, 'Тайфун', 'ч'),
                                                                        (81, 'Михалев', 'д'),
                                                                        (82, 'Китов', 'д'),
                                                                        (84, 'ЖК \"Онегин\"', 'к'),
                                                                        (86, 'Гомбоев', 'д'),
                                                                        (87, 'ЖК \"Подсолнух\"', 'к'),
                                                                        (88, 'Дашинимаева', 'д'),
                                                                        (90, 'Сидоркевич', 'д'),
                                                                        (91, 'Приоритет', 'д'),
                                                                        (94, 'РПТ Групп', 'ч'),
                                                                        (97, 'Исаков', 'д'),
                                                                        (100, 'Субботин', 'д'),
                                                                        (101, 'Дугарова', 'д'),
                                                                        (103, 'Мансуров', 'д'),
                                                                        (104, 'Лесюта', 'д'),
                                                                        (106, 'Саргина', 'д'),
                                                                        (107, 'Деметра ТО+', 'д'),
                                                                        (108, 'Сергеев', 'д'),
                                                                        (111, 'СВК', 'д'),
                                                                        (113, 'Вершинин', 'д'),
                                                                        (118, 'Алькор', 'д'),
                                                                        (120, 'Аюшеев', 'д');

ALTER TABLE `dem_product_instances_al` ADD CONSTRAINT `fk_parent_product` FOREIGN KEY (`parent_product_id`) REFERENCES `dem_product_instances_al` (`id`) ON DELETE SET NULL;

ALTER TABLE `dem_operation_values_al` ADD CONSTRAINT `dem_operation_values_al_ibfk_1` FOREIGN KEY (`product_id`) REFERENCES `dem_product_instances_al` (`id`) ON DELETE CASCADE;

ALTER TABLE  `dem_operation_executors_al` ADD CONSTRAINT `dem_operation_executors_al_ibfk_1` FOREIGN KEY (`product_id`) REFERENCES `dem_product_instances_al` (`id`) ON DELETE CASCADE;

ALTER TABLE  `dem_operation_executors_al` ADD  CONSTRAINT `dem_operation_executors_al_ibfk_2` FOREIGN KEY (`employee_id`) REFERENCES `dem_employees_al` (`id`);