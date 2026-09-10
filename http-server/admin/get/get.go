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

func GetCoefficientAdmin(log *slog.Logger, coef AdminCoefGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.admin.GetCoefficientAdmin"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		coef, err := coef.GetAllCoefficientAdmin(ctx)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("ошибка получения всех коэффициентов для ПЭО")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, coef)
	}
}

type AllEmployeesAdminGetter interface {
	GetAllEmployeesAdmin(ctx context.Context) ([]*storage.EmployeesAdmin, error)
}

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

		w.WriteHeader(http.StatusOK)
	}
}

type AdminAllTeamsGetter interface {
	GetAllTeamsAdmin(ctx context.Context) ([]*storage.TeamAdmin, error)
}

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
