package save

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"vue-golang/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockResultNormSaver struct {
	mock.Mock
}

func (m *MockResultNormSaver) SaveNormOrder(ctx context.Context, result storage.OrderNormDetails) (int64, error) {
	args := m.Mock.Called(ctx, result)

	return args.Get(0).(int64), args.Error(1)
}

func (m *MockResultNormSaver) SaveNormOperation(ctx context.Context, OrderID int64, operations []storage.NormOperation) error {
	args := m.Mock.Called(ctx, OrderID, operations)

	return args.Error(0)
}

func TestSaveNormOrderOperation(t *testing.T) {
	tests := []struct {
		name string

		body storage.OrderNormDetails

		resultID int64

		saveOrderErr     error
		saveOperationErr error

		wantStatus            int
		expectedBody          string
		checkJSON             bool
		needMock              bool
		needSaveOperationMock bool
	}{
		{
			name: "OK",
			body: storage.OrderNormDetails{
				OrderNum: "123",
				Name:     "abc",
				Operations: []storage.NormOperation{
					{
						Name:  "Резка",
						Value: 10,
					},
				},
			},
			resultID:              1,
			wantStatus:            http.StatusOK,
			needMock:              true,
			checkJSON:             true,
			needSaveOperationMock: true,
		},
		{
			name: "saveOrder Error",
			body: storage.OrderNormDetails{
				OrderNum: "123",
				Name:     "abc",
				Operations: []storage.NormOperation{
					{
						Name:  "Резка",
						Value: 10,
					},
				},
			},
			saveOrderErr:          errors.New("database error"),
			wantStatus:            http.StatusInternalServerError,
			expectedBody:          "не удалось сохранить нормировку",
			needMock:              true,
			needSaveOperationMock: false,
		},
		{
			name: "saveOperation Error",
			body: storage.OrderNormDetails{
				OrderNum: "123",
				Name:     "abc",
				Operations: []storage.NormOperation{
					{
						Name:  "Резка",
						Value: 10,
					},
				},
			},
			saveOperationErr:      errors.New("database error"),
			wantStatus:            http.StatusInternalServerError,
			expectedBody:          "не удалось сохранить операции",
			needSaveOperationMock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockResultNormSaver)
			mockService.On("SaveNormOrder", mock.Anything, tt.body).
				Return(tt.resultID, tt.saveOrderErr)

			if tt.needSaveOperationMock {
				mockService.On("SaveNormOperation", mock.Anything, tt.resultID, tt.body.Operations).
					Return(tt.saveOperationErr)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Post("/api/orders/order-norm/template", SaveNormOrderOperation(log, mockService))

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/orders/order-norm/template", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.checkJSON {
				var response Response

				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.resultID, response.OrderID)
			}

			mockService.AssertExpectations(t)
		})
	}
}

type MockSaveNashchelnikSaver struct {
	mock.Mock
}

func (m *MockSaveNashchelnikSaver) SaveNashchelnikNorm(ctx context.Context, legacyID int64, orderNum string, a, b, c, d, sqr, count float64, opsFromFront []storage.NormOperation) (*storage.GetOrderDetails, error) {
	args := m.Mock.Called(ctx, legacyID, orderNum, a, b, c, d, sqr, count, opsFromFront)

	return args.Get(0).(*storage.GetOrderDetails), args.Error(1)
}

type request struct {
	LegacyID   int64                   `json:"legacy_id"`
	OrderNum   string                  `json:"order_num"`
	A          float64                 `json:"a"`
	B          float64                 `json:"b"`
	C          float64                 `json:"c"`
	D          float64                 `json:"d"`
	Count      float64                 `json:"count"`
	Sqr        float64                 `json:"sqr"`
	Operations []storage.NormOperation `json:"operations"`
}

func TestSaveNashchelnikCalc(t *testing.T) {
	tests := []struct {
		name string

		body request

		result    *storage.GetOrderDetails
		mockError error

		wantStatus   int
		checkJSON    bool
		needMock     bool
		expectedBody string
	}{
		{
			name: "OK",
			body: request{
				LegacyID: 10,
				OrderNum: "Q6-777",
				A:        100,
				B:        200,
				C:        300,
				D:        400,
				Count:    2,
				Sqr:      500,
				Operations: []storage.NormOperation{
					{
						Name: "Резка",
					},
				},
			},
			result:     &storage.GetOrderDetails{ID: 1},
			wantStatus: http.StatusOK,
			needMock:   true,
			checkJSON:  true,
		},
		{
			name: "service error",
			body: request{
				LegacyID: 10,
				OrderNum: "Q6-777",
				A:        100,
				B:        200,
				C:        300,
				D:        400,
				Count:    2,
				Sqr:      500,
				Operations: []storage.NormOperation{
					{
						Name: "Резка",
					},
				},
			},
			mockError: errors.New("database error"),

			wantStatus:   http.StatusInternalServerError,
			needMock:     true,
			expectedBody: "Internal Server Error",
		},
		{
			name:         "Invalid json",
			wantStatus:   http.StatusBadRequest,
			expectedBody: "Invalid json",
			needMock:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockSaveNashchelnikSaver)
			reqBody := request{
				LegacyID:   10,
				OrderNum:   "Q6-777",
				A:          100,
				B:          200,
				C:          300,
				D:          400,
				Count:      2,
				Sqr:        500,
				Operations: []storage.NormOperation{{Name: "Резка"}},
			}

			if tt.needMock {
				mockService.On(
					"SaveNashchelnikNorm",
					mock.Anything,
					reqBody.LegacyID,
					reqBody.OrderNum,
					reqBody.A,
					reqBody.B,
					reqBody.C,
					reqBody.D,
					reqBody.Sqr,
					reqBody.Count,
					reqBody.Operations,
				).Return(tt.result, tt.mockError)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()

			r.Post("/api/orders/nashchelnik/calc", SaveNashchelnikCalc(log, mockService))

			var req *http.Request

			if tt.name == "Invalid json" {
				req = httptest.NewRequest(
					http.MethodPost,
					"/api/orders/nashchelnik/calc",
					strings.NewReader("{invalid"),
				)
			} else {
				body, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(
					http.MethodPost,
					"/api/orders/nashchelnik/calc",
					bytes.NewReader(body),
				)
			}

			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.checkJSON {
				var response storage.GetOrderDetails

				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.result.ID, response.ID)
			}

			mockService.AssertExpectations(t)
		})
	}
}
