package generate_excel

import (
	"context"
	"fmt"
	"github.com/xuri/excelize/v2"
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
	// 1. Получаем данные через твой интерфейс/сторедж
	products, employees, err := g.storage.GetPEOProductsByCategory(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("fetch data: %w", err)
	}

	reportType := getReportType(filter.Type)

	for i, pr := range products {
		fmt.Printf("index:%v --- product:%v", i, pr)
	}

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
		// Окна: важнее Профиль и Система
		baseHeaders = []string{"Спецификация", "№ Заказа", "Корп/дил", "Заказчик", "Вид продукции", "Система", "Наименование", "Профиль", "Кол-во", "Площадь", "Н/час",
			"Изготовитель", "Н/руб", "защ. Пленки", "пленка н/р"}
	} else if reportType == "loggia" {
		// Лоджии: важнее Квадратура и Количество
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
			f.SetCellValue(sheet, cellName(10, rowNum), p.Sqr)              // Площадь
			f.SetCellValue(sheet, cellName(11, rowNum), p.TotalTime)        // Н/час
			f.SetCellValue(sheet, cellName(12, rowNum), p.Brigade)          // Н/час
			f.SetCellValue(sheet, cellName(13, rowNum), p.NormMoney)        // Н/час
		} else if reportType == "loggia" {
			// Заполняем 13 колонок для Лоджий
			//"Витраж", "№ Заказа", "Корп/дил", "Заказчик", "Наименование", "Кол-во", "Площадь", "Площадь ", "Н/час", "Изготовитель", "Н/час", "Н/руб", "Разница"}
			f.SetCellValue(sheet, cellName(1, rowNum), p.ParentAssembly) // Витраж (ID/Позиция)
			f.SetCellValue(sheet, cellName(2, rowNum), p.OrderNum)       // № Заказа
			f.SetCellValue(sheet, cellName(3, rowNum), p.CustomerType)   // Корп/дил
			f.SetCellValue(sheet, cellName(4, rowNum), p.Customer)       // Заказчик
			f.SetCellValue(sheet, cellName(5, rowNum), p.TypeIzd)        // Наименование
			f.SetCellValue(sheet, cellName(6, rowNum), p.Count)          // Кол-во
			f.SetCellValue(sheet, cellName(7, rowNum), p.Sqr)            // Площадь
			f.SetCellValue(sheet, cellName(8, rowNum), "-")              // Площадь
			f.SetCellValue(sheet, cellName(9, rowNum), p.TotalTime)
			f.SetCellValue(sheet, cellName(10, rowNum), p.Brigade)
			f.SetCellValue(sheet, cellName(11, rowNum), p.NormMoney)
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
	//f.SetPanes(sheet, `{"freeze":true,"split":false,"x_split":0,"y_split":1,"top_left_cell":"A2"}`)
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
