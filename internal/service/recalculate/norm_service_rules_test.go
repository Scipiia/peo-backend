package recalculate

import (
	"context"
	"testing"
	"vue-golang/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockNormStorage struct {
	mock.Mock
}

func (m *MockNormStorage) GetOrderMaterials(ctx context.Context, orderNum string, pos int) ([]*storage.KlaesMaterials, error) {
	args := m.Called(ctx, orderNum, pos)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*storage.KlaesMaterials), args.Error(1)
}

func (m *MockNormStorage) GetTemplateByCode(ctx context.Context, code string) (*storage.Template, error) {
	args := m.Called(ctx, code)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Template), args.Error(1)
}

func (m *MockNormStorage) GetDopInfoFromDemPrice(ctx context.Context, orderNum string) ([]*storage.DopInfoDemPrice, error) {
	args := m.Called(ctx, orderNum)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*storage.DopInfoDemPrice), args.Error(1)
}

func TestCalculateNorm(t *testing.T) {
	// Подготавливаем тестовые данные, которые будет возвращать наш мок
	mockMaterials := []*storage.KlaesMaterials{
		{NameMat: "Тестовый материал", Count: 1},
	}
	mockTemplate := &storage.Template{
		Code: "TPL-001",
		Operations: []storage.Operation{
			{Name: "Сборка", Group: "", Value: 10, Minutes: 5, Count: 1},
		},
		Rules: []storage.Rule{}, // Пустые правила для простоты теста
	}
	mockDopInfo := []*storage.DopInfoDemPrice{
		{NamePosition: "Доп. соединитель", Count: 2},
	}

	tests := []struct {
		name              string
		orderNum          string
		pos               int
		typeIzd           string
		templateCode      string
		itemCount         int
		permisDopMaterial bool

		// Ожидаемые результаты
		wantErr      bool
		expectedType string
	}{
		{
			name:              "Успешный расчет с доп. материалами",
			orderNum:          "ORD-123",
			pos:               1,
			typeIzd:           "window",
			templateCode:      "TPL-001",
			itemCount:         2,
			permisDopMaterial: true,
			wantErr:           false,
			expectedType:      "window",
		},
		{
			name:              "Успешный расчет БЕЗ доп. материалов (флаг false)",
			orderNum:          "ORD-123",
			pos:               1,
			typeIzd:           "window",
			templateCode:      "TPL-001",
			itemCount:         2,
			permisDopMaterial: false, // <-- Важный кейс!
			wantErr:           false,
			expectedType:      "window",
		},
		{
			name:              "Ошибка при получении материалов из БД",
			orderNum:          "ORD-ERROR",
			pos:               1,
			typeIzd:           "window",
			templateCode:      "TPL-001",
			itemCount:         1,
			permisDopMaterial: true,
			wantErr:           true, // Ожидаем ошибку
		},
		{
			name:              "Ошибка при получении шаблона из БД",
			orderNum:          "ORD-123",
			pos:               1,
			typeIzd:           "window",
			templateCode:      "TPL-ERROR", // Специальный код для триггера ошибки
			itemCount:         1,
			permisDopMaterial: true,
			wantErr:           true,
		},
		{
			name:              "Ошибка BuildContext: неизвестный тип изделия",
			orderNum:          "ORD-123",
			pos:               1,
			typeIzd:           "unknown_type", // Вызовет ошибку в BuildContext
			templateCode:      "TPL-001",
			itemCount:         1,
			permisDopMaterial: true,
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Создаем мок
			mockStorage := new(MockNormStorage)

			// 2. Настраиваем ожидания (Expectations) УСЛОВНО

			// Ожидаем вызов GetOrderMaterials
			if tt.orderNum == "ORD-ERROR" {
				mockStorage.On("GetOrderMaterials", mock.Anything, tt.orderNum, tt.pos).
					Return(nil, assert.AnError)
			} else {
				mockStorage.On("GetOrderMaterials", mock.Anything, tt.orderNum, tt.pos).
					Return(mockMaterials, nil)
			}

			// ИСПРАВЛЕНИЕ: Ожидаем вызов GetTemplateByCode УСЛОВНО
			if tt.templateCode == "TPL-ERROR" {
				mockStorage.On("GetTemplateByCode", mock.Anything, tt.templateCode).
					Return(nil, assert.AnError)
			} else {
				mockStorage.On("GetTemplateByCode", mock.Anything, tt.templateCode).
					Return(mockTemplate, nil)
			}

			// Ожидаем вызов GetDopInfoFromDemPrice (пока всегда успех)
			mockStorage.On("GetDopInfoFromDemPrice", mock.Anything, tt.orderNum).
				Return(mockDopInfo, nil)

			// 3. Создаем сервис с моком
			service := NewNormService(mockStorage)

			// 4. Вызываем тестируемую функцию
			ops, ctxData, err := service.CalculateNorm(
				context.Background(),
				tt.orderNum, tt.pos, tt.typeIzd, tt.templateCode, tt.itemCount, tt.permisDopMaterial,
			)

			// 5. Проверяем результаты
			if tt.wantErr {
				require.Error(t, err)

				// ИСПРАВЛЕНИЕ: Проверяем текст ошибки в зависимости от того, что сломалось
				if tt.orderNum == "ORD-ERROR" {
					assert.Contains(t, err.Error(), "materials:")
				} else if tt.templateCode == "TPL-ERROR" {
					assert.Contains(t, err.Error(), "template:")
				} else if tt.typeIzd == "unknown_type" {
					assert.Contains(t, err.Error(), "неизвестный тип изделия")
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedType, ctxData.Type)
				assert.Len(t, ops, 1)
				assert.Equal(t, "Сборка", ops[0].Name)

				if tt.itemCount == 2 {
					assert.Equal(t, 20.0, ops[0].Value)
				}
			}

			// 6. КРИТИЧЕСКИ ВАЖНО: Проверяем, что все заявленные методы мока БЫЛИ вызваны
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestBuildContextWindow(t *testing.T) {
	tests := []struct {
		name      string
		materials []*storage.KlaesMaterials
		expected  Context
	}{
		{
			name:      "Пустой список материалов",
			materials: nil,
			expected: Context{
				Type:         "window",
				HasImpost:    false,
				ImpostCount:  0,
				StvTCount600: 0,
				StvTCount400: 0,
				TagCountWin:  0,
			},
		},
		{
			name: "2 импоста",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Импост", Count: 2, Width: 500},
			},
			expected: Context{
				Type:        "window",
				HasImpost:   true,
				ImpostCount: 2,
			},
		},
		{
			name: "Фурнитурная тяги 3 шт",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Фурнитурная тяга", Count: 3, Width: 500},
			},
			expected: Context{
				Type:        "window",
				TagCountWin: 3,
			},
		},
		{
			name: "Фурнитурная тяги 4 шт (суммируются)",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Фурнитурная тяга", Count: 2, Width: 500},
				{NameMat: "Фурнитурная тяга", Count: 2, Width: 500},
			},
			expected: Context{
				Type:        "window",
				TagCountWin: 4,
			},
		},
		{
			name: "Створки до 600мм",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Створка Т-образная", Count: 2, Width: 500},
				{NameMat: "Створка Т-образная", Count: 2, Width: 300},
			},
			expected: Context{
				Type:             "window",
				HasImpost:        false,
				ImpostCount:      0,
				StvCountForOpres: 1,
				StvWindowCount:   4,
				StvTCount600:     4,
				StvTCount400:     2,
			},
		},
		{
			name: "неизвестный материал игнорируется",
			materials: []*storage.KlaesMaterials{
				{NameMat: "неизвестный материал", Count: 99, Width: 500},
			},
			expected: Context{
				Type: "window",
			},
		},
		{
			name: "Разные материалы",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Створка Т-образная", Count: 2, Width: 360},
				{NameMat: "Створка Т-образная", Count: 2, Width: 550},
				{NameMat: "Стойка-импост", Count: 4, Width: 650},
				{NameMat: "Стойка-импост", Count: 4, Width: 650},
				{NameMat: "Фурнитурная тяга", Count: 6, Width: 850},
			},
			expected: Context{
				Type:             "window",
				HasImpost:        true,
				ImpostCount:      8,
				StvTCount600:     4,
				StvTCount400:     2,
				TagCountWin:      6,
				StvCountForOpres: 1,
				StvWindowCount:   4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := BuildContextWindow(tt.materials)
			assert.Equal(t, tt.expected, ctx)
		})
	}
}

