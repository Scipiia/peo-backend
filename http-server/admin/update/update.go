package update

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"vue-golang/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type AdminCoefUpdater interface {
	UpdateCoefficientPEOAdmin(ctx context.Context, coeffs []storage.CoefficientPEOAdmin) error
}

// UpdateCoefficientAdmin обновление соэффициентов
// @Summary Обновить соэффициент
// @Tags admin
// @Accept json
// @Produce json
// @Param request body []storage.CoefficientPEOAdmin true "Список коэффициентов"
// @Success 204
// @Failure 400 {string} string "ошибка парсинга JSON"
// @Failure 500 {string} string "Internal server error"
// @Router /api/admin/coefficient/update [put]
func UpdateCoefficientAdmin(log *slog.Logger, update AdminCoefUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.UpdateCoefficientAdmin"

		var coeffs []storage.CoefficientPEOAdmin
		if err := json.NewDecoder(r.Body).Decode(&coeffs); err != nil {
			http.Error(w, "Неверный JSON", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err := update.UpdateCoefficientPEOAdmin(ctx, coeffs)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("ошибка обновления коэффициентов")
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		render.NoContent(w, r)
	}
}

type AdminEmployeesUpdater interface {
	UpdateAllEmployeesAdmin(ctx context.Context, id int64, input storage.UpdateEmployeeInput) error
}

// @Summary Обновить сотрудника
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID сотрудника"
// @Param request body storage.UpdateEmployeeInput true "Данные сотрудника"
// @Success 204 "No Content"
// @Failure 400 {string} string "Неверный JSON или ID"
// @Failure 500 {string} string "Internal server error"
// @Router /api/admin/employees/update/{id} [put]
func UpdateEmployeesAdmin(log *slog.Logger, update AdminEmployeesUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.UpdateEmployeesAdmin"

		idStr := chi.URLParam(r, "id")
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
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("ошибка обновления всех сотрудников")
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		render.NoContent(w, r)
	}
}
