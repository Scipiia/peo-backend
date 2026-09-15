package update

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

type MockNormOrderUpdater struct {
	mock.Mock
}

func (m *MockNormOrderUpdater) UpdateNormOrder(ctx context.Context, ID int64, update storage.UpdateOrderDetails) error {
	return m.Called(ctx, ID, update).Error(0)
}

func TestUpdateNormOrderOperation(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		orderID int64

		body      storage.UpdateOrderDetails
		mockError error

		wantStatus   int
		expectedBody string
		checkJSON    bool
		needMock     bool
		invalidJSON  bool
	}{
		{
			name:    "OK",
			id:      "1",
			orderID: 1,
			body: storage.UpdateOrderDetails{
				OrderNum: "Q6-777",
				Name:     "abc",
				Operations: []storage.NormOperation{
					{Name: "Резка"},
				},
			},
			wantStatus:  http.StatusOK,
			checkJSON:   true,
			needMock:    true,
			invalidJSON: false,
		},
		{
			name:         "Invalid ID",
			id:           "abc",
			wantStatus:   http.StatusBadRequest,
			expectedBody: "Invalid ID",
			needMock:     false,
			invalidJSON:  false,
		},
		{
			name:         "Invalid JSON",
			id:           "1",
			orderID:      1,
			wantStatus:   http.StatusBadRequest,
			expectedBody: "Invalid data",
			needMock:     false,
			invalidJSON:  true,
		},
		{
			name:    "Server error",
			id:      "1",
			orderID: 1,
			body: storage.UpdateOrderDetails{
				OrderNum: "Q6-777",
				Name:     "abc",
				Operations: []storage.NormOperation{
					{Name: "Резка"},
				},
			},
			wantStatus:   http.StatusInternalServerError,
			expectedBody: "Server error",
			needMock:     true,
			invalidJSON:  false,
			mockError:    errors.New("server error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockNormOrderUpdater)
			if tt.needMock {
				mockService.On("UpdateNormOrder", mock.Anything, tt.orderID, tt.body).
					Return(tt.mockError)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Post("/api/orders/order/norm/update/{id}", UpdateNormOrderOperation(log, mockService))

			var req *http.Request

			if tt.invalidJSON {
				req = httptest.NewRequest(http.MethodPost, "/api/orders/order/norm/update/"+tt.id, strings.NewReader("{invalid json}"))
			} else {
				bodyBates, err := json.Marshal(tt.body)
				require.NoError(t, err)
				req = httptest.NewRequest(http.MethodPost, "/api/orders/order/norm/update/"+tt.id, bytes.NewReader(bodyBates))
			}

			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.checkJSON {
				var response int64

				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.orderID, response)
			}

			if tt.expectedBody == "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}

			if tt.needMock {
				mockService.AssertExpectations(t)
			}
		})
	}
}

type MockFinalOrderUpdater struct {
	mock.Mock
}

func (m *MockFinalOrderUpdater) UpdateFinalOrder(ctx context.Context, ID int64, update storage.UpdateFinalOrderDetails) error {
	return m.Called(ctx, ID, update).Error(0)
}

func TestUpdateFinalOrder_OK(t *testing.T) {

	var brigade = "окна и двери"

	updateReq := storage.UpdateFinalOrderDetails{
		Brigade: &brigade,
		ID:      1,
	}
	idStr := "1"

	mockService := new(MockFinalOrderUpdater)
	mockService.On("UpdateFinalOrder", mock.Anything, updateReq.ID, updateReq).Return(nil)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Put("/api/final/update/{id}", UpdateFinalOrder(log, mockService))

	bodyBytes, err := json.Marshal(updateReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/final/update/"+idStr, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	mockService.AssertExpectations(t)
}

func TestUpdateFinalOrder_InvalidJSON(t *testing.T) {

	idStr := "1"

	mockService := new(MockFinalOrderUpdater)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Put("/api/final/update/{id}", UpdateFinalOrder(log, mockService))

	req := httptest.NewRequest(http.MethodPut, "/api/final/update/"+idStr, bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid data")

	mockService.AssertNotCalled(t, "UpdateFinalOrder")
}

func TestUpdateFinalOrder_InvalidID(t *testing.T) {

	mockService := new(MockFinalOrderUpdater)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Put("/api/final/update/{id}", UpdateFinalOrder(log, mockService))

	req := httptest.NewRequest(http.MethodPut, "/api/final/update/abc", bytes.NewBufferString(`{"brigade":"окна двери"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid ID")

	mockService.AssertNotCalled(t, "UpdateFinalOrder")
}

func TestUpdateFinalOrder_InternalError(t *testing.T) {

	var brigade = "окна и двери"

	updateReq := storage.UpdateFinalOrderDetails{
		Brigade: &brigade,
		ID:      1,
	}
	idStr := "1"
	resultError := errors.New("database error")

	mockService := new(MockFinalOrderUpdater)
	mockService.On("UpdateFinalOrder", mock.Anything, updateReq.ID, updateReq).Return(resultError)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Put("/api/final/update/{id}", UpdateFinalOrder(log, mockService))

	bodyBytes, err := json.Marshal(updateReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/final/update/"+idStr, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Ошибка обновления")

	mockService.AssertExpectations(t)
}

type MockCancelStatusUpdater struct {
	mock.Mock
}

func (m *MockCancelStatusUpdater) UpdateStatus(ctx context.Context, rootProductID int64, status string) error {
	return m.Called(ctx, rootProductID, status).Error(0)
}

func TestUpdateCancelStatus_OK(t *testing.T) {
	var rootID int64 = 1
	status := "cancel"

	mockService := new(MockCancelStatusUpdater)
	mockService.On("UpdateStatus", mock.Anything, rootID, status).Return(nil)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Post("/api/orders/cancel", UpdateCancelStatus(log, mockService))

	reqBody := CancelStatusRequest{RootProductID: rootID}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/orders/cancel", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	mockService.AssertExpectations(t)
}

func TestUpdateCancelStatus_InvalidJSON(t *testing.T) {
	mockService := new(MockCancelStatusUpdater)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Post("/api/orders/cancel", UpdateCancelStatus(log, mockService))

	req := httptest.NewRequest(http.MethodPost, "/api/orders/cancel", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request payload")

	mockService.AssertNotCalled(t, "UpdateCancelStatus")
}

func TestUpdateCancelStatus_InternalError(t *testing.T) {
	var rootID int64 = 1
	status := "cancel"
	resultError := errors.New("database error")

	mockService := new(MockCancelStatusUpdater)
	mockService.On("UpdateStatus", mock.Anything, rootID, status).Return(resultError)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Post("/api/orders/cancel", UpdateCancelStatus(log, mockService))

	reqBody := CancelStatusRequest{RootProductID: rootID}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/orders/cancel", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to cancel order")

	mockService.AssertExpectations(t)
}