func TestBuildContextGlyhar(t *testing.T) {
	tests := []struct {
		name      string
		materials []*storage.KlaesMaterials
		expected  Context
	}{
		{
			name:      "Пустой список материалов",
			materials: nil,
			expected: Context{
				Type:         "glyhar",
				HasImpost:    false,
				ImpostCount:  0,
				StvTCount600: 0,
				StvTCount400: 0,
				TagCountWin:  0,
			},
		},
		{
			name: "2 импоста",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Импост", Count: 2, Width: 500},
			},
			expected: Context{
				Type:        "glyhar",
				HasImpost:   true,
				ImpostCount: 2,
			},
		},
		{
			name: "Створки меньше 600мм",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Створка-коробка", Count: 3, Width: 500},
				{NameMat: "Створка-коробка", Count: 3, Width: 350},
			},
			expected: Context{
				Type:         "glyhar",
				StvTCount600: 6,
				StvTCount400: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := BuildContextGlyhar(tt.materials)
			assert.Equal(t, tt.expected, ctx)
		})
	}
}

func TestBuildContextDoor(t *testing.T) {
	tests := []struct {
		name      string
		materials []*storage.KlaesMaterials
		expected  Context
	}{
		{
			name:      "Пустой список материалов",
			materials: nil,
			expected: Context{
				Type:         "door",
				HasImpost:    false,
				ImpostCount:  0,
				StvTCount600: 0,
				StvTCount400: 0,
				TagCountWin:  0,
			},
		},
		{
			name: "2 импоста",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Импост", Count: 2, Width: 500},
			},
			expected: Context{
				Type:        "door",
				HasImpost:   true,
				ImpostCount: 2,
			},
		},
		{
			name: "Створки меньше 600мм",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Створка-коробка", Count: 3, Width: 500},
				{NameMat: "Створка-коробка", Count: 3, Width: 350},
			},
			expected: Context{
				Type:         "door",
				StvTCount600: 6,
				StvTCount400: 3,
			},
		},
		{
			name: "Накладки стаблина",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Накладка на цилиндр Stublina", Count: 3},
				{NameMat: "Накладка на цилиндр Stublina (под покраску)", Count: 3},
			},
			expected: Context{
				Type:          "door",
				StublinaCount: 6,
			},
		},
		{
			name: "Притвор КП40",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Притвор КП40", Count: 1},
			},
			expected: Context{
				Type:           "door",
				PritvorKP40:    1,
				HasPritvorKP40: true,
			},
		},
		{
			name: "Петли стандарт",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Петля двухсекционная 67мм", Count: 1},
			},
			expected: Context{
				Type:                "door",
				PetliStand:          1,
				PetliForNaveshCount: 1,
			},
		},
		{
			name: "Петли роликовые",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Петля роликовая для КП45", Count: 1},
			},
			expected: Context{
				Type:                "door",
				PetliRolik:          1,
				PetliForNaveshCount: 1,
			},
		},
		{
			name: "Петли 3 секционные",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Петля дверная трехсекционная с удлиненной базой", Count: 1},
			},
			expected: Context{
				Type:          "door",
				Petli3Section: 1,
			},
		},
		{
			name: "Петли фурал",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Петля Фурал дверная 2-част. с подшипником", Count: 1},
			},
			expected: Context{
				Type:                "door",
				PetliFural:          1,
				PetliForNaveshCount: 1,
				HasPetliFural:       true,
			},
		},
		{
			name: "Петли RDRH",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Петля роликовая RDRH", Count: 3},
			},
			expected: Context{
				Type:         "door",
				PetliRDRH:    3,
				HasPetliRDRH: true,
			},
		},
		{
			name: "Многозапорный замок",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Многозапорный замок Stublina с управлением от ручки", Count: 1},
				{NameMat: "Многозапорный замок KFV AS4350 с управлением от ручки", Count: 1},
			},
			expected: Context{
				Type:          "door",
				MnogozapZamok: 2,
			},
		},
		{
			name: "Стандартный замок",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Замок Elementis 1153 (D30) (под нажимной гарнитур)", Count: 1},
				{NameMat: "Замок MACO G-TS 57819(232011)", Count: 1},
			},
			expected: Context{
				Type:       "door",
				StandZamok: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := BuildContextDoor(tt.materials, nil)
			assert.Equal(t, tt.expected, ctx)
		})
	}
}

