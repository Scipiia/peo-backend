package recalculate_norm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"vue-golang/internal/service/recalculate"
	"vue-golang/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockNormCalculator struct {
	mock.Mock
}

func (m *MockNormCalculator) CalculateNorm(ctx context.Context, orderNum string, pos int, typeIzd string, templateCode string, itemCount int, permisDopMaterial bool) ([]storage.Operation, recalculate.Context, error) {
	args := m.Called(ctx, orderNum, pos, typeIzd, templateCode, itemCount, permisDopMaterial)

	// Безопасная обработка nil для слайса операций
	var ops []storage.Operation
	if args.Get(0) != nil {
		ops = args.Get(0).([]storage.Operation)
	}

	// Безопасная обработка nil для контекста
	var ctxData recalculate.Context
	if args.Get(1) != nil {
		ctxData = args.Get(1).(recalculate.Context)
	}

	return ops, ctxData, args.Error(2)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCalculateNormOperations(t *testing.T) {
	// Тестовые данные, которые будет возвращать наш мок
	mockOperations := []storage.Operation{
		{Name: "Сборка", Group: "", Value: 10, Minutes: 5, Count: 1},
	}
	mockContext := recalculate.Context{Type: "window"}

	tests := []struct {
		name           string
		reqBody        string
		setupMock      func(m *MockNormCalculator) // <-- Теперь используем наш локальный мок
		expectedStatus int
		expectedInBody string
	}{
		{
			name: "Успешный расчет (window)",
			reqBody: `{
				"order_num": "ORD-123",
				"position": 1,
				"type": "window",
				"template": "TPL-001",
				"count": 2,
				"permis_dop_material": true
			}`,
			setupMock: func(m *MockNormCalculator) {
				m.On("CalculateNorm", mock.Anything, "ORD-123", 1, "window", "TPL-001", 2, true).
					Return(mockOperations, mockContext, nil)
			},
			expectedStatus: http.StatusOK,
			expectedInBody: `"operation"`,
		},
		{
			name: "Некорректный JSON",
			reqBody: `{
				"order_num": "ORD-123",
				"position": 1,
				"type": "window" 
				// битый JSON
			}`,
			setupMock:      func(m *MockNormCalculator) {}, // Мок не вызывается, упадет на json.Decode
			expectedStatus: http.StatusBadRequest,
			expectedInBody: "Некорректный JSON",
		},
		{
			name: "Ошибка внутри сервиса (БД недоступна)",
			reqBody: `{
				"order_num": "ORD-ERR",
				"position": 1,
				"type": "window",
				"template": "TPL-001",
				"count": 1,
				"permis_dop_material": true
			}`,
			setupMock: func(m *MockNormCalculator) {
				m.On("CalculateNorm", mock.Anything, "ORD-ERR", 1, "window", "TPL-001", 1, true).
					Return(nil, recalculate.Context{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedInBody: "Internal error",
		},
		{
			name: "Принудительное включение permis_dop_material для типа door",
			reqBody: `{
				"order_num": "ORD-DOOR",
				"position": 1,
				"type": "door",
				"template": "TPL-001",
				"count": 1,
				"permis_dop_material": false
			}`,
			setupMock: func(m *MockNormCalculator) {
				// ВАЖНО: Проверяем, что хендлер передал true, несмотря на false в JSON
				m.On("CalculateNorm", mock.Anything, "ORD-DOOR", 1, "door", "TPL-001", 1, true).
					Return(mockOperations, mockContext, nil)
			},
			expectedStatus: http.StatusOK,
			expectedInBody: `"operation"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Создаем и настраиваем локальный мок
			mockCalc := new(MockNormCalculator)
			tt.setupMock(mockCalc)

			// 2. Создаем хендлер, передавая ему мок (так как мок реализует NormCalculator)
			handler := CalculateNormOperations(testLogger(), mockCalc)

			// 3. Формируем HTTP-запрос
			req := httptest.NewRequest(http.MethodPost, "/api/calculate", bytes.NewBufferString(tt.reqBody))
			req.Header.Set("Content-Type", "application/json")

			// 4. Создаем "ловушку" для ответа
			rr := httptest.NewRecorder()

			// 5. Вызываем хендлер
			handler.ServeHTTP(rr, req)

			// 6. Проверяем статус код
			assert.Equal(t, tt.expectedStatus, rr.Code, "Несовпадение HTTP статуса")

			// 7. Проверяем тело ответа
			if tt.expectedInBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedInBody)
			}

			// 8. Если успех, проверяем структуру JSON
			if tt.expectedStatus == http.StatusOK {
				var resp Resp
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(t, err, "Ответ должен быть валидным JSON")
				assert.Len(t, resp.Operation, 1)
				assert.Equal(t, "Сборка", resp.Operation[0].Name)
			}

			// 9. Проверяем, что все ожидаемые вызовы мока действительно произошли
			mockCalc.AssertExpectations(t)
		})
	}
}
