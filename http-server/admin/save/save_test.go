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
	"testing"
	"vue-golang/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockEmployeesProvider struct {
	mock.Mock
}

func (m *MockEmployeesProvider) CreateEmployeeAdmin(ctx context.Context, input storage.CreateEmployeeInput) error {
	args := m.Called(ctx, input)

	return args.Error(0)
}

func TestSaveEmployerAdmin_OK(t *testing.T) {

	input := storage.CreateEmployeeInput{
		Name:    "Ivanov",
		TeamIDs: []int{1, 2},
	}

	mockService := new(MockEmployeesProvider)
	mockService.On("CreateEmployeeAdmin", mock.Anything, input).Return(nil)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Post("/api/admin/employees/save", SaveEmployerAdmin(log, mockService))

	bodyBytes, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/employees/save", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestSaveEmployerAdmin_InvalidJSON(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockService := new(MockEmployeesProvider)

	r := chi.NewRouter()
	r.Post("/api/admin/employees/save", SaveEmployerAdmin(log, mockService))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/employees/save", bytes.NewBufferString(`{"name":"Ivanov",`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Неверный JSON")

	mockService.AssertNotCalled(t, "CreateEmployeeAdmin")
}

func TestSaveEmployerAdmin_InternalError(t *testing.T) {

	input := storage.CreateEmployeeInput{
		Name:    "Ivanov",
		TeamIDs: []int{1, 2},
	}

	resultError := errors.New("database error")

	mockService := new(MockEmployeesProvider)
	mockService.On("CreateEmployeeAdmin", mock.Anything, input).Return(resultError)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Post("/api/admin/employees/save", SaveEmployerAdmin(log, mockService))

	bodyBytes, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/employees/save", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}
