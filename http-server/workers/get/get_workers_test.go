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
)

type MockWorkersGetter struct {
	mock.Mock
}

func (m *MockWorkersGetter) GetAllWorkers(ctx context.Context, typeIzd string) ([]storage.GetWorkers, error) {
	args := m.Called(ctx, typeIzd)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]storage.GetWorkers), args.Error(1)
}

func TestGetWorkers(t *testing.T) {
	tests := []struct {
		name       string
		typeIzd    string
		mockResult []storage.GetWorkers
		wantStatus int

		mockError error
		wantBody  string
		needMock  bool
	}{
		{
			name:    "OK",
			typeIzd: "door",
			mockResult: []storage.GetWorkers{
				{
					ID:   1,
					Name: "Ivan",
				},
				{
					ID:   2,
					Name: "Petr",
				},
			},
			wantStatus: http.StatusOK,
			needMock:   true,
		},
		{
			name:       "Service error",
			typeIzd:    "door",
			wantStatus: http.StatusInternalServerError,
			mockError:  errors.New("database error"),
			wantBody:   "Внутренняя ошибка сервера",
			needMock:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockWorkersGetter)
			if tt.needMock {
				mockService.On("GetAllWorkers", mock.Anything, tt.typeIzd).
					Return(tt.mockResult, tt.mockError)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/workers/all", GetWorkers(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/workers/all?type="+tt.typeIzd, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.mockResult != nil {
				var response []storage.GetWorkers

				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, tt.mockResult, response)
			}

			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}

			mockService.AssertExpectations(t)
		})
	}
}
