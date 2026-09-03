package generate_excel

import (
	"context"
	"fmt"
	"math"
	"strings"
	"vue-golang/internal/storage"
	"vue-golang/internal/storage/mysql"

	"github.com/xuri/excelize/v2"
)

type GenerateExcelStorage interface {
	GetPEOProductsByCategory(ctx context.Context, filter mysql.ProductFilter) ([]storage.PEOProduct, []storage.GetWorkers, error)
}

type GenerateExcelService struct {
	storage GenerateExcelStorage
}

func NewGenerateService(storage GenerateExcelStorage) *GenerateExcelService {
	return &GenerateExcelService{storage: storage}
}

func (g *GenerateExcelService) GenerateExcel(ctx context.Context, filter mysql.ProductFilter) ([]byte, error) {

	products, employees, err := g.storage.GetPEOProductsByCategory(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("fetch data: %w", err)
	}

	reportType := getReportType(filter.Type)

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Отчет ПЭО"
	f.SetSheetName("Sheet1", sheet)

	// --- СТИЛИ ---
	// Жирный шрифт для шапки
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:   &excelize.Font{Bold: true},
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"E0E0E0"}, Pattern: 1},
		Border: []excelize.Border{{Type: "bottom", Color: "000000", Style: 2}},
	})

	// 2. ФОРМИРУЕМ ШАПКУ
	var baseHeaders []string
	if reportType == "window" {
		baseHeaders = []string{"Спецификация", "№ Заказа", "Корп/дил", "Заказчик", "Вид продукции", "Система", "Наименование", "Профиль", "Кол-во", "Площадь", "Н/час",
			"Изготовитель", "Н/руб", "защ. Пленки", "пленка н/р"}
	} else if reportType == "loggia" {
		baseHeaders = []string{"Витраж", "№ Заказа", "Корп/дил", "Заказчик", "Наименование", "Кол-во", "Площадь", "Площадь створки", "Н/час", "Изготовитель", "Н/час", "Н/руб", "Разница"}
	} else if reportType == "mosquito" {
		baseHeaders = []string{"№ Заказа", "№ Партии", "Заказчик", "Наименование", "Кол-во", "Площадь", "Н/час", "Изготовитель", "Тип клиента", "Вид изделия", "Н/час", "Н/руб", "VSN"}
	} else if reportType == "vodootliv" {
		baseHeaders = []string{"№ Заказа", "Заказчик", "Наименование", "Кол-во", "Площадь", "Н/час", "Н/руб", "Изготовитель", "Тип клиента", "Вид изделия"}
	} else if reportType == "vitrage" {
		baseHeaders = []string{"№ Заказа", "Тип клиента", "Заказчик", "Наименование", "Система", "Категория", "Система", "Кол-во", "Площадь", "Н/час", "Н/руб", "Изготовитель"}
	}

	// 2. Пишем базовую шапку
	for i, name := range baseHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, name)
	}

	// 3. Динамическая шапка сотрудников (начинается СРАЗУ после baseHeaders)
	empColMap := make(map[int64]int)
	baseLen := len(baseHeaders)
	for i, emp := range employees {
		colIdx := baseLen + i + 1
		empColMap[emp.ID] = colIdx
		cell, _ := excelize.CoordinatesToCellName(colIdx, 1)
		f.SetCellValue(sheet, cell, emp.Name)
	}

	// Применяем стиль к шапке
	lastCol, _ := excelize.CoordinatesToCellName(baseLen+len(employees), 1)
	f.SetCellStyle(sheet, "A1", lastCol, headerStyle)

	// 3. ЗАПОЛНЯЕМ ДАННЫЕ
	for rowIdx, p := range products {
		rowNum := rowIdx + 2

		if reportType == "window" {
			// Заполняем 15 колонок для Окон
			f.SetCellValue(sheet, cellName(1, rowNum), p.ParentAssembly)    // Спецификация
			f.SetCellValue(sheet, cellName(2, rowNum), p.OrderNum)          // № Заказа
			f.SetCellValue(sheet, cellName(3, rowNum), p.CustomerType)      // Корп/дил
			f.SetCellValue(sheet, cellName(4, rowNum), p.Customer)          // Заказчик
			f.SetCellValue(sheet, cellName(5, rowNum), convertType(p.Type)) // Вид продукции
			f.SetCellValue(sheet, cellName(6, rowNum), p.Systema)           // Система
			f.SetCellValue(sheet, cellName(7, rowNum), p.TypeIzd)           // Наименование
			f.SetCellValue(sheet, cellName(8, rowNum), p.Profile)           // Профиль
			f.SetCellValue(sheet, cellName(9, rowNum), p.Count)             // Кол-во
			f.SetCellValue(sheet, cellName(10, rowNum), round(p.Sqr))       // Площадь
			f.SetCellValue(sheet, cellName(11, rowNum), round(p.TotalTime)) // Н/час
			f.SetCellValue(sheet, cellName(12, rowNum), p.Brigade)
			f.SetCellValue(sheet, cellName(13, rowNum), round(p.NormMoney))
		} else if reportType == "loggia" {
			// Заполняем 13 колонок для Лоджий
			//"Витраж", "№ Заказа", "Корп/дил", "Заказчик", "Наименование", "Кол-во", "Площадь", "Площадь ", "Н/час", "Изготовитель", "Н/час", "Н/руб", "Разница"}
			f.SetCellValue(sheet, cellName(1, rowNum), p.ParentAssembly)
			f.SetCellValue(sheet, cellName(2, rowNum), p.OrderNum)
			f.SetCellValue(sheet, cellName(3, rowNum), p.CustomerType)
			f.SetCellValue(sheet, cellName(4, rowNum), p.Customer)
			f.SetCellValue(sheet, cellName(5, rowNum), p.TypeIzd)
			f.SetCellValue(sheet, cellName(6, rowNum), p.Count)
			f.SetCellValue(sheet, cellName(7, rowNum), round(p.Sqr))
			f.SetCellValue(sheet, cellName(8, rowNum), "-")
			f.SetCellValue(sheet, cellName(9, rowNum), round(p.TotalTime))
			f.SetCellValue(sheet, cellName(10, rowNum), p.Brigade)
			f.SetCellValue(sheet, cellName(11, rowNum), round(p.NormMoney))
		} else if reportType == "mosquito" {
			f.SetCellValue(sheet, cellName(1, rowNum), p.OrderNum)
			f.SetCellValue(sheet, cellName(2, rowNum), "")
			f.SetCellValue(sheet, cellName(3, rowNum), p.Customer)
			f.SetCellValue(sheet, cellName(4, rowNum), p.Name)
			f.SetCellValue(sheet, cellName(5, rowNum), p.Count)
			f.SetCellValue(sheet, cellName(6, rowNum), p.Sqr)
			f.SetCellValue(sheet, cellName(7, rowNum), round(p.TotalTime))
			f.SetCellValue(sheet, cellName(8, rowNum), "")
			f.SetCellValue(sheet, cellName(9, rowNum), p.CustomerType)
			f.SetCellValue(sheet, cellName(10, rowNum), p.TypeIzd)
			f.SetCellValue(sheet, cellName(11, rowNum), round(p.TotalTime))
			f.SetCellValue(sheet, cellName(12, rowNum), round(p.NormMoney))
			f.SetCellValue(sheet, cellName(13, rowNum), "")
		} else if reportType == "vodootliv" {
			f.SetCellValue(sheet, cellName(1, rowNum), p.OrderNum)
			f.SetCellValue(sheet, cellName(2, rowNum), p.Customer)
			f.SetCellValue(sheet, cellName(3, rowNum), p.Name)
			f.SetCellValue(sheet, cellName(4, rowNum), p.Count)
			f.SetCellValue(sheet, cellName(5, rowNum), p.Sqr)
			f.SetCellValue(sheet, cellName(6, rowNum), round(p.TotalTime))
			f.SetCellValue(sheet, cellName(7, rowNum), round(p.NormMoney))
			f.SetCellValue(sheet, cellName(8, rowNum), "-")
			f.SetCellValue(sheet, cellName(9, rowNum), p.CustomerType)
			f.SetCellValue(sheet, cellName(10, rowNum), p.TypeIzd)
		} else if reportType == "vitrage" {
			f.SetCellValue(sheet, cellName(1, rowNum), p.OrderNum)
			f.SetCellValue(sheet, cellName(2, rowNum), p.CustomerType)
			f.SetCellValue(sheet, cellName(3, rowNum), p.Customer)
			f.SetCellValue(sheet, cellName(4, rowNum), convertType(p.Type))
			f.SetCellValue(sheet, cellName(5, rowNum), p.Systema)
			f.SetCellValue(sheet, cellName(6, rowNum), safeFloat64(p.SqrStv))
			f.SetCellValue(sheet, cellName(7, rowNum), p.TypeIzd)
			f.SetCellValue(sheet, cellName(8, rowNum), p.Count)
			f.SetCellValue(sheet, cellName(9, rowNum), p.Sqr)
			f.SetCellValue(sheet, cellName(10, rowNum), round(p.TotalTime))
			f.SetCellValue(sheet, cellName(11, rowNum), round(p.NormMoney))
			f.SetCellValue(sheet, cellName(12, rowNum), p.Brigade)
		}

		// 4. Сотрудники (всегда СРАБОТАЕТ ПРАВИЛЬНО благодаря empColMap)
		for empID, val := range p.EmployeeValue {
			if colIdx, ok := empColMap[empID]; ok {
				cell, _ := excelize.CoordinatesToCellName(colIdx, rowNum)
				f.SetCellValue(sheet, cell, val)
			}
		}
	}

	// 4. Закрепляем первую строку
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "",
		Selection:   nil,
	})

	// 5. Авто-ширина колонок (базовая реализация)
	f.SetColWidth(sheet, "A", "G", 15)

	winStats := g.getWindowStats(products)
	doorStats := g.getDoorStats(products)
	loggiaStats := g.getLoggiaStats(products)
	mosquitoStats := g.getMosquitoStats(products)
	vodootlivStats := g.getVodootlivStats(products)
	vitrageStats := g.getVitrageStats(products)

	//slog.Info("VODOOTL", vodootlivStats)

	var allStats []StatsRow

	if reportType == "window" {
		allStats = append(allStats, winStats...)
		allStats = append(allStats, doorStats...)
	} else if reportType == "loggia" {
		allStats = append(allStats, loggiaStats...)
	} else if reportType == "mosquito" {
		allStats = append(allStats, mosquitoStats...)
	} else if reportType == "vodootliv" {
		allStats = append(allStats, vodootlivStats...)
	} else if reportType == "vitrage" {
		allStats = append(allStats, vitrageStats...)
	}

	startRowStats := len(products) + 10
	f.SetCellValue(sheet, cellName(1, startRowStats), "Сводная статистика")

	statsHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"CCCCCC"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// 2. Пишем шапку таблицы статистики
	statsHeaders := []string{"Наименование", "Кол-во (шт)", "Площадь (м2)", "Н/час всего", "Н/руб (сумма)"}
	for i, name := range statsHeaders {
		cell := cellName(i+1, startRowStats+1)
		f.SetCellValue(sheet, cell, name)
		f.SetCellStyle(sheet, cell, cell, statsHeaderStyle)
	}

	// 3. Выводим данные из winStats
	for i, row := range allStats {
		currentRow := startRowStats + 2 + i

		f.SetCellValue(sheet, cellName(1, currentRow), row.Label)
		f.SetCellValue(sheet, cellName(2, currentRow), row.Count)
		f.SetCellValue(sheet, cellName(3, currentRow), round(row.Sqr))
		f.SetCellValue(sheet, cellName(4, currentRow), round(row.Hours))
		f.SetCellValue(sheet, cellName(5, currentRow), round(row.Money))

		// Если это последняя строка (ИТОГО), можно сделать её жирной
		if row.Label == "Всего окон" {
			boldStyle, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
			f.SetCellStyle(sheet, cellName(1, currentRow), cellName(5, currentRow), boldStyle)
		}
	}

	empStats := g.getEmployeeStats(products, employees)
	if len(empStats) > 0 {
		startRowEmp := startRowStats + len(allStats) + 5

		f.SetCellValue(sheet, cellName(1, startRowEmp), "Статистика по сотрудникам")

		empHeaders := []string{"Сотрудник", "Н/ч(месяц)"}

		//for week := 1; week <= 5; week++ {
		//	empHeaders = append(empHeaders, fmt.Sprintf("Неделя %d", week))
		//}

		headerRow := startRowEmp + 1
		for i, name := range empHeaders {
			cell := cellName(i+1, headerRow)
			f.SetCellValue(sheet, cell, name)
			f.SetCellStyle(sheet, cell, cell, statsHeaderStyle) // используем тот же стиль, что и для сводной
		}

		for i, emp := range empStats {
			rowNum := headerRow + 1 + i
			f.SetCellValue(sheet, cellName(1, rowNum), emp.Name)
			f.SetCellValue(sheet, cellName(2, rowNum), round(emp.TotalHours))

			// Пример вывода по неделям (если заполняешь WeeklyHours):
			//for week := 1; week <= 4; week++ {
			//	f.SetCellValue(sheet, cellName(3+week, rowNum), round(emp.WeeklyHours[week]))
			//}
		}

		f.SetColWidth(sheet, "A", "C", 20)
	}

	// Генерируем буфер
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

