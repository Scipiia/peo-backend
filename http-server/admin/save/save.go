package save

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage"

	"github.com/go-chi/render"
)

type EmployeesProvider interface {
	CreateEmployeeAdmin(ctx context.Context, input storage.CreateEmployeeInput) error
}

// SaveTemplateAdmin создание и сохранение нового шаблона операции
// @Summary Сохранить новый шаблон
// @Tags admin
// @Accept json
// @Produce json
// @Param request body storage.CreateEmployeeInput true "Данные нового сотрудника"
// @Success 204
// @Failure 400 {string} string "ошибка парсинга JSON"
// @Failure 500 {string} string "Internal server error"
// @Router /api/admin/template/new [post]
func SaveEmployerAdmin(log *slog.Logger, emp EmployeesProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.SaveEmployerAdmin"

		var employer storage.CreateEmployeeInput

		err := json.NewDecoder(r.Body).Decode(&employer)
		if err != nil {
			http.Error(w, "Неверный JSON", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = emp.CreateEmployeeAdmin(ctx, employer)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("ошибка добавления сотрудников")
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		render.NoContent(w, r)
	}
}
