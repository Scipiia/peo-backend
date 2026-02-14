package update

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage"
)

type UpdateCoefProvider interface {
	UpdateCoefficientPEOAdmin(ctx context.Context, coeffs []storage.CoefficientPEOAdmin) error
	UpdateAllEmployeesAdmin(ctx context.Context, emps []storage.EmployeesAdmin) error
}

func UpdateCoefficientAdmin(log *slog.Logger, update UpdateCoefProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.UpdateCoefficientAdmin"

		if r.Method != http.MethodPut {
			http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
			return
		}

		var coeffs []storage.CoefficientPEOAdmin
		if err := json.NewDecoder(r.Body).Decode(&coeffs); err != nil {
			http.Error(w, "Неверный JSON", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 50*time.Millisecond)
		defer cancel()

		err := update.UpdateCoefficientPEOAdmin(ctx, coeffs)
		if err != nil {
			log.Error("Ошибка обновления коэффициентов", "error", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func UpdateEmployeesAdmin(log *slog.Logger, update UpdateCoefProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.UpdateEmployeesAdmin"
		if r.Method != http.MethodPut {
			http.Error(w, "Метод не разрешён", http.StatusMethodNotAllowed)
			return
		}

		var employees []storage.EmployeesAdmin

		if err := json.NewDecoder(r.Body).Decode(&employees); err != nil {
			http.Error(w, "Неверный JSON", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 50*time.Millisecond)
		defer cancel()

		err := update.UpdateAllEmployeesAdmin(ctx, employees)
		if err != nil {
			log.Error("Ошибка обновления коэффициентов", "error", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
