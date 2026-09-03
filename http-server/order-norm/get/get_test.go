package get

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"vue-golang/internal/storage"
	"vue-golang/internal/storage/mysql"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrderGetter struct {
	mock.Mock
}

func (m *MockOrderGetter) GetNormOrder(ctx context.Context, id int64) (*storage.GetOrderDetails, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.GetOrderDetails), args.Error(1)
}

func TestGetNormOrder(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		orderID    int64
		mockResult *storage.GetOrderDetails
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "OK",
			id:         "10",
			orderID:    10,
			mockResult: &storage.GetOrderDetails{ID: 10},
			mockError:  nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "Not found",
			id:         "10",
			orderID:    10,
			mockResult: nil,
			mockError:  errors.New("нормировка не найдена"),
			wantStatus: http.StatusNotFound,
			wantBody:   "Нормировка не найдена",
		},
		{
			name:       "Internal error",
			id:         "10",
			orderID:    10,
			mockResult: nil,
			mockError:  errors.New("database error"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Внутренняя ошибка",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrderGetter)
			mockService.On("GetNormOrder", mock.Anything, tt.orderID).
				Return(tt.mockResult, tt.mockError)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/orders/order/norm/{id}", GetNormOrder(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/orders/order/norm/"+tt.id, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}

			mockService.AssertExpectations(t)
		})
	}
}

type MockOrderByOrderNumGetter struct {
	mock.Mock
}

func (m *MockOrderByOrderNumGetter) GetNormOrdersByOrderNum(ctx context.Context, orderNum string, position int) ([]*storage.GetOrderDetails, error) {
	args := m.Called(ctx, orderNum, position)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*storage.GetOrderDetails), args.Error(1)
}

func TestGetNormOrdersOrderNum(t *testing.T) {
	tests := []struct {
		name       string
		orderNum   string
		position   string
		mockResult []*storage.GetOrderDetails
		mockError  error
		wantStatus int
	}{
		{
			name:       "OK",
			orderNum:   "123",
			position:   "1",
			mockResult: []*storage.GetOrderDetails{{ID: 1}, {ID: 2}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Internal error",
			orderNum:   "123",
			position:   "1",
			mockResult: nil,
			mockError:  errors.New("database error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrderByOrderNumGetter)
			mockService.On("GetNormOrdersByOrderNum", mock.Anything, tt.orderNum, 1).
				Return(tt.mockResult, tt.mockError)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/orders/order-norm/by-order", GetNormOrdersOrderNum(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/orders/order-norm/by-order?order_num="+tt.orderNum+"&position="+tt.position, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.mockResult != nil {
				var response []*storage.GetOrderDetails

				err := json.Unmarshal(w.Body.Bytes(), &response)

				assert.NoError(t, err)

				assert.Len(t, response, len(tt.mockResult))
			}

			mockService.AssertExpectations(t)
		})
	}
}

type MockOrdersGetter struct {
	mock.Mock
}

func (m *MockOrdersGetter) GetNormOrders(ctx context.Context, orderNum, orderType string) ([]storage.GetOrderDetails, error) {
	args := m.Called(ctx, orderNum, orderType)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]storage.GetOrderDetails), args.Error(1)
}

