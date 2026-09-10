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
	"testing"
	"vue-golang/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAdminCoefUpdater struct {
	mock.Mock
}

func (m *MockAdminCoefUpdater) UpdateCoefficientPEOAdmin(ctx context.Context, coeffs []storage.CoefficientPEOAdmin) error {
	return m.Called(ctx, coeffs).Error(0)
}

func TestUpdateCoefficientAdmin_OK(t *testing.T) {
	coeffs := []storage.CoefficientPEOAdmin{
		{
			Type:        "window",
			Coefficient: 99.9,
		},
	}

	mockService := new(MockAdminCoefUpdater)
	mockService.On("UpdateCoefficientPEOAdmin", mock.Anything, coeffs).Return(nil)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Put("/api/admin/coefficient/update", UpdateCoefficientAdmin(log, mockService))

	bodyBytes, err := json.Marshal(coeffs)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/coefficient/update", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestUpdateCoefficientAdmin_InvalidJSON(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockService := new(MockAdminCoefUpdater)

	r := chi.NewRouter()
	r.Put("/api/admin/coefficient/update", UpdateCoefficientAdmin(log, mockService))

	req := httptest.NewRequest(http.MethodPut, "/api/admin/coefficient/update", bytes.NewBufferString(`{{`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Неверный JSON")

	mockService.AssertNotCalled(t, "MockAdminCoefUpdater")
}

func TestUpdateCoefficientAdmin_InternalError(t *testing.T) {

	coeffs := []storage.CoefficientPEOAdmin{
		{
			Type:        "window",
			Coefficient: 99.9,
		},
	}
	resultError := errors.New("database error")

	mockService := new(MockAdminCoefUpdater)
	mockService.On("UpdateCoefficientPEOAdmin", mock.Anything, coeffs).Return(resultError)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Put("/api/admin/coefficient/update", UpdateCoefficientAdmin(log, mockService))

	bodyBytes, err := json.Marshal(coeffs)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/coefficient/update", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

type MockAdminEmployeesUpdater struct {
	mock.Mock
}

func (m *MockAdminEmployeesUpdater) UpdateAllEmployeesAdmin(ctx context.Context, id int64, input storage.UpdateEmployeeInput) error {
	return m.Called(ctx, id, input).Error(0)
}

func TestUpdateEmployeesAdmin_OK(t *testing.T) {
	input := storage.UpdateEmployeeInput{
		Name:     "Ivan",
		IsActive: true,
	}

	var idStr = "1"
	var id int64 = 1

	mockService := new(MockAdminEmployeesUpdater)
	mockService.On("UpdateAllEmployeesAdmin", mock.Anything, id, input).Return(nil)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Put("/api/admin/employees/update/{id}", UpdateEmployeesAdmin(log, mockService))

	bodyBytes, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/employees/update/"+idStr, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	mockService.AssertExpectations(t)
}

func TestUpdateEmployeesAdmin_InvalidID(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockService := new(MockAdminEmployeesUpdater)

	r := chi.NewRouter()
	r.Put("/api/admin/employees/update/{id}", UpdateEmployeesAdmin(log, mockService))

	req := httptest.NewRequest(http.MethodPut, "/api/admin/employees/update/abc",
		bytes.NewBufferString(`{"name":"Ivan","is_active":true}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "неверный ID")

	mockService.AssertNotCalled(t, "UpdateAllEmployeesAdmin")
}

func TestUpdateEmployeesAdmin_InvalidJSON(t *testing.T) {
	var idStr = "1"

	mockService := new(MockAdminEmployeesUpdater)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Put("/api/admin/employees/update/{id}", UpdateEmployeesAdmin(log, mockService))

	req := httptest.NewRequest(http.MethodPut, "/api/admin/employees/update/"+idStr, bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Неверный JSON")

	mockService.AssertNotCalled(t, "UpdateEmployeesAdmin")
}

func TestUpdateEmployeesAdmin_InternalError(t *testing.T) {
	input := storage.UpdateEmployeeInput{
		Name:     "Ivan",
		IsActive: true,
	}

	resultError := errors.New("database error")

	var idStr = "1"
	var id int64 = 1

	mockService := new(MockAdminEmployeesUpdater)
	mockService.On("UpdateAllEmployeesAdmin", mock.Anything, id, input).Return(resultError)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Put("/api/admin/employees/update/{id}", UpdateEmployeesAdmin(log, mockService))

	bodyBytes, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/employees/update/"+idStr, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Ошибка сервера")

	mockService.AssertExpectations(t)
}
