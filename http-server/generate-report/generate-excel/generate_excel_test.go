package generate_excel

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"vue-golang/internal/storage/mysql"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

type GenerateExcelHandlerMock struct {
	mock.Mock
}

func (m *GenerateExcelHandlerMock) GenerateExcel(ctx context.Context, filter mysql.ProductFilter) ([]byte, error) {
	args := m.Mock.Called(ctx, filter)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}

func TestGenerateReportExcel(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		mockResult []byte
		mockError  error
		wantStatus int
		needMock   bool
	}{
		{
			name:       "OK",
			query:      "?from=2023-01-01&to=2023-01-02&order_num=Q6-777&type=door",
			mockResult: []byte("excel content"),
			wantStatus: http.StatusOK,
			needMock:   true,
		},
		{
			name:       "Invalid from",
			query:      "?from=invalid-date",
			wantStatus: http.StatusBadRequest,
			needMock:   false,
		},
		{
			name:       "Invalid to",
			query:      "?to=invalid-date",
			wantStatus: http.StatusBadRequest,
			needMock:   false,
		},
		{
			name:       "Server error",
			query:      "?from=2023-01-01&to=2023-01-02&order_num=Q6-777&type=door",
			wantStatus: http.StatusInternalServerError,
			mockError:  errors.New("database error"),
			needMock:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(GenerateExcelHandlerMock)
			if tt.needMock {
				mockService.On("GenerateExcel", mock.Anything, mock.MatchedBy(
					func(filter mysql.ProductFilter) bool {
						return filter.OrderNum == "Q6-777" && len(filter.Type) == 1 && filter.Type[0] == "door"
					})).
					Return(tt.mockResult, tt.mockError)
			}
			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/report/excel", GenerateReportExcel(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/report/excel"+tt.query, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				assert.Equal(
					t,
					"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
					w.Header().Get("Content-Type"),
				)

				assert.Equal(t, tt.mockResult, w.Body.Bytes())
			}

			mockService.AssertExpectations(t)
		})
	}
}
