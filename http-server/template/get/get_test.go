package get

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-chi/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"vue-golang/internal/storage"
)

// MockTemplateJSON реализует интерфейс TemplateJSON для тестов
type MockTemplateJSON struct {
	mock.Mock
}

func (m *MockTemplateJSON) GetTemplateByCode(ctx context.Context, code string) (*storage.Template, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.Template), args.Error(1)
}

func (m *MockTemplateJSON) GetAllTemplates(ctx context.Context) ([]*storage.Template, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.Template), args.Error(1)
}

func (m *MockTemplateJSON) GetTemplateByCodeAdmin(ctx context.Context, code string) (*storage.Template, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.Template), args.Error(1)
}

func (m *MockTemplateJSON) GetAllTemplatesAdmin(ctx context.Context) ([]*storage.Template, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.Template), args.Error(1)
}

// Тест: успешное получение шаблона по коду
func TestGetTemplatesByCode_Success(t *testing.T) {
	mockStorage := new(MockTemplateJSON)

	// Подготовка фейкового шаблона
	template := &storage.Template{
		ID:       56,
		Code:     "56",
		Name:     "Дверь с петлями RDRH",
		Category: "door",
		Systema:  strPtr("КП45"),
		Operations: []storage.Operation{
			{Name: "сборка", Minutes: 45.0},
			{Name: "адаптер ПДП 1001-00", Minutes: 0.0},
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

	mockStorage.On("GetTemplateByCode", mock.Anything, "56").
		Return(template, nil)

	logger := slog.Default()
	handler := GetTemplatesByCode(logger, mockStorage)

	// Создаём запрос с query параметром ?code=DOOR-56
	req := httptest.NewRequest(http.MethodGet, "/api/templates?code=56", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Проверяем статус
	assert.Equal(t, http.StatusOK, rr.Code)

	// Проверяем ответ
	var resp ResponseForm
	err := render.DecodeJSON(strings.NewReader(rr.Body.String()), &resp)
	assert.NoError(t, err)

	assert.Equal(t, 56, resp.ID)
	assert.Equal(t, "56", resp.Code)
	assert.Equal(t, "Дверь с петлями RDRH", resp.Name)
	assert.Equal(t, "door", resp.Category)
	assert.Equal(t, "КП45", *resp.Systema)
	assert.Len(t, resp.Operations, 2)
	assert.Equal(t, "сборка", resp.Operations[0].Name)

	mockStorage.AssertExpectations(t)
}

// Тест: отсутствует параметр 'code'
func TestGetTemplatesByCode_MissingCode(t *testing.T) {
	mockStorage := new(MockTemplateJSON)
	logger := slog.Default()
	handler := GetTemplatesByCode(logger, mockStorage)

	// Запрос БЕЗ параметра code
	req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Ожидаем 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Missing required query parameter 'code'")

	// Мок не должен был быть вызван
	mockStorage.AssertNotCalled(t, "GetTemplateByCode")
}

// Тест: шаблон не найден (404)
func TestGetTemplatesByCode_NotFound(t *testing.T) {
	mockStorage := new(MockTemplateJSON)

	// Возвращаем ошибку "не найден"
	mockStorage.On("GetTemplateByCode", mock.Anything, "UNKNOWN").
		Return(nil, sql.ErrNoRows)

	logger := slog.Default()
	handler := GetTemplatesByCode(logger, mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/api/templates?code=UNKNOWN", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Ожидаем 404 Not Found
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Form not found")

	mockStorage.AssertExpectations(t)
}

// Тест: ошибка базы данных (500)
func TestGetTemplatesByCode_DBError(t *testing.T) {
	mockStorage := new(MockTemplateJSON)

	// Возвращаем произвольную ошибку БД
	mockStorage.On("GetTemplateByCode", mock.Anything, "56").
		Return(nil, errors.New("connection timeout"))

	logger := slog.Default()
	handler := GetTemplatesByCode(logger, mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/api/templates?code=56", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Ожидаем 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Internal server error")

	mockStorage.AssertExpectations(t)
}

// Тест: успешное получение всех шаблонов
func TestGetAllTemplates_Success(t *testing.T) {
	mockStorage := new(MockTemplateJSON)

	templates := []*storage.Template{
		{ID: 1, Code: "WIN-01", Name: "Окно стандарт", Category: "window"},
		{ID: 56, Code: "DOOR-56", Name: "Дверь с RDRH", Category: "door"},
	}

	mockStorage.On("GetAllTemplates", mock.Anything).
		Return(templates, nil)

	logger := slog.Default()
	handler := GetAllTemplates(logger, mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/api/templates/all", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Проверяем структуру ответа ResponseAllForm
	var resp ResponseAllForm
	err := render.DecodeJSON(strings.NewReader(rr.Body.String()), &resp)
	assert.NoError(t, err)

	assert.Len(t, resp.Template, 2)
	assert.Equal(t, "WIN-01", resp.Template[0].Code)
	assert.Equal(t, "DOOR-56", resp.Template[1].Code)
	assert.Empty(t, resp.Error)

	mockStorage.AssertExpectations(t)
}

func TestGetAllTemplates_DBError(t *testing.T) {
	mockStorage := new(MockTemplateJSON)

	mockStorage.On("GetAllTemplates", mock.Anything).Return(nil, errors.New("connection timeout"))

	logger := slog.Default()
	handler := GetAllTemplates(logger, mockStorage)

	req := httptest.NewRequest(http.MethodGet, "/api/template/all", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Internal server error")

	mockStorage.AssertExpectations(t)
}

// Вспомогательная функция для создания указателя на строку
func strPtr(s string) *string {
	return &s
}