func TestGetNormOrders(t *testing.T) {
	tests := []struct {
		name       string
		orderNum   string
		orderType  string
		mockResult []storage.GetOrderDetails
		mockError  error
		wantStatus int
	}{
		{
			name:       "OK",
			orderNum:   "123",
			orderType:  "window",
			mockResult: []storage.GetOrderDetails{{ID: 1}, {ID: 2}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Internal error",
			orderNum:   "123",
			orderType:  "window",
			mockResult: nil,
			mockError:  errors.New("database error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrdersGetter)
			mockService.On("GetNormOrders", mock.Anything, tt.orderNum, tt.orderType).
				Return(tt.mockResult, tt.mockError)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/orders/order/norm/all", GetNormOrders(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/orders/order/norm/all?order_num="+tt.orderNum+"&type="+tt.orderType, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.mockResult != nil {
				var response []storage.GetOrderDetails

				err := json.Unmarshal(w.Body.Bytes(), &response)

				require.NoError(t, err)

				assert.Equal(t, tt.mockResult, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

type MockOrderDoubleGetter struct {
	mock.Mock
}

func (m *MockOrderDoubleGetter) GetNormOrderIdSub(ctx context.Context, id int64) ([]*storage.GetOrderDetails, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*storage.GetOrderDetails), args.Error(1)
}

func (m *MockOrderDoubleGetter) GetMosquitoOrderDetails(ctx context.Context, requestedID int64) (*storage.GetOrderDetails, error) {
	args := m.Called(ctx, requestedID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.GetOrderDetails), args.Error(1)
}

func (m *MockOrderDoubleGetter) GetGutterOrderDetails(ctx context.Context, requestedID int64) (*storage.GetOrderDetails, error) {
	args := m.Called(ctx, requestedID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.GetOrderDetails), args.Error(1)
}

func TestDoubleReportOrder(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		orderID        int64
		source         string
		subResult      []*storage.GetOrderDetails
		mosquitoResult *storage.GetOrderDetails
		gutterResult   *storage.GetOrderDetails

		subErr      error
		mosquitoErr error
		gutterErr   error

		wantStatus int
	}{
		{
			name:       "default OK",
			id:         "10",
			orderID:    10,
			source:     "window",
			subResult:  []*storage.GetOrderDetails{{ID: 1}, {ID: 2}},
			wantStatus: http.StatusOK,
		},
		{
			name:           "Mosquito OK",
			id:             "10",
			orderID:        10,
			source:         "mosquito",
			mosquitoResult: &storage.GetOrderDetails{ID: 1},
			wantStatus:     http.StatusOK,
		},
		{
			name:         "Gutter OK",
			id:           "10",
			orderID:      10,
			source:       "vodootliv",
			gutterResult: &storage.GetOrderDetails{ID: 4},
			wantStatus:   http.StatusOK,
		},
		{
			name:       "invalid id",
			id:         "invalid",
			source:     "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "default error",
			id:         "10",
			orderID:    10,
			source:     "window",
			subErr:     errors.New("database error"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:        "mosquito error",
			id:          "10",
			orderID:     10,
			source:      "mosquito",
			mosquitoErr: errors.New("database error"),
			wantStatus:  http.StatusInternalServerError,
		},
		{
			name:       "gutter error",
			id:         "10",
			orderID:    10,
			source:     "vodootliv",
			gutterErr:  errors.New("database error"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "calculator required",
			id:         "10",
			orderID:    10,
			source:     "vodootliv",
			gutterErr:  errors.New("REQUIRES_CALCULATOR"),
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrderDoubleGetter)

			if tt.wantStatus != http.StatusBadRequest {
				switch tt.source {
				case "mosquito":
					mockService.On("GetMosquitoOrderDetails", mock.Anything, tt.orderID).
						Return(tt.mosquitoResult, tt.mosquitoErr)
				case "vodootliv":
					mockService.On("GetGutterOrderDetails", mock.Anything, tt.orderID).
						Return(tt.gutterResult, tt.gutterErr)
				default:
					mockService.On("GetNormOrderIdSub", mock.Anything, tt.orderID).
						Return(tt.subResult, tt.subErr)
				}
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/orders/order-norm/{id}/details", DoubleReportOrder(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/orders/order-norm/"+tt.id+"/details?source="+tt.source, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus != http.StatusBadRequest {
				mockService.AssertExpectations(t)
			}
		})
	}
}

type MockOrderFinalGetter struct {
	mock.Mock
}

func (m *MockOrderFinalGetter) GetSimpleOrderReport(ctx context.Context, orderNum string) (*storage.OrderFinalReport, error) {
	args := m.Called(ctx, orderNum)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.OrderFinalReport), args.Error(1)
}

func TestFinalReportNormOrder(t *testing.T) {
	tests := []struct {
		name       string
		orderNum   string
		result     *storage.OrderFinalReport
		wantStatus int
		mockError  error
		checkJSON  bool
	}{
		{
			name:       "OK",
			orderNum:   "Q6-777",
			result:     &storage.OrderFinalReport{OrderNum: "Q6-777"},
			wantStatus: http.StatusOK,
			mockError:  nil,
			checkJSON:  true,
		},
		{
			name:       "error",
			orderNum:   "Q6-777",
			result:     nil,
			wantStatus: http.StatusInternalServerError,
			mockError:  errors.New("database error"),
			checkJSON:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrderFinalGetter)
			mockService.On("GetSimpleOrderReport", mock.Anything, tt.orderNum).
				Return(tt.result, tt.mockError)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/allians/{order_num}", FinalReportNormOrder(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/allians/"+tt.orderNum, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.checkJSON {
				var response storage.OrderFinalReport
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.result.OrderNum, response.OrderNum)
			}
			mockService.AssertExpectations(t)
		})
	}
}

type MockOrdersPEOGetter struct {
	mock.Mock
}

func (m *MockOrdersPEOGetter) GetPEOProductsByCategory(ctx context.Context, filter mysql.ProductFilter) ([]storage.PEOProduct, []storage.GetWorkers, error) {
	args := m.Called(ctx, filter)

	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}

	return args.Get(0).([]storage.PEOProduct), args.Get(1).([]storage.GetWorkers), args.Error(2)
}

func TestFinalReportNormOrders(t *testing.T) {
	tests := []struct {
		name  string
		query string

		expectedOrderNum string
		expectedType     []string

		products  []storage.PEOProduct
		employees []storage.GetWorkers
		mockError error

		wantStatus   int
		checkJSON    bool
		needMock     bool
		expectedBody string
	}{
		{
			name:             "OK",
			query:            "?order_num=Q6-777&type=window&type=door",
			expectedOrderNum: "Q6-777",
			expectedType:     []string{"window", "door"},
			products:         []storage.PEOProduct{{}},
			employees:        []storage.GetWorkers{{}},
			wantStatus:       http.StatusOK,
			needMock:         true,
			checkJSON:        true,
		},
		{
			name:             "error",
			query:            "?order_num=Q6-777&type=window&type=door",
			expectedOrderNum: "Q6-777",
			expectedType:     []string{"window", "door"},
			products:         nil,
			employees:        nil,
			wantStatus:       http.StatusInternalServerError,
			needMock:         true,
			mockError:        errors.New("database error"),
			checkJSON:        false,
		},
		{
			name:         "invalid from data",
			query:        "?from=abc",
			wantStatus:   http.StatusBadRequest,
			needMock:     false,
			checkJSON:    false,
			expectedBody: "Неверный формат даты 'from'",
		},
		{
			name:         "invalid to data",
			query:        "?to=abc",
			wantStatus:   http.StatusBadRequest,
			needMock:     false,
			checkJSON:    false,
			expectedBody: "Неверный формат даты 'to'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrdersPEOGetter)

			if tt.needMock {
				mockService.On("GetPEOProductsByCategory", mock.Anything,
					mock.MatchedBy(func(filter mysql.ProductFilter) bool {
						if filter.OrderNum != tt.expectedOrderNum {
							return false
						}
						if len(filter.Type) != len(tt.expectedType) {
							return false
						}
						for i := range filter.Type {
							if filter.Type[i] != tt.expectedType[i] {
								return false
							}
						}
						return true
					})).Return(tt.products, tt.employees, tt.mockError)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/all_final_order", FinalReportNormOrders(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/all_final_order"+tt.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}

			if tt.checkJSON {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "products")
				assert.Contains(t, response, "employees")
			}

			if tt.needMock {
				mockService.AssertExpectations(t)
			}
		})
	}
}

