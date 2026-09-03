package get

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"vue-golang/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrdersGetter struct {
	mock.Mock
}

func (m *MockOrdersGetter) GetOrdersMonth(ctx context.Context, year int, month int, search string) ([]*storage.Order, error) {
	args := m.Called(ctx, year, month, search)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.Order), args.Error(1)
}

func TestGetOrdersFilter(t *testing.T) {
	tests := []struct {
		name       string
		year       int
		yearStr    string
		month      int
		monthStr   string
		search     string
		mockResult []*storage.Order
		mockError  error
		wantStatus int
		wantBody   string
		needMock   bool
	}{
		{
			name:       "OK search",
			search:     "Q6-777",
			mockResult: []*storage.Order{{ID: 1}, {ID: 2}},
			wantStatus: http.StatusOK,
			needMock:   true,
		},
		{
			name:       "OK year month",
			year:       2026,
			month:      12,
			mockResult: []*storage.Order{{ID: 1}, {ID: 2}},
			wantStatus: http.StatusOK,
			needMock:   true,
		},
		{
			name:       "Missing year and month",
			year:       0,
			month:      0,
			search:     "",
			wantStatus: http.StatusBadRequest,
			needMock:   false,
		},
		{
			name:       "Invalid year",
			yearStr:    "abc",
			search:     "",
			wantStatus: http.StatusBadRequest,
			needMock:   false,
		},
		{
			name:       "Invalid month",
			year:       2026,
			monthStr:   "abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Service error",
			year:       2026,
			month:      12,
			mockError:  errors.New("database error"),
			wantStatus: http.StatusInternalServerError,
			needMock:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrdersGetter)
			if tt.needMock {
				mockService.On("GetOrdersMonth", mock.Anything, tt.year, tt.month, tt.search).
					Return(tt.mockResult, tt.mockError)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/orders", GetOrdersFilter(log, mockService))

			query := url.Values{}

			if tt.search != "" {
				query.Set("search", tt.search)
			}

			if tt.yearStr != "" {
				query.Set("year", tt.yearStr)
			} else if tt.year != 0 {
				query.Set("year", strconv.Itoa(tt.year))
			}

			if tt.monthStr != "" {
				query.Set("month", tt.monthStr)
			} else if tt.month != 0 {
				query.Set("month", strconv.Itoa(tt.month))
			}

			req := httptest.NewRequest(http.MethodGet, "/api/orders?"+query.Encode(), nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.mockResult != nil {
				var response []*storage.Order

				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.mockResult, response)
			}

			mockService.AssertExpectations(t)
		})
	}
}
