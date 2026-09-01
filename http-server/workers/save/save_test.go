package save

import (
	"bytes"
	"context"
	"encoding/json"
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

type MockResultWorkersGetter struct {
	mock.Mock
}

func (m *MockResultWorkersGetter) SaveOperationWorkers(ctx context.Context, req storage.SaveWorkers) error {
	return m.Mock.Called(ctx, req).Error(0)
}

func TestSaveWorkersOperation(t *testing.T) {
	tests := []struct {
		name        string
		body        storage.SaveWorkers
		mockError   error
		wantStatus  int
		invalidJSON bool
		checkJSON   bool
		needMock    bool
	}{
		{
			name: "OK",
			body: storage.SaveWorkers{
				UpdateStatus: "assigned",
				Assignments: []storage.OperationWorkers{
					{
						ProductID:     1,
						OperationName: "Резка",
						EmployeeID:    2,
						ActualMinutes: 10,
					},
				},
			},
			wantStatus:  http.StatusOK,
			invalidJSON: false,
			needMock:    true,
		},
		{
			name: "Invalid json",
			body: storage.SaveWorkers{
				UpdateStatus: "assigned",
				Assignments: []storage.OperationWorkers{
					{
						ProductID:     1,
						OperationName: "Резка",
						EmployeeID:    2,
						ActualMinutes: 10,
					},
				},
			},
			wantStatus:  http.StatusBadRequest,
			invalidJSON: true,
			needMock:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockResultWorkersGetter)
			if tt.needMock {
				mockService.On("SaveOperationWorkers", mock.Anything, tt.body).
					Return(tt.mockError)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Post("/api/workers", SaveWorkersOperation(log, mockService))

			bodyBates, err := json.Marshal(tt.body)
			require.NoError(t, err)

			var req *http.Request

			if tt.invalidJSON {
				req = httptest.NewRequest(http.MethodPost, "/api/workers", strings.NewReader("{invalid"))
			} else {
				req = httptest.NewRequest(http.MethodPost, "/api/workers", bytes.NewReader(bodyBates))
			}
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			require.Equal(t, tt.wantStatus, w.Code)

			if tt.checkJSON {
				var response map[string]interface{}

				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, "success", response["status"])
				assert.Equal(t, float64(1), response["saved"])
			}

			mockService.AssertExpectations(t)
		})
	}
}
