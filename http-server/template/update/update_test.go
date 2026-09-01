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

type MockUpdateTemplateProvider struct {
	mock.Mock
}

func (m *MockUpdateTemplateProvider) UpdateTemplateAdmin(ctx context.Context, id int, update storage.TemplateAdmin) error {
	return m.Called(ctx, id, update).Error(0)
}

func TestUpdateTemplateAdmin(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		orderID int
		body    map[string]interface{}

		wantStatus   int
		expectedBody string
		invalidJSON  bool
		needMock     bool
		mockError    error
	}{
		{
			name:    "OK",
			id:      "1",
			orderID: 1,
			body: map[string]interface{}{
				"name":       "abc",
				"code":       "1",
				"category":   "test",
				"operations": []storage.Operation{},
			},
			needMock:   true,
			wantStatus: http.StatusOK,
		},
		{
			name:    "Error",
			id:      "1",
			orderID: 1,
			body: map[string]interface{}{
				"name":       "abc",
				"code":       "1",
				"category":   "test",
				"operations": []storage.Operation{},
			},
			needMock:   true,
			wantStatus: http.StatusInternalServerError,
			mockError:  errors.New("database error"),
		},
		{
			name:        "Invalid json",
			id:          "1",
			wantStatus:  http.StatusBadRequest,
			invalidJSON: true,
			needMock:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUpdateTemplateProvider)
			if tt.needMock {
				mockService.On("UpdateTemplateAdmin", mock.Anything, tt.orderID, mock.MatchedBy(
					func(res storage.TemplateAdmin) bool {
						return res.Name == "abc" && res.Code == "1" && res.Category == "test"
					})).
					Return(tt.mockError)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Post("/template/update/{id}", UpdateTemplateAdmin(log, mockService))

			bodyBates, err := json.Marshal(tt.body)
			require.NoError(t, err)

			var req *http.Request

			if tt.invalidJSON {
				req = httptest.NewRequest(http.MethodPost, "/template/update/"+tt.id, strings.NewReader("{invalid"))
			} else {
				req = httptest.NewRequest(http.MethodPost, "/template/update/"+tt.id, bytes.NewReader(bodyBates))
			}

			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			mockService.AssertExpectations(t)
		})
	}
}
