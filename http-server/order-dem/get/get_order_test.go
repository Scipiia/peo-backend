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

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrderDetailsGetter struct {
	mock.Mock
}

func (m *MockOrderDetailsGetter) GetOrderDetails(ctx context.Context, orderNum string) ([]*storage.ResultOrderDetails, error) {
	args := m.Called(ctx, orderNum)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*storage.ResultOrderDetails), args.Error(1)
}

func TestGetOrderDetails(t *testing.T) {
	tests := []struct {
		name       string
		orderNum   string
		mockResult []*storage.ResultOrderDetails
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "OK",
			orderNum:   "Q6-123",
			mockResult: []*storage.ResultOrderDetails{{ID: 1}, {ID: 2}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Service error",
			orderNum:   "Q6-123",
			mockError:  errors.New("service error"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrderDetailsGetter)
			mockService.On("GetOrderDetails", mock.Anything, tt.orderNum).
				Return(tt.mockResult, tt.mockError)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/orders/order/{orderNum}", GetOrderDetails(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/orders/order/"+tt.orderNum, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.mockResult != nil {
				var response []*storage.ResultOrderDetails

				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tt.mockResult, response)
			}

			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}

			mockService.AssertExpectations(t)
		})
	}
}