func TestBuildContextLoggia(t *testing.T) {
	tests := []struct {
		name      string
		materials []*storage.KlaesMaterials
		dopInfo   []*storage.DopInfoDemPrice
		expected  Context
	}{
		{
			name:      "Пустой список материалов",
			materials: nil,
			expected: Context{
				Type:         "loggia",
				HasImpost:    false,
				ImpostCount:  0,
				StvTCount600: 0,
				StvTCount400: 0,
				TagCountWin:  0,
			},
		},
		{
			name: "Количество рам 2шт",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Рама нижняя", Count: 2, Width: 500},
			},
			expected: Context{
				Type:        "loggia",
				LogRamCount: 2,
			},
		},
		{
			name: "Количество створок 6шт",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Створка верх/низ", Count: 3, Width: 500},
				{NameMat: "Створка верх/низ", Count: 3, Width: 350},
			},
			expected: Context{
				Type:        "loggia",
				LogStvCount: 3,
			},
		},
		{
			name: "Соединители 2шт",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Соединитель /сл.60-сл.60/", Count: 2},
			},
			expected: Context{
				Type:         "loggia",
				LogSoedPrice: 2,
			},
		},
		{
			name: "Притворы для ручки 2шт",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Притвор для ручки с защёлкой", Count: 2},
			},
			expected: Context{
				Type:            "loggia",
				LogPritvorPrice: 2,
			},
		},
		{
			name: "Набор вставок 2шт",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Набор вставок", Count: 2},
			},
			expected: Context{
				Type:        "loggia",
				LogKomplVst: 2,
			},
		},
		{
			name: "Соединители из dem_price + dem_klaes_materials 4шт",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Соединитель /сл.60-сл.60/", Count: 2},
			},
			dopInfo: []*storage.DopInfoDemPrice{
				{NamePosition: "Соединитель", Count: 2},
			},
			expected: Context{
				Type:         "loggia",
				LogSoedPrice: 4,
			},
		},
		{
			name: "Притворы из dem_price + dem_klaes_materials 4шт",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Притвор для ручки с защёлкой", Count: 2},
			},
			dopInfo: []*storage.DopInfoDemPrice{
				{NamePosition: "Притвор", Count: 2},
			},
			expected: Context{
				Type:            "loggia",
				LogPritvorPrice: 4,
			},
		},
		{
			name: "Количество створок 2шт",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Створка верх/низ", Count: 4},
			},
			expected: Context{
				Type:        "loggia",
				LogStvCount: 2,
			},
		},
		{
			name: "Подготовка комплектующих 5шт",
			materials: []*storage.KlaesMaterials{
				{NameMat: "Набор вставок", Count: 4.5},
			},
			expected: Context{
				Type:        "loggia",
				LogKomplVst: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := BuildContextLoggia(tt.materials, tt.dopInfo)
			assert.Equal(t, tt.expected, ctx)
		})
	}
}

