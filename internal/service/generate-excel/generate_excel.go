package generate_excel

import (
	"context"
	"fmt"
	"github.com/xuri/excelize/v2"
	"math"
	"strings"
	"vue-golang/internal/storage"
	"vue-golang/internal/storage/mysql"
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
	} //else {
	//baseHeaders = []string{"ID", "Заказ", "Клиент", "Тип", "Дата", "Сумма"}
	//}

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
			f.SetCellValue(sheet, cellName(1, rowNum), p.ParentAssembly) // Витраж (ID/Позиция)
			f.SetCellValue(sheet, cellName(2, rowNum), p.OrderNum)       // № Заказа
			f.SetCellValue(sheet, cellName(3, rowNum), p.CustomerType)   // Корп/дил
			f.SetCellValue(sheet, cellName(4, rowNum), p.Customer)       // Заказчик
			f.SetCellValue(sheet, cellName(5, rowNum), p.TypeIzd)        // Наименование
			f.SetCellValue(sheet, cellName(6, rowNum), p.Count)          // Кол-во
			f.SetCellValue(sheet, cellName(7, rowNum), round(p.Sqr))     // Площадь
			f.SetCellValue(sheet, cellName(8, rowNum), "-")              // Площадь створки для лоджии(пока пусто)
			f.SetCellValue(sheet, cellName(9, rowNum), round(p.TotalTime))
			f.SetCellValue(sheet, cellName(10, rowNum), p.Brigade)
			f.SetCellValue(sheet, cellName(11, rowNum), round(p.NormMoney))
		}

		// 4. Сотрудники (всегда СРАБОТАЕТ ПРАВИЛЬНО благодаря empColMap)
		for empID, val := range p.EmployeeValue {
			if colIdx, ok := empColMap[empID]; ok {
				cell, _ := excelize.CoordinatesToCellName(colIdx, rowNum)
				f.SetCellValue(sheet, cell, val)
			}
		}
	}

	// --- ФИНАЛЬНЫЕ ШТРИХИ ---
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

	var allStats []StatsRow

	if reportType == "window" {
		allStats = append(allStats, winStats...)
		allStats = append(allStats, doorStats...)
	} else if reportType == "loggia" {
		//allStats = append(allStats, loggiaStats...)
	}

	startRowStats := len(products) + 10
	f.SetCellValue(sheet, cellName(1, startRowStats), "Сводная статистика")

	// 1. Создаем стиль для шапки статистики (серый фон, жирный)
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
		case "loggia", "vitrage":
			return "loggia"
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
	default:
		return ""
	}
}

func round(num float64) float64 {
	return math.Round(num*1000) / 1000
}

//TODO суммарная статистика по заказам

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

// Вспомогательная функция, чтобы не дублировать код прибавления цифр
func addStats(row *StatsRow, p storage.PEOProduct) {
	row.Count += p.Count
	row.Sqr += p.Sqr
	row.Hours += p.TotalTime
	row.Money += p.NormMoney
}
