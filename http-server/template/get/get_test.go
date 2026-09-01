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

type MockTemplateByCodeGetter struct {
	mock.Mock
}

func (m *MockTemplateByCodeGetter) GetTemplateByCode(ctx context.Context, code string) (*storage.Template, error) {
	args := m.Called(ctx, code)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*storage.Template), args.Error(1)
}

func TestGetTemplatesByCode(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		wantStatus int
		mockResult *storage.Template
		mockError  error
		wantBody   string
		needMock   bool
	}{
		{
			name:       "OK",
			code:       "1",
			wantStatus: http.StatusOK,
			mockResult: &storage.Template{Code: "1"},
			needMock:   true,
		},
		{
			name:       "Service error",
			code:       "1",
			wantStatus: http.StatusInternalServerError,
			mockError:  errors.New("service error"),
			wantBody:   "Internal server error",
			needMock:   true,
		},
		{
			name:       "empty code",
			code:       "",
			wantStatus: http.StatusBadRequest,
			mockError:  errors.New("empty code"),
			wantBody:   "Missing required query parameter",
			needMock:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockTemplateByCodeGetter)
			if tt.needMock {
				mockService.On("GetTemplateByCode", mock.Anything, tt.code).
					Return(tt.mockResult, tt.mockError)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/template", GetTemplatesByCode(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/template?code="+tt.code, nil)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}

			if tt.mockResult != nil {
				var response storage.Template

				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, tt.mockResult.Code, response.Code)
			}

			mockService.AssertExpectations(t)
		})
	}
}

type MockAllTemplatesGetter struct {
	mock.Mock
}

func (m *MockAllTemplatesGetter) GetAllTemplates(ctx context.Context) ([]*storage.Template, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*storage.Template), args.Error(1)
}

func TestGetAllTemplates(t *testing.T) {
	tests := []struct {
		name       string
		mockResult []*storage.Template
		mockError  error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "OK",
			mockResult: []*storage.Template{{Code: "1"}, {Code: "2"}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Internal error",
			mockError:  errors.New("database error"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Internal server error",
		},
		{
			name:       "Empty result",
			mockResult: []*storage.Template{},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockAllTemplatesGetter)
			mockService.On("GetAllTemplates", mock.Anything).
				Return(tt.mockResult, tt.mockError)

			log := slog.New(slog.NewTextHandler(io.Discard, nil))

			r := chi.NewRouter()
			r.Get("/api/templates", GetAllTemplates(log, mockService))

			req := httptest.NewRequest(http.MethodGet, "/api/templates", nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.mockResult != nil {
				var response []*storage.Template

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
