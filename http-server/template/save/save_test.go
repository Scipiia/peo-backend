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

type MockTemplateCreateProvider struct {
	mock.Mock
}

func (m *MockTemplateCreateProvider) CreateTemplateAdmin(ctx context.Context, res storage.TemplateAdmin) error {
	return m.Mock.Called(ctx, res).Error(0)
}

func TestSaveTemplateAdmin(t *testing.T) {
	tests := []struct {
		name        string
		body        map[string]interface{}
		mockError   error
		wantStatus  int
		invalidJSON bool
		needMock    bool
	}{
		{
			name: "OK",
			body: map[string]interface{}{
				"code":       "1",
				"category":   "test",
				"profile":    "test",
				"operations": []storage.Operation{},
				"rules":      []storage.Rule{},
			},
			wantStatus:  http.StatusOK,
			invalidJSON: false,
			needMock:    true,
		},
		{
			name: "Create template error",
			body: map[string]interface{}{
				"code":       "1",
				"category":   "test",
				"profile":    "test",
				"operations": []storage.Operation{},
				"rules":      []storage.Rule{},
			},
			mockError:   errors.New("database error"),
			wantStatus:  http.StatusInternalServerError,
			invalidJSON: false,
			needMock:    true,
		},
		{
			name:        "Invalid json",
			wantStatus:  http.StatusBadRequest,
			invalidJSON: true,
			needMock:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTemplateCreateProvider)
			if tt.needMock {
				mockService.On("CreateTemplateAdmin", mock.Anything, mock.MatchedBy(func(res storage.TemplateAdmin) bool {
					return res.Code == tt.body["code"].(string) && res.Category == tt.body["category"].(string) && res.Profile == tt.body["profile"].(string)
				})).Return(tt.mockError)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Post("/template/new", SaveTemplateAdmin(log, mockService))

			bodyBates, err := json.Marshal(tt.body)
			require.NoError(t, err)

			var req *http.Request

			if tt.invalidJSON {
				req = httptest.NewRequest(http.MethodPost, "/template/new", strings.NewReader("{invalid"))
			} else {
				req = httptest.NewRequest(http.MethodPost, "/template/new", bytes.NewReader(bodyBates))
			}

			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}

				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, "created", response["status"])
			}

			if tt.mockError != nil {
				assert.Contains(t, w.Body.String(), "ошибка создания шаблона")
			}

			if tt.invalidJSON {
				assert.Contains(t, w.Body.String(), "ошибка парсинга JSON")
			}

			mockService.AssertExpectations(t)
		})
	}
}
