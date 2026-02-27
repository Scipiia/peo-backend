package recalculate

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"
	"vue-golang/internal/storage"
)

type MockNormStorage struct {
	mock.Mock
}

func (m *MockNormStorage) GetOrderMaterials(ctx context.Context, orderNum string, pos int) ([]*storage.KlaesMaterials, error) {
	args := m.Called(ctx, orderNum, pos)

	// Безопасное извлечение: проверяем тип перед приведением
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	materials, ok := args.Get(0).([]*storage.KlaesMaterials)
	if !ok {
		// Если тип не совпадает — возвращаем nil + ошибка
		return nil, fmt.Errorf("expected []*storage.KlaesMaterials, got %T", args.Get(0))
	}

	return materials, args.Error(1)
}

func (m *MockNormStorage) GetTemplateByCode(ctx context.Context, code string) (*storage.Template, error) {
	args := m.Called(ctx, code)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	template, ok := args.Get(0).(*storage.Template)
	if !ok {
		return nil, fmt.Errorf("expected *storage.Template, got %T", args.Get(0))
	}

	return template, args.Error(1)
}

func (m *MockNormStorage) GetDopInfoFromDemPrice(ctx context.Context, orderNum string) ([]*storage.DopInfoDemPrice, error) {
	args := m.Called(ctx, orderNum)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	dopInfo, ok := args.Get(0).([]*storage.DopInfoDemPrice)
	if !ok {
		return nil, fmt.Errorf("expected []*storage.DopInfoDemPrice, got %T", args.Get(0))
	}

	return dopInfo, args.Error(1)
}

func newMaterial(name string, count float64, width float64) *storage.KlaesMaterials {
	return &storage.KlaesMaterials{
		NameMat:    name,
		Width:      width,
		Count:      count,
		Position:   1,
		ArticulMat: "test-art-" + name,
	}
}

func TestCalculateNorm_AdditivePlusMultiplied(t *testing.T) {
	// 1. Создаём мок
	mockStorage := new(MockNormStorage)

	// 2. Подготавливаем фейковые материалы
	materials := []*storage.KlaesMaterials{
		newMaterial("Петля роликовая RDRH", 3.0, 350.0),
		newMaterial("Импост", 1.0, 50.0),
	}

	// 3. Подготавливаем фейковый шаблон
	template := &storage.Template{
		Code: "56",
		Operations: []storage.Operation{
			{Name: "сборка", Group: "", Value: 10.0, Minutes: 30.0, Count: 1.0},
			{Name: "адаптер ПДП 1001-00", Group: "", Value: 0.0, Minutes: 2.0, Count: 1.0},
		},
		Rules: []storage.Rule{
			{
				Operation:      "адаптер ПДП 1001-00",
				Condition:      map[string]interface{}{"HasPetliRDRH": true},
				Mode:           "additivePlusMultiplied",
				UnitField:      "ItemCountForRDRH",
				MinutesPerUnit: 4.5,
			},
		},
	}

	// 4. Настраиваем мок на возврат данных
	mockStorage.On("GetOrderMaterials", mock.Anything, "ORD-123", 1).Return(materials, nil)
	mockStorage.On("GetDopInfoFromDemPrice", mock.Anything, "ORD-123").Return([]*storage.DopInfoDemPrice{}, nil)
	mockStorage.On("GetTemplateByCode", mock.Anything, "56").Return(template, nil)

	// 5. Создаём сервис с моком
	service := NewNormService(mockStorage)

	// 6. Выполняем расчёт
	operations, ctx, err := service.CalculateNorm(context.Background(), "ORD-123", 1, "door", "56", 2)

	// 7. Проверяем результат
	assert.NoError(t, err)
	assert.True(t, ctx.HasPetliRDRH)
	assert.Equal(t, 2, len(operations))

	// Проверяем операцию "адаптер ПДП"
	adapterOp := operations[1]
	assert.Equal(t, "адаптер ПДП 1001-00", adapterOp.Name)
	assert.Equal(t, 13.0, adapterOp.Minutes) // 9 базовое время + 2+2 на адаптер

	// 8. Проверяем вызовы мока
	mockStorage.AssertExpectations(t)
}