func TestCompareFloatField(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		expected bool
	}{
		{
			name:     "1.0 == 1.0",
			a:        1.0,
			b:        1.0,
			expected: true,
		},
		{
			name:     "1.0 != 2.0",
			a:        1.0,
			b:        2.0,
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, compareFloatField(tt.a, tt.b))
		})
	}
}

func TestApplyRules(t *testing.T) {
	baseOps := []storage.Operation{
		{Name: "Напиловка", Group: "", Value: 10, Minutes: 10, Count: 1},
		{Name: "Упаковка", Group: "ign", Value: 999, Minutes: 999, Count: 1},
	}

	ctx := Context{ImpostCount: 2}

	tests := []struct {
		name          string
		rules         []storage.Rule
		itemCount     int
		targetOpName  string // <-- ДОБАВИЛИ: имя операции, которую проверяем
		expectedValue float64
	}{
		{
			name: "Set: режим полной замены значения",
			rules: []storage.Rule{
				{Operation: "Напиловка", Mode: "set", SetValue: 100, SetMinutes: 100},
			},
			itemCount:     2,
			targetOpName:  "Напиловка",
			expectedValue: 100,
		},
		{
			name: "Additive: режим добавления значения",
			rules: []storage.Rule{
				{Operation: "Напиловка", Mode: "additive", ValuePerUnit: 10},
			},
			itemCount:     2,
			targetOpName:  "Напиловка",
			expectedValue: 30, // База 10 * 2 (itemCount) = 20, + 10 (additive) = 30
		},
		{
			name: "Multiplied: режим умножения значения",
			rules: []storage.Rule{
				{Operation: "Напиловка", Mode: "multiplied", ValuePerUnit: 10, UnitField: "HasImpostCount"},
			},
			itemCount:     2,
			targetOpName:  "Напиловка",
			expectedValue: 20, // 10 (ValuePerUnit) * 2 (ImpostCount из ctx) = 20
		},
		{
			name: "AdditivePlusMultiplied: режим умножения и сложения",
			rules: []storage.Rule{
				{Operation: "Напиловка", Mode: "additivePlusMultiplied", ValuePerUnit: 10, UnitField: "HasImpostCount"},
			},
			itemCount:     2,
			targetOpName:  "Напиловка",
			expectedValue: 40,
		},
		{
			name: "Minus: режим вычитания",
			rules: []storage.Rule{
				{Operation: "Напиловка", Mode: "minus", ValuePerUnit: 3},
			},
			itemCount:     2,
			targetOpName:  "Напиловка",
			expectedValue: 17,
		},
		{
			name:          "Ign: группа ign НЕ умножается на itemCount",
			rules:         []storage.Rule{},
			itemCount:     3,
			targetOpName:  "Упаковка", // <-- ПРОВЕРЯЕМ ИМЕННО ЕЁ
			expectedValue: 999,        // <-- Значение должно остаться 999, а не 2997
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := make([]storage.Operation, len(baseOps))
			copy(ops, baseOps)

			result := ApplyRules(ops, tt.rules, ctx, tt.itemCount)

			// Ищем именно ту операцию, которую указали в targetOpName
			var targetOp *storage.Operation
			for i := range result {
				if result[i].Name == tt.targetOpName {
					targetOp = &result[i]
					break
				}
			}

			require.NotNil(t, targetOp, "Операция '%s' не найдена в результате", tt.targetOpName)
			assert.InDelta(t, tt.expectedValue, targetOp.Value, 0.001, "Несовпадение значения для операции %s", tt.targetOpName)
		})
	}
}

func TestGetCountMaterials(t *testing.T) {
	ctx := Context{
		ImpostCount:    5.5,
		StvTCount600:   3.0,
		HasPritvorKP40: true, // Это bool, но функция должна вернуть itemCount
		PetliRDRH:      2.0,
	}

	tests := []struct {
		name      string
		field     string
		itemCount int
		expected  float64
	}{
		{"специальное поле itemsCount", "itemsCount", 10, 10.0},
		{"обычное поле из контекста", "HasImpostCount", 10, 5.5},
		{"другое поле из контекста", "StvTCount600", 10, 3.0},
		{"булево поле HasPritvorKP40 возвращает itemCount", "HasPritvorKP40", 7, 7.0},
		{"неизвестное поле возвращает 0", "НеизвестноеПоле", 10, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCountMaterials(tt.field, ctx, tt.itemCount)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}
