package generate_excel

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"vue-golang/internal/storage"
	"vue-golang/internal/storage/mysql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

type MockGenerateExcelStorage struct {
	mock.Mock
}

func (m *MockGenerateExcelStorage) GetPEOProductsByCategory(ctx context.Context, filter mysql.ProductFilter) ([]storage.PEOProduct, []storage.GetWorkers, error) {
	args := m.Called(ctx, filter)

	var products []storage.PEOProduct
	if args.Get(0) != nil {
		products = args.Get(0).([]storage.PEOProduct)
	}

	var workers []storage.GetWorkers
	if args.Get(1) != nil {
		workers = args.Get(1).([]storage.GetWorkers)
	}

	return products, workers, args.Error(2)
}

func TestGenerateExcel_storageError(t *testing.T) {
	mockStorage := new(MockGenerateExcelStorage)
	mockStorage.On("GetPEOProductsByCategory", mock.Anything, mock.Anything).
		Return(nil, nil, errors.New("database error"))

	service := NewGenerateService(mockStorage)

	result, err := service.GenerateExcel(context.Background(), mysql.ProductFilter{})
	require.Error(t, err)
	assert.Nil(t, result)

	mockStorage.AssertExpectations(t)
}

func TestGenerateExcel_OK(t *testing.T) {
	mockStorage := new(MockGenerateExcelStorage)
	products := []storage.PEOProduct{
		{
			OrderNum:  "Q6-777",
			Type:      "door",
			TypeIzd:   "дверь",
			Count:     2,
			Sqr:       0.43,
			TotalTime: 7.7,
			NormMoney: 1000,
			Systema:   "т",
		},
	}

	workers := []storage.GetWorkers{
		{ID: 1, Name: "Ivan"},
		{ID: 2, Name: "Petr"},
	}

	mockStorage.On("GetPEOProductsByCategory", mock.Anything, mock.Anything).
		Return(products, workers, nil)

	service := NewGenerateService(mockStorage)

	result, err := service.GenerateExcel(context.Background(), mysql.ProductFilter{
		Type: []string{"door"},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	mockStorage.AssertExpectations(t)

	file, err := excelize.OpenReader(bytes.NewReader(result))
	require.NoError(t, err)
	defer file.Close()

	sheet := file.GetSheetList()
	assert.Contains(t, sheet, "Отчет ПЭО")

	value, err := file.GetCellValue("Отчет ПЭО", "A1")
	require.NoError(t, err)
	assert.Equal(t, "Спецификация", value)
}

func TestGetWindowStats(t *testing.T) {
	products := []storage.PEOProduct{
		{
			Type:      "window",
			Count:     2,
			Sqr:       0.5,
			TotalTime: 7.7,
			NormMoney: 1000,
			Systema:   "х",
		},
	}

	s := &GenerateExcelService{}

	stats := s.getWindowStats(products)
	assert.Equal(t, "Холодные окна", stats[0].Label)
	assert.Equal(t, 2, stats[0].Count)
}

func TestGetDoorStats(t *testing.T) {
	products := []storage.PEOProduct{
		{
			Type:      "door",
			Count:     3,
			Sqr:       0.5,
			TotalTime: 7.7,
			NormMoney: 1000,
			Systema:   "т",
			TypeIzd:   "1П",
		},
	}

	s := &GenerateExcelService{}

	stats := s.getDoorStats(products)
	assert.Equal(t, "Всего 1П дверей", stats[0].Label)
	assert.Equal(t, 3, stats[0].Count)
}