func TestCalculateNorm_Multiplied(t *testing.T) {
	mockStorage := new(MockNormStorage)

	materials := []*storage.KlaesMaterials{
		newMaterial("Импост", 2.0, 390.0),
		newMaterial("Петля роликовая для КП45", 3.0, 390.0),
	}

	template := &storage.Template{
		Code: "57",
		Operations: []storage.Operation{
			{Name: "напил импоста", Group: "", Value: 0.0, Minutes: 0.0, Count: 0.0},
			{Name: "сбор петли", Group: "", Value: 0.0, Minutes: 0.0, Count: 0.0},
		},
		Rules: []storage.Rule{
			{
				Operation:      "напил импоста",
				Condition:      map[string]interface{}{"HasImpost": true},
				Mode:           "multiplied",
				UnitField:      "HasImpostCount",
				MinutesPerUnit: 2,
			},
			{
				Operation:      "сбор петли",
				Condition:      map[string]interface{}{"PetliRolik": map[string]interface{}{"min": 1}},
				Mode:           "multiplied",
				UnitField:      "PetliRolik",
				MinutesPerUnit: 4,
			},
		},
	}

	//mockStorage.On("GetOrderMaterials", mock.Anything, "ORD-123", 1).Return(materials, nil)
	//mockStorage.On("GetDopInfoFromDemPrice", mock.Anything, "ORD-123").Return([]*storage.DopInfoDemPrice{}, nil)
	//mockStorage.On("GetTemplateByCode", mock.Anything, "56").Return(template, nil)

	mockStorage.On("GetOrderMaterials", mock.Anything, "ORD-123", 1).Return(materials, nil)
	mockStorage.On("GetDopInfoFromDemPrice", mock.Anything, "ORD-123").Return([]*storage.DopInfoDemPrice{}, nil)
	mockStorage.On("GetTemplateByCode", mock.Anything, "56").Return(template, nil)

	service := NewNormService(mockStorage)

	operation, ctx, err := service.CalculateNorm(context.Background(), "ORD-123", 1, "door", "56", 1)

	assert.NoError(t, err)
	assert.True(t, ctx.HasImpost)
	assert.Equal(t, 3.0, ctx.PetliRolik)
	assert.Equal(t, 2.0, ctx.ImpostCount)
	assert.Equal(t, 2, len(operation))

	operationNapilImp := operation[0]
	assert.Equal(t, "напил импоста", operationNapilImp.Name)
	assert.Equal(t, 4.0, operationNapilImp.Minutes)

	operationSborPtl := operation[1]
	assert.Equal(t, "сбор петли", operationSborPtl.Name)
	assert.Equal(t, 12.0, operationSborPtl.Minutes)
}

func TestCalculateNorm_MaterialsError(t *testing.T) {
	mockStorage := new(MockNormStorage)

	// ✅ Вариант 1: типизированный nil (рекомендуется)
	mockStorage.On("GetOrderMaterials", mock.Anything, "ORD-123", 1).
		Return(([]*storage.KlaesMaterials)(nil), errors.New("база недоступна"))

	// ✅ Вариант 2: можно и просто nil — наш улучшенный мок обработает безопасно
	// mockStorage.On("GetOrderMaterials", mock.Anything, "ORD-123", 1).
	//     Return(nil, errors.New("база недоступна"))

	mockStorage.On("GetDopInfoFromDemPrice", mock.Anything, "ORD-123").
		Return(([]*storage.DopInfoDemPrice)(nil), nil)
	mockStorage.On("GetTemplateByCode", mock.Anything, "TEST").
		Return((*storage.Template)(nil), nil)

	service := NewNormService(mockStorage)
	_, _, err := service.CalculateNorm(context.Background(), "ORD-123", 1, "door", "TEST", 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "materials:") // в твоём текущем коде префикс "materials:"

	mockStorage.AssertExpectations(t)
}
func TestCalculateNorm_ParallelCalls(t *testing.T) {
	mockStorage := new(MockNormStorage)

	// Используем каналы для отслеживания порядка вызовов
	callOrder := []string{}
	mu := sync.Mutex{}

	mockStorage.On("GetOrderMaterials", mock.Anything, "ORD-123", 1).Run(func(args mock.Arguments) {
		mu.Lock()
		callOrder = append(callOrder, "materials")
		mu.Unlock()
		time.Sleep(10 * time.Millisecond) // имитируем задержку
	}).Return([]*storage.KlaesMaterials{newMaterial("рама", 4.0, 600.0)}, nil)

	mockStorage.On("GetTemplateByCode", mock.Anything, "TEST").Run(func(args mock.Arguments) {
		mu.Lock()
		callOrder = append(callOrder, "template")
		mu.Unlock()
		time.Sleep(15 * time.Millisecond)
	}).Return(&storage.Template{Code: "TEST", Operations: []storage.Operation{}}, nil)

	mockStorage.On("GetDopInfoFromDemPrice", mock.Anything, "ORD-123").Run(func(args mock.Arguments) {
		mu.Lock()
		callOrder = append(callOrder, "dop_info")
		mu.Unlock()
		time.Sleep(5 * time.Millisecond)
	}).Return([]*storage.DopInfoDemPrice{}, nil)

	service := NewNormService(mockStorage)
	_, _, err := service.CalculateNorm(context.Background(), "ORD-123", 1, "door", "TEST", 1)

	assert.NoError(t, err)

	// Проверяем, что все три вызова произошли (порядок может быть любым из-за параллелизма)
	assert.Len(t, callOrder, 3)
	assert.Contains(t, callOrder, "materials")
	assert.Contains(t, callOrder, "template")
	assert.Contains(t, callOrder, "dop_info")

	mockStorage.AssertExpectations(t)
}

