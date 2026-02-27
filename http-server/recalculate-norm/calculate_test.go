package recalculate_norm

import (
	"context"
	"github.com/go-chi/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"vue-golang/internal/service/recalculate"
	"vue-golang/internal/storage"
)

type MockNormCalculation struct {
	mock.Mock
}

func (m *MockNormCalculation) CalculateNorm(ctx context.Context, orderNum string, pos int, typeIzd string, templateCode string, itemCount int) ([]storage.Operation, recalculate.Context, error) {
	args := m.Called(ctx, orderNum, pos, typeIzd, templateCode, itemCount)

	ops := []storage.Operation{}
	if args.Get(0) != nil {
		ops = args.Get(0).([]storage.Operation)
	}

	ctxData := recalculate.Context{}
	if args.Get(1) != nil {
		ctxData = args.Get(1).(recalculate.Context)
	}

	return ops, ctxData, args.Error(2)
}

func TestCalculateNormOperations_Success(t *testing.T) {
	// 1. Создаём мок калькулятора
	mockCalc := new(MockNormCalculation)

	// 2. Настраиваем мок на успешный ответ
	operations := []storage.Operation{
		{Name: "сборка", Minutes: 60.0, Value: 15.0, Count: 2.0},
		{Name: "адаптер ПДП 1001-00", Minutes: 9.0, Value: 0.0, Count: 4.0},
	}
	ctxData := recalculate.Context{
		Type:         "door",
		HasPetliRDRH: true,
		PetliRDRH:    3.0,
	}

	mockCalc.On("CalculateNorm",
		mock.Anything, // context
		"ORD-789",     // orderNum
		1,             // position
		"door",        // typeIzd
		"56",          // templateCode
		2,             // itemCount
	).Return(operations, ctxData, nil)

	// 3. Создаём фейковый логгер
	logger := slog.Default()

	// 4. Создаём хендлер
	handler := CalculateNormOperations(logger, mockCalc)

	// 5. Создаём фейковый HTTP запрос с валидным JSON
	reqBody := `{
		"order_num": "ORD-789",
		"position": 1,
		"type": "door",
		"template": "56",
		"count": 2
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/materials/calculation", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// 6. Создаём фейковый ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// 7. Вызываем хендлер
	handler.ServeHTTP(rr, req)

	// 8. Проверяем статус код
	assert.Equal(t, http.StatusOK, rr.Code, "ожидался статус 200")

	// 9. Проверяем заголовок Content-Type
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// 10. Проверяем тело ответа
	var resp Resp
	err := render.DecodeJSON(strings.NewReader(rr.Body.String()), &resp)
	assert.NoError(t, err, "ошибка декодирования JSON ответа")

	// 11. Проверяем операции в ответе
	assert.Len(t, resp.Operation, 2, "ожидалось 2 операции")
	assert.Equal(t, "сборка", resp.Operation[0].Name)
	assert.Equal(t, 60.0, resp.Operation[0].Minutes)
	assert.Equal(t, "адаптер ПДП 1001-00", resp.Operation[1].Name)
	assert.Equal(t, 9.0, resp.Operation[1].Minutes)

	// 12. Проверяем контекст в ответе
	assert.Equal(t, "door", resp.Context.Type)
	assert.True(t, resp.Context.HasPetliRDRH)
	assert.Equal(t, 3.0, resp.Context.PetliRDRH)

	// 13. Проверяем вызовы мока
	mockCalc.AssertExpectations(t)
}

func TestCalculateNormOperations_InvalidJSON(t *testing.T) {
	mockCalc := new(MockNormCalculation)
	logger := slog.Default()
	handler := CalculateNormOperations(logger, mockCalc)

	// Невалидный JSON (отсутствует закрывающая скобка)
	req := httptest.NewRequest(http.MethodPost, "/api/materials/calculation", strings.NewReader(`{`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Ожидаем 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Некорректный JSON")

	// Мок не должен был быть вызван
	mockCalc.AssertNotCalled(t, "CalculateNorm")
}

func TestCalculateNormOperations_ServiceError(t *testing.T) {
	mockCalc := new(MockNormCalculation)

	// Настраиваем мок на возврат ошибки
	mockCalc.On("CalculateNorm",
		mock.Anything, "ORD-123", 1, "door", "TEST", 1,
	).Return([]storage.Operation{}, recalculate.Context{}, assert.AnError)

	logger := slog.Default()
	handler := CalculateNormOperations(logger, mockCalc)

	reqBody := `{
		"order_num": "ORD-123",
		"position": 1,
		"type": "door",
		"template": "TEST",
		"count": 1
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/norms/recalculate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Ожидаем 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "Internal error")

	mockCalc.AssertExpectations(t)
}

func TestCalculateNormOperations_ContextCanceled(t *testing.T) {
	mockCalc := new(MockNormCalculation)

	// Мок "ждёт", пока контекст не будет отменён
	mockCalc.On("CalculateNorm", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			<-ctx.Done() // ждём отмены контекста
		}).
		Return([]storage.Operation{}, recalculate.Context{}, context.Canceled)

	logger := slog.Default()
	handler := CalculateNormOperations(logger, mockCalc)

	// Создаём запрос с КОРОТКИМ таймаутом (10мс вместо 5 сек)
	// Но хендлер внутри устанавливает 5 сек, поэтому имитируем отмену через родительский контекст
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // немедленно отменяем

	reqBody := `{"order_num": "ORD-123", "position": 1, "type": "door", "template": "TEST", "count": 1}`
	req := httptest.NewRequest(http.MethodPost, "/api/norms/recalculate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx) // передаём отменённый контекст

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Ожидаем 500 из-за отмены контекста
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockCalc.AssertExpectations(t)
}