func getReportType(types []string) string {
	for _, t := range types {
		switch t {
		case "window", "door", "glyhar":
			return "window"
		case "loggia":
			return "loggia"
		case "mosquito":
			return "mosquito"
		case "vodootliv":
			return "vodootliv"
		case "vitrage":
			return "vitrage"
		}
	}
	return "window"
}

func convertType(nameType string) string {
	switch nameType {
	case "window":
		return "окно"
	case "door":
		return "дверь"
	case "glyhar":
		return "окно"
	case "vitrage":
		return "витраж"
	default:
		return ""
	}
}

func round(num float64) float64 {
	return math.Round(num*1000) / 1000
}

func safeFloat64(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

// TODO суммарная статистика по заказам
type StatsRow struct {
	Label string  // Название (например, "Холодные окна")
	Count int     // Кол-во
	Sqr   float64 // Площадь
	Hours float64 // Н/час
	Money float64 // Сумма (Н/руб)
}

func (g *GenerateExcelService) getWindowStats(products []storage.PEOProduct) []StatsRow {
	var coldWindow, hotWindow, vitrageDoor, unknown, totalWindow StatsRow

	coldWindow.Label = "Холодные окна"
	hotWindow.Label = "Теплые окна"
	vitrageDoor.Label = "Витраж к двери"
	totalWindow.Label = "Всего окон"

	unknown.Label = "Неизвестное изделие(окна)"

	for _, p := range products {
		systema := strings.ToLower(p.Systema)
		typeIzd := strings.ToLower(p.TypeIzd)
		//fmt.Printf("DEBUG: Type='%s', TypeIzd='%s', Systema='%s'\n", p.Type, p.TypeIzd, p.Systema)

		if p.Type == "window" || p.Type == "glyhar" {
			if typeIzd == "витраж к двери" {
				addStats(&vitrageDoor, p)
			} else if systema == "х" {
				addStats(&coldWindow, p)
			} else if systema == "т" {
				addStats(&hotWindow, p)
			} else {
				addStats(&unknown, p)
			}
		}

	}

	totalWindow.Count = coldWindow.Count + hotWindow.Count + vitrageDoor.Count
	totalWindow.Sqr = coldWindow.Sqr + hotWindow.Sqr + vitrageDoor.Sqr
	totalWindow.Hours = coldWindow.Hours + hotWindow.Hours + vitrageDoor.Hours
	totalWindow.Money = coldWindow.Money + hotWindow.Money + vitrageDoor.Money

	var result []StatsRow

	result = append(result, coldWindow)
	result = append(result, hotWindow)
	result = append(result, vitrageDoor)
	result = append(result, totalWindow)
	result = append(result, unknown)

	return result
}

func (g *GenerateExcelService) getDoorStats(products []storage.PEOProduct) []StatsRow {
	var door1P, door15P, door2P, coldDoor, hotDoor, unknown StatsRow

	door1P.Label = "Всего 1П дверей"
	door15P.Label = "Всего 1.5П дверей"
	door2P.Label = "Всего 2П дверей"

	hotDoor.Label = "Всего теплых дверей"
	coldDoor.Label = "Всего холодных дверей"

	unknown.Label = "Неизвестное изделие(двери)"

	for _, p := range products {
		systema := strings.ToLower(p.Systema)
		typeIzd := strings.ToLower(strings.TrimSpace(p.TypeIzd))

		if p.Type == "door" {
			if typeIzd == "1п" || typeIzd == "1пт" {
				addStats(&door1P, p)
			} else if typeIzd == "1.5п" || typeIzd == "1.5пт" {
				addStats(&door15P, p)
			} else if typeIzd == "2п" || typeIzd == "2пт" {
				addStats(&door2P, p)
			} else {
				addStats(&unknown, p)
			}

			if systema == "х" || systema == "x" {
				addStats(&coldDoor, p)
			} else if systema == "т" {
				addStats(&hotDoor, p)
			}
		}
	}

	var result []StatsRow

	result = append(result, door1P)
	result = append(result, door15P)
	result = append(result, door2P)
	result = append(result, coldDoor)
	result = append(result, hotDoor)
	result = append(result, unknown)

	return result
}

// TODO добавить GetLoggiaStats
func (g *GenerateExcelService) getLoggiaStats(products []storage.PEOProduct) []StatsRow {
	var stv, stvTwo, stvThree, stvFour, stvFive, stvSix, stvAll, unknown StatsRow

	stv.Label = "створка"
	stvTwo.Label = "2ств.лр"
	stvThree.Label = "3ств.лр"
	stvFour.Label = "4ств.лр"
	stvFive.Label = "5ств.лр"
	stvSix.Label = "6ств.лр"
	stvAll.Label = "всего лоджии"
	unknown.Label = "разное"

	for _, p := range products {
		typeIzd := strings.ToLower(strings.TrimSpace(p.TypeIzd))

		if p.Type == "loggia" {
			addStats(&stvAll, p)
			if typeIzd == "створка" {
				addStats(&stv, p)
			} else if typeIzd == "2ств.лр" {
				addStats(&stvTwo, p)
			} else if typeIzd == "3ств.лр" {
				addStats(&stvThree, p)
			} else if typeIzd == "4ств.лр" {
				addStats(&stvFour, p)
			} else if typeIzd == "5ств.лр" {
				addStats(&stvFive, p)
			} else if typeIzd == "6ств.лр" {
				addStats(&stvSix, p)
			} else {
				addStats(&unknown, p)
			}
		}
	}

	var result []StatsRow

	result = append(result, stv)
	result = append(result, stvTwo)
	result = append(result, stvThree)
	result = append(result, stvFour)
	result = append(result, stvFive)
	result = append(result, stvSix)
	result = append(result, stvAll)
	result = append(result, unknown)

	return result
}

func (g *GenerateExcelService) getMosquitoStats(products []storage.PEOProduct) []StatsRow {
	var vsn, ms, combined, totalMs, unknown StatsRow

	vsn.Label = "VSN"
	ms.Label = "Обычная"
	combined.Label = "Смешанные заказы(vsn+ms)"
	totalMs.Label = "Всего москиток"
	unknown.Label = "Неизвестные изделия"

	for _, p := range products {
		typeIzd := strings.ToLower(strings.TrimSpace(p.TypeIzd))

		if p.Type == "mosquito" {
			addStats(&totalMs, p)
			if typeIzd == "vsn" {
				addStats(&vsn, p)
			} else if typeIzd == "ms" {
				addStats(&ms, p)
			} else if strings.Contains(typeIzd, "+") {
				addStats(&combined, p)
			} else {
				addStats(&unknown, p)
			}
		}
	}
	var result []StatsRow

	result = append(result, vsn)
	result = append(result, ms)
	result = append(result, combined)
	result = append(result, totalMs)
	result = append(result, unknown)

	return result
}

func (g *GenerateExcelService) getVodootlivStats(products []storage.PEOProduct) []StatsRow {
	var vo, ocn, combined, totalVo, unknown StatsRow

	vo.Label = "Водоотлив"
	ocn.Label = "Оцинковка"
	combined.Label = "Смешанные заказы(vo+ocn)"
	totalVo.Label = "Всего водоотливов"
	unknown.Label = "Неизвестные изделия"

	for _, p := range products {
		typeIzd := strings.ToLower(strings.TrimSpace(p.TypeIzd))

		if p.Type == "vodootliv" {
			addStats(&totalVo, p)
			if typeIzd == "vo" {
				addStats(&vo, p)
			} else if typeIzd == "ocn" {
				addStats(&ocn, p)
			} else if strings.Contains(typeIzd, "+") {
				addStats(&combined, p)
			} else {
				addStats(&unknown, p)
			}
		}
	}
	var result []StatsRow

	result = append(result, vo)
	result = append(result, ocn)
	result = append(result, combined)
	result = append(result, totalVo)
	result = append(result, unknown)

	return result
}

//
//func (g *GenerateExcelService) getVitrageStats(products []storage.PEOProduct) []StatsRow {
//	var vitrage45Kat1, vitrage45Kat2, vitrage45Kat3, vitrage45Kat4, vitrage74Kat1, vitrage74Kat2, vitrage74Kat3, vitrage74Kat4,
//		vitrageAlutechKat1X, vitrageAlutechKat2X, vitrageAlutechKat3X, vitrageAlutechKat4X, vitrageAlutechKat1T, vitrageAlutechKat2T,
//		vitrageAlutechKat3T, vitrageAlutechKat4T, sum45And74 StatsRow
//
//	vo.Label = "Водоотлив"
//	ocn.Label = "Оцинковка"
//	combined.Label = "Смешанные заказы(vo+ocn)"
//	totalVo.Label = "Всего водоотливов"
//	unknown.Label = "Неизвестные изделия"
//
//	for _, p := range products {
//		typeIzd := strings.ToLower(strings.TrimSpace(p.TypeIzd))
//
//		if p.Type == "vodootliv" {
//			addStats(&totalVo, p)
//			if typeIzd == "vo" {
//				addStats(&vo, p)
//			} else if typeIzd == "ocn" {
//				addStats(&ocn, p)
//			} else if strings.Contains(typeIzd, "+") {
//				addStats(&combined, p)
//			} else {
//				addStats(&unknown, p)
//			}
//		}
//	}
//	var result []StatsRow
//
//	result = append(result, vo)
//	result = append(result, ocn)
//	result = append(result, combined)
//	result = append(result, totalVo)
//	result = append(result, unknown)
//
//	return result
//}

func (g *GenerateExcelService) getVitrageStats(products []storage.PEOProduct) []StatsRow {
	// Создаем мапы для накопления данных по каждой группе
	stats := make(map[string]*StatsRow)

	// Инициализируем все возможные комбинации
	profiles := []string{"45", "74"}
	categories := []string{"1", "2", "3", "4"}

	// Для 45 и 74
	for _, profile := range profiles {
		for _, cat := range categories {
			key := fmt.Sprintf("%s_%s", profile, cat)
			stats[key] = &StatsRow{
				Label: fmt.Sprintf("%s - Категория %s", profile, cat),
			}
		}
	}

	// Для Алютеха (с разделением на Х/Т)
	for _, sys := range []string{"х", "т"} {
		for _, cat := range categories {
			key := fmt.Sprintf("Алютех_%s_%s", sys, cat)
			label := "Алютех"
			if sys == "х" {
				label += " Х"
			} else {
				label += " Т"
			}
			label += fmt.Sprintf(" - Категория %s", cat)
			stats[key] = &StatsRow{Label: label}
		}
	}

	// Обрабатываем продукты
	for _, p := range products {
		if p.Type != "vitrage" {
			continue
		}

		// Определяем категорию
		cat := ""
		if p.SqrStv != nil {
			cat = fmt.Sprintf("%.0f", *p.SqrStv)
		}
		if cat == "" {
			continue // Пропускаем витражи без категории
		}

		// Определяем ключ
		profile := strings.TrimSpace(p.Profile)
		key := ""

		if profile == "45" || profile == "74" {
			key = fmt.Sprintf("%s_%s", profile, cat)
		} else if strings.Contains(strings.ToLower(profile), "алютех") {
			sys := strings.ToLower(strings.TrimSpace(p.Systema))
			if sys == "х" || sys == "т" {
				key = fmt.Sprintf("Алютех_%s_%s", sys, cat)
			}
		}

		if key == "" || stats[key] == nil {
			continue
		}

		// Используем существующую функцию addStats
		addStats(stats[key], p)
	}

	// Формируем результат в нужном порядке
	var result []StatsRow

	// 1. Система 45
	for _, cat := range categories {
		key := fmt.Sprintf("45_%s", cat)
		if stats[key].Count > 0 {
			result = append(result, *stats[key])
		}
	}
	// Итого по 45
	total45 := StatsRow{Label: "ИТОГО по 45"}
	for _, cat := range categories {
		key := fmt.Sprintf("45_%s", cat)
		total45.Count += stats[key].Count
		total45.Sqr += stats[key].Sqr
		total45.Hours += stats[key].Hours
		total45.Money += stats[key].Money
	}
	if total45.Count > 0 {
		result = append(result, total45)
	}

	// 2. Система 74
	for _, cat := range categories {
		key := fmt.Sprintf("74_%s", cat)
		if stats[key].Count > 0 {
			result = append(result, *stats[key])
		}
	}
	// Итого по 74
	total74 := StatsRow{Label: "ИТОГО по 74"}
	for _, cat := range categories {
		key := fmt.Sprintf("74_%s", cat)
		total74.Count += stats[key].Count
		total74.Sqr += stats[key].Sqr
		total74.Hours += stats[key].Hours
		total74.Money += stats[key].Money
	}
	if total74.Count > 0 {
		result = append(result, total74)
	}

	// 3. Сумма 45 + 74
	sum4574 := StatsRow{
		Label: "СУММА 45 + 74",
		Count: total45.Count + total74.Count,
		Sqr:   total45.Sqr + total74.Sqr,
		Hours: total45.Hours + total74.Hours,
		Money: total45.Money + total74.Money,
	}
	if sum4574.Count > 0 {
		result = append(result, sum4574)
	}

	// 4. Алютех Х
	for _, cat := range categories {
		key := fmt.Sprintf("Алютех_х_%s", cat)
		if stats[key].Count > 0 {
			result = append(result, *stats[key])
		}
	}
	totalAluX := StatsRow{Label: "ИТОГО Алютех Х"}
	for _, cat := range categories {
		key := fmt.Sprintf("Алютех_х_%s", cat)
		totalAluX.Count += stats[key].Count
		totalAluX.Sqr += stats[key].Sqr
		totalAluX.Hours += stats[key].Hours
		totalAluX.Money += stats[key].Money
	}
	if totalAluX.Count > 0 {
		result = append(result, totalAluX)
	}

	// 5. Алютех Т
	for _, cat := range categories {
		key := fmt.Sprintf("Алютех_т_%s", cat)
		if stats[key].Count > 0 {
			result = append(result, *stats[key])
		}
	}
	totalAluT := StatsRow{Label: "ИТОГО Алютех Т"}
	for _, cat := range categories {
		key := fmt.Sprintf("Алютех_т_%s", cat)
		totalAluT.Count += stats[key].Count
		totalAluT.Sqr += stats[key].Sqr
		totalAluT.Hours += stats[key].Hours
		totalAluT.Money += stats[key].Money
	}
	if totalAluT.Count > 0 {
		result = append(result, totalAluT)
	}

	// 6. Сумма Алютех (Х + Т)
	sumAlu := StatsRow{
		Label: "СУММА Алютех (Х + Т)",
		Count: totalAluX.Count + totalAluT.Count,
		Sqr:   totalAluX.Sqr + totalAluT.Sqr,
		Hours: totalAluX.Hours + totalAluT.Hours,
		Money: totalAluX.Money + totalAluT.Money,
	}
	if sumAlu.Count > 0 {
		result = append(result, sumAlu)
	}

	return result
}

// TODO по работникам статистика
type EmpStats struct {
	EmpID       int64
	Name        string
	TotalHours  float64
	TotalMoney  float64
	WeeklyHours map[int]float64
}

func (g *GenerateExcelService) getEmployeeStats(products []storage.PEOProduct, employees []storage.GetWorkers) []EmpStats {
	stats := make(map[int64]*EmpStats)

	for _, emp := range employees {
		stats[emp.ID] = &EmpStats{
			EmpID:       emp.ID,
			Name:        emp.Name,
			WeeklyHours: make(map[int]float64),
		}
	}

	for _, p := range products {
		for empID, value := range p.EmployeeValue {
			if row, ok := stats[empID]; ok {
				row.TotalHours += value
			}
		}
	}

	var result []EmpStats
	for _, row := range stats {
		if row.TotalHours > 0 {
			result = append(result, *row)
		}
	}

	return result
}

// Вспомогательная функция, чтобы не дублировать код прибавления цифр
func addStats(row *StatsRow, p storage.PEOProduct) {
	row.Count += p.Count
	row.Sqr += p.Sqr
	row.Hours += p.TotalTime
	row.Money += p.NormMoney
}