func TestCompareFloatField(t *testing.T) {
	tests := []struct {
		name     string
		actual   float64
		expected interface{}
		want     bool
	}{
		{
			name:     "точное совпадение",
			actual:   3.0,
			expected: 3.0,
			want:     true,
		},
		{
			name:     "не совпадает",
			actual:   3.0,
			expected: 5.0,
			want:     false,
		},
		{
			name:     "min выполняется",
			actual:   5.0,
			expected: map[string]interface{}{"min": 3.0},
			want:     true,
		},
		{
			name:     "min не выполняется",
			actual:   2.0,
			expected: map[string]interface{}{"min": 3.0},
			want:     false,
		},
		{
			name:     "диапазон min-max выполняется",
			actual:   4.0,
			expected: map[string]interface{}{"min": 3.0, "max": 5.0},
			want:     true,
		},
		{
			name:     "диапазон: меньше min",
			actual:   2.0,
			expected: map[string]interface{}{"min": 3.0, "max": 5.0},
			want:     false,
		},
		{
			name:     "диапазон: больше max",
			actual:   6.0,
			expected: map[string]interface{}{"min": 3.0, "max": 5.0},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareFloatField(tt.actual, tt.expected)
			if got != tt.want {
				t.Errorf("compareFloatField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldMatches(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected interface{}
		ctx      Context
		want     bool
	}{
		{
			name:     "HasImpost: true",
			key:      "HasImpost",
			expected: true,
			ctx:      Context{HasImpost: true},
			want:     true,
		},
		{
			name:     "HasImpost: false",
			key:      "HasImpost",
			expected: true,
			ctx:      Context{HasImpost: false},
			want:     false,
		},
		{
			name:     "HasPetliRDRH: true",
			key:      "HasPetliRDRH",
			expected: true,
			ctx:      Context{HasPetliRDRH: true},
			want:     true,
		},
		{
			name:     "HasPetliRDRH: false",
			key:      "HasPetliRDRH",
			expected: true,
			ctx:      Context{HasPetliRDRH: false},
			want:     false,
		},
		{
			name:     "HasPetliFural: true",
			key:      "HasPetliFural",
			expected: true,
			ctx:      Context{HasPetliFural: true},
			want:     true,
		},
		{
			name:     "HasPetliFural: false",
			key:      "HasPetliFural",
			expected: true,
			ctx:      Context{HasPetliFural: false},
			want:     false,
		},
		{
			name:     "HasPritvorKP40: true",
			key:      "HasPritvorKP40",
			expected: true,
			ctx:      Context{HasPritvorKP40: true},
			want:     true,
		},
		{
			name:     "HasPritvorKP40: false",
			key:      "HasPritvorKP40",
			expected: true,
			ctx:      Context{HasPritvorKP40: false},
			want:     false,
		},
		{
			name:     "неизвестное поле",
			key:      "UnknownField",
			expected: true,
			ctx:      Context{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fieldMatches(tt.key, tt.expected, tt.ctx)
			if got != tt.want {
				t.Errorf("fieldMatches(%q, %v, ctx) = %v, want %v", tt.key, tt.expected, got, tt.want)
			}
		})
	}
}

func TestGetCountMaterials(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		ctx       Context
		itemCount int
		want      float64
	}{
		{
			name:      "HasImpostCount",
			field:     "HasImpostCount",
			ctx:       Context{ImpostCount: 3.0},
			itemCount: 1,
			want:      3.0,
		},
		{
			name:      "StvTCount600",
			field:     "StvTCount600",
			ctx:       Context{StvTCount600: 3.0},
			itemCount: 2,
			want:      3.0,
		},
		{
			name:      "StvTCount400",
			field:     "StvTCount400",
			ctx:       Context{StvTCount400: 3.0},
			itemCount: 2,
			want:      3.0,
		},
		{
			name:      "HasPritvorKP40 = true → itemCount",
			field:     "HasPritvorKP40",
			ctx:       Context{HasPritvorKP40: true},
			itemCount: 5,
			want:      5.0,
		},
		{
			name:      "HasPritvorKP40 = false → itemCount всё равно возвращается",
			field:     "HasPritvorKP40",
			ctx:       Context{HasPritvorKP40: false},
			itemCount: 3,
			want:      3.0,
		},
		{
			name:      "ItemCountForRDRH: HasPetliRDRH = true",
			field:     "ItemCountForRDRH",
			ctx:       Context{HasPetliRDRH: true},
			itemCount: 2,
			want:      2.0,
		},
		{
			name:      "ItemCountForRDRH: HasPetliRDRH = false",
			field:     "ItemCountForRDRH",
			ctx:       Context{HasPetliRDRH: false},
			itemCount: 2,
			want:      0.0,
		},
		{
			name:      "неизвестное поле → 0",
			field:     "UnknownField",
			ctx:       Context{},
			itemCount: 1,
			want:      0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCountMaterials(tt.field, tt.ctx, tt.itemCount)
			if got != tt.want {
				t.Errorf("getCountMaterials() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildContextWindow(t *testing.T) {
	tests := []struct {
		name      string
		materials []*storage.KlaesMaterials
		wantCtx   Context
	}{
		{
			name: "окно без импоста",
			materials: []*storage.KlaesMaterials{
				newMaterial("рама", 4.0, 100.0),
			},
			wantCtx: Context{
				Type:         "window",
				HasImpost:    false,
				ImpostCount:  0.0,
				StvTCount600: 0.0,
				StvTCount400: 0.0,
			},
		},
		{
			name: "окно с 1 импостом",
			materials: []*storage.KlaesMaterials{
				newMaterial("Импост", 1.0, 100.0),
			},
			wantCtx: Context{
				Type:         "window",
				HasImpost:    true,
				ImpostCount:  1.0,
				StvTCount600: 0.0,
				StvTCount400: 0.0,
			},
		},
		{
			name: "окно с 2 импостами",
			materials: []*storage.KlaesMaterials{
				newMaterial("Импост", 2.0, 300.0),
			},
			wantCtx: Context{
				Type:         "window",
				HasImpost:    true,
				ImpostCount:  2.0,
				StvTCount400: 0.0,
				StvTCount600: 0.0,
			},
		},
		{
			name: "створка < 600мм",
			materials: []*storage.KlaesMaterials{
				newMaterial("Створка Т - образ.", 1.0, 600.0),
			},
			wantCtx: Context{
				Type:         "window",
				HasImpost:    false,
				ImpostCount:  0.0,
				StvTCount600: 1.0,
				StvTCount400: 0.0,
			},
		},
		{
			name: "створка < 400мм (попадает в 600мм и 400мм)",
			materials: []*storage.KlaesMaterials{
				newMaterial("Створка Т - образ.", 1.0, 350.0),
			},
			wantCtx: Context{
				Type:         "window",
				HasImpost:    false,
				ImpostCount:  0.0,
				StvTCount600: 1.0,
				StvTCount400: 1.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildContextWindow(tt.materials)

			// Проверяем поля по одному — так понятнее, где ошибка
			if got.Type != tt.wantCtx.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantCtx.Type)
			}
			if got.HasImpost != tt.wantCtx.HasImpost {
				t.Errorf("HasImpost = %v, want %v", got.HasImpost, tt.wantCtx.HasImpost)
			}
			if got.ImpostCount != tt.wantCtx.ImpostCount {
				t.Errorf("ImpostCount = %v, want %v", got.ImpostCount, tt.wantCtx.ImpostCount)
			}
			if got.StvTCount600 != tt.wantCtx.StvTCount600 {
				t.Errorf("StvTCount600 = %v, want %v", got.StvTCount600, tt.wantCtx.StvTCount600)
			}
			if got.StvTCount400 != tt.wantCtx.StvTCount400 {
				t.Errorf("StvTCount400 = %v, want %v", got.StvTCount400, tt.wantCtx.StvTCount400)
			}
		})
	}
}

func TestBuildContextGlyhar(t *testing.T) {
	tests := []struct {
		name      string
		materials []*storage.KlaesMaterials
		wantCtx   Context
	}{
		{
			name: "окно без импоста",
			materials: []*storage.KlaesMaterials{
				newMaterial("рама", 4.0, 100.0),
			},
			wantCtx: Context{
				Type:         "glyhar",
				HasImpost:    false,
				ImpostCount:  0.0,
				StvTCount600: 0.0,
				StvTCount400: 0.0,
			},
		},
		{
			name: "окно с 1 импостом",
			materials: []*storage.KlaesMaterials{
				newMaterial("Импост", 1.0, 100.0),
			},
			wantCtx: Context{
				Type:         "glyhar",
				HasImpost:    true,
				ImpostCount:  1.0,
				StvTCount600: 0.0,
				StvTCount400: 0.0,
			},
		},
		{
			name: "окно с 2 импостами",
			materials: []*storage.KlaesMaterials{
				newMaterial("Импост", 2.0, 300.0),
			},
			wantCtx: Context{
				Type:         "glyhar",
				HasImpost:    true,
				ImpostCount:  2.0,
				StvTCount400: 0.0,
				StvTCount600: 0.0,
			},
		},
		{
			name: "створка < 600мм",
			materials: []*storage.KlaesMaterials{
				newMaterial("Створка Т - образ.", 1.0, 600.0),
			},
			wantCtx: Context{
				Type:         "glyhar",
				HasImpost:    false,
				ImpostCount:  0.0,
				StvTCount600: 1.0,
				StvTCount400: 0.0,
			},
		},
		{
			name: "створка < 400мм (попадает в 600мм и 400мм)",
			materials: []*storage.KlaesMaterials{
				newMaterial("Створка Т - образ.", 1.0, 350.0),
			},
			wantCtx: Context{
				Type:         "glyhar",
				HasImpost:    false,
				ImpostCount:  0.0,
				StvTCount600: 1.0,
				StvTCount400: 1.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildContextGlyhar(tt.materials)

			// Проверяем поля по одному — так понятнее, где ошибка
			if got.Type != tt.wantCtx.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantCtx.Type)
			}
			if got.HasImpost != tt.wantCtx.HasImpost {
				t.Errorf("HasImpost = %v, want %v", got.HasImpost, tt.wantCtx.HasImpost)
			}
			if got.ImpostCount != tt.wantCtx.ImpostCount {
				t.Errorf("ImpostCount = %v, want %v", got.ImpostCount, tt.wantCtx.ImpostCount)
			}
			if got.StvTCount600 != tt.wantCtx.StvTCount600 {
				t.Errorf("StvTCount600 = %v, want %v", got.StvTCount600, tt.wantCtx.StvTCount600)
			}
			if got.StvTCount400 != tt.wantCtx.StvTCount400 {
				t.Errorf("StvTCount400 = %v, want %v", got.StvTCount400, tt.wantCtx.StvTCount400)
			}
		})
	}
}

func TestBuildContextDoor(t *testing.T) {
	tests := []struct {
		name      string
		materials []*storage.KlaesMaterials
		wantCtx   Context
	}{
		{
			name: "дверь без импостов",
			materials: []*storage.KlaesMaterials{
				newMaterial("рама", 4.0, 600.0),
			},
			wantCtx: Context{
				Type:         "door",
				HasImpost:    false,
				ImpostCount:  0.0,
				StvTCount600: 0.0,
				StvTCount400: 0.0,
			},
		},
		{
			name: "дверь с 2 импостами",
			materials: []*storage.KlaesMaterials{
				newMaterial("Импост", 2.0, 300.0),
			},
			wantCtx: Context{
				Type:         "door",
				HasImpost:    true,
				ImpostCount:  2.0,
				StvTCount600: 0.0,
				StvTCount400: 0.0,
			},
		},
		{
			name: "накладка стаблина",
			materials: []*storage.KlaesMaterials{
				newMaterial("Накладка на цилиндр Stublina", 1.0, 300.0),
			},
			wantCtx: Context{
				Type:          "door",
				StublinaCount: 1.0,
			},
		},
		{
			name: "створки < 600",
			materials: []*storage.KlaesMaterials{
				newMaterial("Створка Т-образная", 1.0, 600.0),
			},
			wantCtx: Context{
				Type:         "door",
				StvTCount600: 1.0,
			},
		},
		{
			name: "створки < 400",
			materials: []*storage.KlaesMaterials{
				newMaterial("Створка Т-образная", 1.0, 350.0),
			},
			wantCtx: Context{
				Type:         "door",
				StvTCount600: 1.0,
				StvTCount400: 1.0,
			},
		},
		{
			name: "притвор кп40",
			materials: []*storage.KlaesMaterials{
				newMaterial("Притвор КП40", 1.0, 350.0),
			},
			wantCtx: Context{
				Type:           "door",
				PritvorKP40:    1.0,
				HasPritvorKP40: true,
			},
		},
		{
			name: "петли стандартные",
			materials: []*storage.KlaesMaterials{
				newMaterial("Петля двухсекционная 67мм", 2.0, 350.0),
			},
			wantCtx: Context{
				Type:       "door",
				PetliStand: 2.0,
			},
		},
		{
			name: "петли роликовые",
			materials: []*storage.KlaesMaterials{
				newMaterial("Петля роликовая для КП45", 2.0, 350.0),
			},
			wantCtx: Context{
				Type:       "door",
				PetliRolik: 2.0,
			},
		},
		{
			name: "петли трехсекционные",
			materials: []*storage.KlaesMaterials{
				newMaterial("Петля дверная трехсекционная с удлиненной базой", 2.0, 350.0),
			},
			wantCtx: Context{
				Type:          "door",
				Petli3Section: 2.0,
			},
		},
		{
			name: "петли фурал",
			materials: []*storage.KlaesMaterials{
				newMaterial("Петля Фурал дверная 2-част. с подшипником", 2.0, 350.0),
			},
			wantCtx: Context{
				Type:          "door",
				PetliFural:    2.0,
				HasPetliFural: true,
			},
		},
		{
			name: "петли RDRH",
			materials: []*storage.KlaesMaterials{
				newMaterial("Петля роликовая RDRH", 2.0, 350.0),
			},
			wantCtx: Context{
				Type:         "door",
				PetliRDRH:    2.0,
				HasPetliRDRH: true,
			},
		},
		{
			name: "Многозапорный замок",
			materials: []*storage.KlaesMaterials{
				newMaterial("Многозапорный замок Stublina с управлением от ручки", 1.0, 350.0),
			},
			wantCtx: Context{
				Type:          "door",
				MnogozapZamok: 1.0,
				StandZamok:    1.0,
			},
		},
		{
			name: "Обычный замок",
			materials: []*storage.KlaesMaterials{
				newMaterial("Замок Elementis 1153 (D30) (под нажимной гарнитур)", 1.0, 350.0),
			},
			wantCtx: Context{
				Type:       "door",
				StandZamok: 1.0,
			},
		},
		{
			name: "Обычный замок",
			materials: []*storage.KlaesMaterials{
				newMaterial("Замок Elementis 1153 (D30) (под нажимной гарнитур)", 2.0, 350.0),
			},
			wantCtx: Context{
				Type:       "door",
				StandZamok: 2.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildContextDoor(tt.materials, nil)

			if got.Type != tt.wantCtx.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantCtx.Type)
			}
			if got.HasImpost != tt.wantCtx.HasImpost {
				t.Errorf("HasImpost = %v, want %v", got.HasImpost, tt.wantCtx.HasImpost)
			}
			if got.ImpostCount != tt.wantCtx.ImpostCount {
				t.Errorf("HasImpost = %v, want %v", got.ImpostCount, tt.wantCtx.ImpostCount)
			}
			if got.StvTCount600 != tt.wantCtx.StvTCount600 {
				t.Errorf("StvTCount600 = %v, want %v", got.StvTCount600, tt.wantCtx.StvTCount600)
			}
			if got.StvTCount400 != tt.wantCtx.StvTCount400 {
				t.Errorf("StvTCount400 = %v, want %v", got.StvTCount400, tt.wantCtx.StvTCount400)
			}
			if got.HasPritvorKP40 != tt.wantCtx.HasPritvorKP40 {
				t.Errorf("HasPritvorKP40 = %v, want %v", got.HasPritvorKP40, tt.wantCtx.HasPritvorKP40)
			}
			if got.PritvorKP40 != tt.wantCtx.PritvorKP40 {
				t.Errorf("PritvorKP40 = %v, want %v", got.PritvorKP40, tt.wantCtx.PritvorKP40)
			}
			if got.PetliStand != tt.wantCtx.PetliStand {
				t.Errorf("PetliStand = %v, want %v", got.PetliStand, tt.wantCtx.PetliStand)
			}
			if got.PetliRolik != tt.wantCtx.PetliRolik {
				t.Errorf("PetliRolik = %v, want %v", got.PetliRolik, tt.wantCtx.PetliRolik)
			}
			if got.Petli3Section != tt.wantCtx.Petli3Section {
				t.Errorf("Petli3Section = %v, want %v", got.Petli3Section, tt.wantCtx.Petli3Section)
			}
			if got.PetliFural != tt.wantCtx.PetliFural {
				t.Errorf("PetliFural = %v, want %v", got.PetliFural, tt.wantCtx.PetliFural)
			}
			if got.HasPetliFural != tt.wantCtx.HasPetliFural {
				t.Errorf("HasPetliFural = %v, want %v", got.HasPetliFural, tt.wantCtx.HasPetliFural)
			}
			if got.PetliRDRH != tt.wantCtx.PetliRDRH {
				t.Errorf("PetliRDRH = %v, want %v", got.PetliRDRH, tt.wantCtx.PetliRDRH)
			}
			if got.HasPetliRDRH != tt.wantCtx.HasPetliRDRH {
				t.Errorf("HasPetliRDRH = %v, want %v", got.HasPetliRDRH, tt.wantCtx.HasPetliRDRH)
			}
			if got.MnogozapZamok != tt.wantCtx.MnogozapZamok {
				t.Errorf("MnogozapZamok = %v, want %v", got.MnogozapZamok, tt.wantCtx.MnogozapZamok)
			}
			if got.StandZamok != tt.wantCtx.StandZamok {
				t.Errorf("StandZamok = %v, want %v", got.StandZamok, tt.wantCtx.StandZamok)
			}
		})
	}
}

