package save

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage"

	"github.com/go-chi/render"
)

type ResultWorkersGetter interface {
	SaveOperationWorkers(ctx context.Context, req storage.SaveWorkers) error
}

type SaveWorkersResponse struct {
	Status  string                     `json:"status"`
	Saved   int                        `json:"saved"`
	Details []storage.OperationWorkers `json:"details"` // подставьте реальный тип элемента req.Assignments
}

// SaveWorkersOperation назначение и сохранение сотрудников на операции
// @Summary Сохранить сотрудников
// @Tags workers
// @Accept json
// @Produce json
// @Param request body storage.SaveWorkers true "Данные об операциях и сотрудниках"
// @Success 200 {object} SaveWorkersResponse
// @Failure 400 {string} string "Bad request: invalid JSON / No assignments provided / Assignment N: <field> is required"
// @Failure 500 {string} string "Internal server error"
// @Router /api/workers [post]
func SaveWorkersOperation(log *slog.Logger, result ResultWorkersGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.executor.SaveWorkersOperation"

		var req storage.SaveWorkers
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("Invalid JSON", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		if len(req.Assignments) == 0 {
			log.Warn("Пустой лист назначения сотрудников на операции", slog.String("op", op))
			http.Error(w, "No assignments provided", http.StatusBadRequest)
			return
		}

		for i, a := range req.Assignments {
			if a.ProductID == 0 {
				log.Error("Ошибка id в dem_product_instance_al", slog.Int("index", i), slog.Any("assignment", a))
				http.Error(w, fmt.Sprintf("Assignment %d: product_id is required", i), http.StatusBadRequest)
				return
			}
			if a.EmployeeID == 0 {
				log.Error("Ошибка назначения сотрудника id", slog.Int("index", i), slog.Any("assignment", a))
				http.Error(w, fmt.Sprintf("Assignment %d: employee_id is required", i), http.StatusBadRequest)
				return
			}
			if a.OperationName == "" {
				log.Error("Ошибка операции", slog.Int("index", i), slog.Any("assignment", a))
				http.Error(w, fmt.Sprintf("Assignment %d: operation_name is required", i), http.StatusBadRequest)
				return
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()

		err := result.SaveOperationWorkers(ctx, req)
		if err != nil {
			log.Error("Ошибка сохранения назначении сотрудников", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, SaveWorkersResponse{
			Status:  "success",
			Saved:   len(req.Assignments),
			Details: req.Assignments,
		})
	}
}