type MockOrderNashelGetter struct {
	mock.Mock
}

func (m *MockOrderNashelGetter) GetNashchelnikRawData(ctx context.Context, legacyID int64) (*storage.NashchelnikRawData, error) {
	args := m.Called(ctx, legacyID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.NashchelnikRawData), args.Error(1)
}

func TestGetNashchelnikRawHandler(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		orderId int64

		result *storage.NashchelnikRawData

		expectedBody string
		mockError    error
		wantStatus   int
		checkJSON    bool
		needMock     bool
	}{
		{
			name:    "OK",
			id:      "1",
			orderId: 1,
			result: &storage.NashchelnikRawData{
				OrderNum: "Q6-777",
			},
			wantStatus: http.StatusOK,
			checkJSON:  true,
			needMock:   true,
		},
		{
			name:         "error",
			id:           "abc",
			wantStatus:   http.StatusBadRequest,
			expectedBody: "Invalid ID",
			needMock:     false,
		},
		{
			name:         "not found",
			id:           "1",
			orderId:      1,
			wantStatus:   http.StatusNotFound,
			mockError:    errors.New("not found"),
			expectedBody: "Not found",
			needMock:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrderNashelGetter)

			mockService.On("GetNashchelnikRawData", mock.Anything, tt.orderId).
				Return(tt.result, tt.mockError)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/orders/nashchelnik/raw/{id}", GetNashchelnikRawHandler(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/orders/nashchelnik/raw/"+tt.id, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}

			if tt.checkJSON {
				var response storage.NashchelnikRawData
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.result.OrderNum, response.OrderNum)
			}

			if tt.needMock {
				mockService.AssertExpectations(t)
			}
		})
	}
}

type MockOrderVitrageGetter struct {
	mock.Mock
}

func (m *MockOrderVitrageGetter) GetNormOrderVitrage(ctx context.Context, id int64) ([]storage.GetWorkersVitrage, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]storage.GetWorkersVitrage), args.Error(1)
}

func TestGetVitrageAssignments(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		orderId int64

		result []storage.GetWorkersVitrage

		expectedBody string
		mockError    error
		wantStatus   int
		checkJSON    bool
		needMock     bool
	}{
		{
			name:       "OK",
			id:         "1",
			orderId:    1,
			result:     []storage.GetWorkersVitrage{{ID: 1}},
			wantStatus: http.StatusOK,
			checkJSON:  true,
			needMock:   true,
		},
		{
			name:         "invalid id",
			id:           "abc",
			wantStatus:   http.StatusBadRequest,
			needMock:     false,
			expectedBody: "неверный id заказа",
		},
		{
			name:         "service error",
			id:           "1",
			orderId:      1,
			mockError:    errors.New("db error"),
			wantStatus:   http.StatusInternalServerError,
			needMock:     true,
			expectedBody: "ошибка получения назначений",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrderVitrageGetter)

			if tt.needMock {
				mockService.On("GetNormOrderVitrage", mock.Anything, tt.orderId).
					Return(tt.result, tt.mockError)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/orders/{id}/vitr-assign", GetVitrageAssignments(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/orders/"+tt.id+"/vitr-assign", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}

			if tt.checkJSON {
				var response []storage.GetWorkersVitrage
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.result, response)
			}

			if tt.needMock {
				mockService.AssertExpectations(t)
			}
		})
	}
}
