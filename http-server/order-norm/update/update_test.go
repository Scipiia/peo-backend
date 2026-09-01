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
	return m.Mock.Called(ctx, ID, update).Error(0)
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
				var response map[string]interface{}

				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, "200", response["status"])
				assert.Equal(t, float64(tt.orderID), response["norm_id"])
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
