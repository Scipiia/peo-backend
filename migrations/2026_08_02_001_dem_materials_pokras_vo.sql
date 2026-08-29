-- Migration: 2026_08_02_001_dem_materials_pokras_vo.sql
-- Назначение: Новые нормы покраски водоотливов, нащельников и оцинковки, деактивация прежних норм.

-- Заменяет поштучную линейную норму 0.243*… на набор технологических операций:
-- поштучные (минуты, переведённые в часы через /60), партийные (ceil(кол-во/6)*N/60)
-- и заказные (ceil(площадь/6|12)*N/60). Работает ТОЛЬКО вместе с доработкой движка:
--   - automatic.php: агрегаты vo_aggregates(), order_level, ceil() в matcalc,
--     проверка условия в заказном блоке, защита пустого mat_if для class 6;
--   - отчёт vo.trud.color.php: FIELD()-сортировка и площади из query_ksp.

UPDATE `dem_materials` SET `active` = 0
WHERE `idmat`=1238 AND `class_id` = 6 AND `type` = 59  AND `active` = 1;

UPDATE `dem_materials` SET `active` = 0
WHERE `idmat`=1239 AND `class_id` = 6 AND `type` = 59  AND `active` = 1;

UPDATE `dem_materials` SET `active` = 0
WHERE `idmat`=1240 AND `class_id` = 6 AND `type` = 59  AND `active` = 1;

UPDATE `dem_materials` SET `active` = 0
WHERE `idmat`=1241 AND `class_id` = 6 AND `type` = 59  AND `active` = 1;

UPDATE `dem_materials` SET `active` = 0
WHERE `idmat`=1242 AND `class_id` = 6 AND `type` = 59  AND `active` = 1;

UPDATE `dem_materials` SET `active` = 0
WHERE `idmat`=1243 AND `class_id` = 6 AND `type` = 59  AND `active` = 1;

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Покраска ВО/нащельник',12,59,'color~0 or color2~0','20/60','ceil(sumsqr_vo/6)*1',0,0,'~',1
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `dem_materials` WHERE `class_id`=6 AND `name_mat`='Покраска ВО/нащельник');

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Подписать',12,59,'color~0 or color2~0','0.6/60','1',0,0,'~',1
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `dem_materials` WHERE `class_id`=6 AND `name_mat`='Подписать');

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Ободрать пленку',12,59,'color~0 or color2~0','0.6/60','1',0,0,'~',1
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `dem_materials` WHERE `class_id`=6 AND `name_mat`='Ободрать пленку');

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Просверлить',12,59,'color~0 or color2~0','0.6/60','1',0,0,'~',1
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `dem_materials` WHERE `class_id`=6 AND `name_mat`='Просверлить');

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Помыть',12,59,'color~0 or color2~0','0.5/60','1',0,0,'~',1
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `dem_materials` WHERE `class_id`=6 AND `name_mat`='Помыть');

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Упаковать',12,59,'color~0 or color2~0','2.5/60','1',0,0,'~',1
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `dem_materials` WHERE `class_id`=6 AND `name_mat`='Упаковать');

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Подготовить крючки',12,59,'color~0 or color2~0','0.5/60','2',0,0,'~',1
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `dem_materials` WHERE `class_id`=6 AND `name_mat`='Подготовить крючки');

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Принести',12,59,'color~0 or color2~0','2/60','ceil(sumkol_all/6)*1',0,0,'~',1
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `dem_materials` WHERE `class_id`=6 AND `name_mat`='Принести');

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Вынос изделия',12,59,'color~0 or color2~0','2/60','ceil(sumkol_all/6)*1',0,0,'~',1
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `dem_materials` WHERE `class_id`=6 AND `name_mat`='Вынос изделия');

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Подвесить (линия)',12,59,'color~0 or color2~0','15/60','ceil(sumsqr_vo_single/6)*1',0,0,'~',1;

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Снять и осмотреть',12,59,'color~0 or color2~0','10/60','ceil(sumsqr_vo_single/6)*1',0,0,'~',1;

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Продувка',12,59,'sumkol_all>0','23/60','1',0,0,'~',1
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `dem_materials` WHERE `class_id`=6 AND `name_mat`='Продувка');

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Снять и осмотреть',12,59,'color~0 or color2~0','10/60','ceil(sumsqr_oc_single/12)*1',0,0,'~',1;

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Подвесить (линия)',12,59,'color~0 or color2~0','15/60','ceil(sumsqr_oc_single/12)*1',0,0,'~',1;

INSERT INTO `dem_materials` (`class_id`,`articul_mat`,`name_mat`,`messureid`,`type`,`mat_if`,`razmer`,`kolvo`, `price`, `cur_id`, `skl`,  `active`)
SELECT 6,'ПОКРАСКА','Покраска оцинковки',12,59,'color~0 or color2~0','20/60','ceil(sumsqr_oc/12)*1',0,0,'~',1
FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM `dem_materials` WHERE `class_id`=6 AND `name_mat`='Покраска оцинковки');