func TestApplyRules_BasicMultiplication(t *testing.T) {
	// Исходные операции из шаблона
	operations := []storage.Operation{
		{
			Name:    "сборка",
			Group:   "", // не "ign" → умножается
			Value:   10.0,
			Minutes: 30.0,
			Count:   1.0,
		},
		{
			Name:    "настройка оборудования",
			Group:   "ign", // группа "ign" → НЕ умножается
			Value:   5.0,
			Minutes: 15.0,
			Count:   1.0,
		},
	}

	// Пустые правила — просто умножение на количество
	rules := []storage.Rule{}
	ctx := Context{Type: "door"}
	itemCount := 3

	// Применяем правила
	result := ApplyRules(operations, rules, ctx, itemCount)

	// Проверяем результат
	assert.Len(t, result, 2, "должно быть 2 операции")

	// Операция "сборка" умножена на 3
	assert.Equal(t, "сборка", result[0].Name)
	assert.Equal(t, 30.0, result[0].Value, "Value должно быть 10 * 3")
	assert.Equal(t, 90.0, result[0].Minutes, "Minutes должно быть 30 * 3")
	assert.Equal(t, 3.0, result[0].Count, "Count должно быть 1 * 3")

	// Операция "настройка" НЕ умножена (группа "ign")
	assert.Equal(t, "настройка оборудования", result[1].Name)
	assert.Equal(t, 5.0, result[1].Value, "Value НЕ должно умножаться (группа 'ign')")
	assert.Equal(t, 15.0, result[1].Minutes, "Minutes НЕ должно умножаться (группа 'ign')")
	assert.Equal(t, 1.0, result[1].Count, "Count НЕ должно умножаться (группа 'ign')")
}

