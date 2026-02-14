package get

import (
	"context"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage"
)

type AdminCoefProvider interface {
	GetAllCoefficientAdmin(ctx context.Context) ([]*storage.CoefficientPEOAdmin, error)
	GetAllEmployeesAdmin(ctx context.Context) ([]*storage.EmployeesAdmin, error)
}

func GetCoefficientAdmin(log *slog.Logger, coef AdminCoefProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.admin.GetCoefficientAdmin"

		ctx, cancel := context.WithTimeout(r.Context(), 50*time.Millisecond)
		defer cancel()

		coef, err := coef.GetAllCoefficientAdmin(ctx)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("ошибка получения всех коэффициентов для ПЭО")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, coef)

		w.WriteHeader(http.StatusOK)
	}
}

func GetAllEmployeesAdmin(log *slog.Logger, emp AdminCoefProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.admin.GetAllEmployeesAdmin"

		ctx, cancel := context.WithTimeout(r.Context(), 50*time.Millisecond)
		defer cancel()

		employees, err := emp.GetAllEmployeesAdmin(ctx)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("ошибка получения всех коэффициентов для ПЭО")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, employees)

		w.WriteHeader(http.StatusOK)
	}
}
