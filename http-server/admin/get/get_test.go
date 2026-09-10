package get

import (
	"context"
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

type MockAdminCoefGetter struct {
	mock.Mock
}

func (m *MockAdminCoefGetter) GetAllCoefficientAdmin(ctx context.Context) ([]*storage.CoefficientPEOAdmin, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*storage.CoefficientPEOAdmin), args.Error(1)
}

func TestGetCoefficientAdmin_OK(t *testing.T) {
	result := []*storage.CoefficientPEOAdmin{
		{
			Type:        "window",
			Coefficient: 1.2,
		},
	}

	mockService := new(MockAdminCoefGetter)
	mockService.On("GetAllCoefficientAdmin", mock.Anything).Return(result, nil)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Get("/api/admin/coefficient", GetCoefficientAdmin(log, mockService))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/coefficient", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "window")

	mockService.AssertExpectations(t)
}

func TestGetCoefficientAdmin_InternalError(t *testing.T) {
	resultError := errors.New("database error")

	mockService := new(MockAdminCoefGetter)
	mockService.On("GetAllCoefficientAdmin", mock.Anything).Return(nil, resultError)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Get("/api/admin/coefficient", GetCoefficientAdmin(log, mockService))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/coefficient", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

type MockAllEmployeesAdminGetter struct {
	mock.Mock
}

func (m *MockAllEmployeesAdminGetter) GetAllEmployeesAdmin(ctx context.Context) ([]*storage.EmployeesAdmin, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*storage.EmployeesAdmin), args.Error(1)
}

func TestGetAllEmployeesAdmin_OK(t *testing.T) {
	result := []*storage.EmployeesAdmin{
		{
			Name: "Ivanov",
		},
	}

	mockService := new(MockAllEmployeesAdminGetter)
	mockService.On("GetAllEmployeesAdmin", mock.Anything).Return(result, nil)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Get("/api/admin/employees", GetAllEmployeesAdmin(log, mockService))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/employees", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Ivanov")

	mockService.AssertExpectations(t)
}

func TestGetAllEmployeesAdmin_InternalError(t *testing.T) {
	resultError := errors.New("database error")

	mockService := new(MockAllEmployeesAdminGetter)
	mockService.On("GetAllEmployeesAdmin", mock.Anything).Return(nil, resultError)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Get("/api/admin/employees", GetAllEmployeesAdmin(log, mockService))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/employees", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}

type MockAdminAllTeamsGetter struct {
	mock.Mock
}

func (m *MockAdminAllTeamsGetter) GetAllTeamsAdmin(ctx context.Context) ([]*storage.TeamAdmin, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*storage.TeamAdmin), args.Error(1)
}

func TestGetAllTeams_OK(t *testing.T) {
	result := []*storage.TeamAdmin{
		{
			Name: "окна и двери",
			Slug: "window",
		},
	}

	mockService := new(MockAdminAllTeamsGetter)
	mockService.On("GetAllTeamsAdmin", mock.Anything).Return(result, nil)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Get("/api/admin/employees/teams", GetAllTeams(log, mockService))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/employees/teams", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "окна и двери")

	mockService.AssertExpectations(t)
}

func TestGetAllTeams_InternalError(t *testing.T) {
	resultError := errors.New("database error")

	mockService := new(MockAdminAllTeamsGetter)
	mockService.On("GetAllTeamsAdmin", mock.Anything).Return(nil, resultError)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	r := chi.NewRouter()
	r.Get("/api/admin/employees/teams", GetAllTeams(log, mockService))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/employees/teams", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockService.AssertExpectations(t)
}