func TestApplyRules_SetMode(t *testing.T) {
	operations := []storage.Operation{
		{
			Name:    "монтаж петель",
			Group:   "",
			Value:   10.0,
			Minutes: 20.0,
			Count:   1.0,
		},
	}

	// Правило: если есть импост → заменить время на фиксированное
	rules := []storage.Rule{
		{
			Operation: "монтаж петель",
			Condition: map[string]interface{}{
				"HasImpost": true,
			},
			Mode:       "set",
			SetValue:   5.0,
			SetMinutes: 45.0,
		},
	}

	ctx := Context{
		Type:      "door",
		HasImpost: true, // ← условие выполняется
	}
	itemCount := 2

	result := ApplyRules(operations, rules, ctx, itemCount)

	assert.Len(t, result, 1)
	assert.Equal(t, "монтаж петель", result[0].Name)

	// После умножения на itemCount (2) применяется правило "set"
	// ВАЖНО: в текущей реализации сначала умножение, потом правило "set" перезаписывает
	assert.Equal(t, 5.0, result[0].Value, "Value заменено правилом 'set'")
	assert.Equal(t, 45.0, result[0].Minutes, "Minutes заменено правилом 'set'")
}

func TestApplyRules_MultipliedMode(t *testing.T) {
	operations := []storage.Operation{
		{
			Name:    "установка замка",
			Group:   "",
			Value:   1.0,  // базовая стоимость за 1 шт
			Minutes: 10.0, // базовое время за 1 шт
			Count:   1.0,
		},
	}

	// Правило: умножить на количество замков из контекста
	rules := []storage.Rule{
		{
			Operation:      "установка замка",
			Condition:      map[string]interface{}{}, // всегда применяется
			Mode:           "multiplied",
			UnitField:      "StandZamok", // брать из этого поля контекста
			ValuePerUnit:   1.0,
			MinutesPerUnit: 10.0,
		},
	}

	ctx := Context{
		Type:       "door",
		StandZamok: 3.0, // ← 3 замка в заказе
	}
	itemCount := 1 // количество изделий (не влияет — правило берёт из контекста)

	result := ApplyRules(operations, rules, ctx, itemCount)

	assert.Len(t, result, 1)
	assert.Equal(t, "установка замка", result[0].Name)
	assert.Equal(t, 3.0, result[0].Value, "Value = 1.0 * 3 замка")
	assert.Equal(t, 30.0, result[0].Minutes, "Minutes = 10.0 * 3 замка")
	assert.Equal(t, 3.0, result[0].Count, "Count = 3 замка")
}

