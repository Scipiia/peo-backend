package get

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage"

	"github.com/go-chi/render"
)

type AdminCoefGetter interface {
	GetAllCoefficientAdmin(ctx context.Context) ([]*storage.CoefficientPEOAdmin, error)
}

// GetCoefficientAdmin возвращает коэффициенты для ПЭО
// @Summary Получить коэффициенты
// @Tags admin
// @Produce json
// @Success 200 {array} storage.CoefficientPEOAdmin "Список коэффициентов"
// @Failure 500 {string} string "Internal server error"
// @Router /api/admin/coefficient [get]
func GetCoefficientAdmin(log *slog.Logger, coef AdminCoefGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.admin.GetCoefficientAdmin"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		coefficients, err := coef.GetAllCoefficientAdmin(ctx)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("ошибка получения всех коэффициентов для ПЭО")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, coefficients)
	}
}

type AllEmployeesAdminGetter interface {
	GetAllEmployeesAdmin(ctx context.Context) ([]*storage.EmployeesAdmin, error)
}

// GetAllEmployeesAdmin возвращает список сотрудников
// @Summary Получить сотрудников
// @Tags admin
// @Produce json
// @Success 200 {array} storage.EmployeesAdmin
// @Failure 500 {string} string "Internal server error"
// @Router /api/admin/employees [get]
func GetAllEmployeesAdmin(log *slog.Logger, emp AllEmployeesAdminGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.admin.GetAllEmployeesAdmin"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		employees, err := emp.GetAllEmployeesAdmin(ctx)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("ошибка получения всех сотрудников")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, employees)
	}
}

type AdminAllTeamsGetter interface {
	GetAllTeamsAdmin(ctx context.Context) ([]*storage.TeamAdmin, error)
}

// GetAllTeams возвращает список бригад
// @Summary Получить бригады
// @Tags admin
// @Produce json
// @Success 200 {array} storage.TeamAdmin
// @Failure 500 {string} string "Internal server error"
// @Router /api/admin/employees/teams [get]
func GetAllTeams(log *slog.Logger, emp AdminAllTeamsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.admin.GetAllTeams"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		teams, err := emp.GetAllTeamsAdmin(ctx)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("ошибка получения всех бригад")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, teams)
	}
}
