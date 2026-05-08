package update

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"vue-golang/internal/storage"
)

type UpdateCoefProvider interface {
	UpdateCoefficientPEOAdmin(ctx context.Context, coeffs []storage.CoefficientPEOAdmin) error
	UpdateAllEmployeesAdmin(ctx context.Context, id int64, input storage.UpdateEmployeeInput) error
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

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
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

		idStr := chi.URLParam(r, "id") // или r.PathValue("id") для Go 1.22+
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "неверный ID", http.StatusBadRequest)
			return
		}

		var employees storage.UpdateEmployeeInput

		if err := json.NewDecoder(r.Body).Decode(&employees); err != nil {
			http.Error(w, "Неверный JSON", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = update.UpdateAllEmployeesAdmin(ctx, id, employees)
		if err != nil {
			log.Error("Ошибка обновления всех работников", "error", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