func TestApplyRules_AdditivePlusMultiplied(t *testing.T) {
	operations := []storage.Operation{
		{
			Name:    "установка петель RDRH",
			Group:   "",
			Value:   5.0,
			Minutes: 20.0,
			Count:   1.0,
		},
	}

	rules := []storage.Rule{
		{
			Operation: "установка петель RDRH",
			Condition: map[string]interface{}{
				"HasPetliRDRH": true,
			},
			Mode:           "additivePlusMultiplied",
			UnitField:      "ItemCountForRDRH",
			ValuePerUnit:   0.0,
			MinutesPerUnit: 4.5,
		},
	}

	ctx := Context{
		Type:         "door",
		HasPetliRDRH: true,
		PetliRDRH:    3.0,
	}
	itemCount := 2

	// Ожидаем:
	// - Базовое умножение: Count = 1 * 2 = 2
	// - additivePlusMultiplied: Count += 2 (itemCount) → итого 4
	// - Value: 5 * 2 = 10 (базовое) + 0 * 2 = 10
	// - Minutes: 20 * 2 = 40 + 4.5 * 2 = 49
	result := ApplyRules(operations, rules, ctx, itemCount)

	assert.Len(t, result, 1)
	assert.Equal(t, "установка петель RDRH", result[0].Name)
	assert.Equal(t, 10.0, result[0].Value, "Value = 5.0 * 2")
	assert.Equal(t, 49.0, result[0].Minutes, "Minutes = 20*2 + 4.5*2 = 49")
	assert.Equal(t, 4.0, result[0].Count, "Count = 1*2 (базовое) + 2 (доп. за RDRH) = 4")
}

func TestBuildContext(t *testing.T) {
	materials := []*storage.KlaesMaterials{
		newMaterial("Импост", 1.0, 500.0),
	}

	ctx, err := BuildContext(materials, nil, "door")
	assert.NoError(t, err)
	assert.Equal(t, "door", ctx.Type)
	assert.True(t, ctx.HasImpost)

	ctx, err = BuildContext(materials, nil, "window")
	assert.NoError(t, err)
	assert.Equal(t, "window", ctx.Type)
	assert.True(t, ctx.HasImpost)

	ctx, err = BuildContext(materials, nil, "glyhar")
	assert.NoError(t, err)
	assert.Equal(t, "glyhar", ctx.Type)
	assert.True(t, ctx.HasImpost)

	// Тест для неизвестного типа
	_, err = BuildContext(materials, nil, "unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "неизвестный тип изделия")
}
