package save

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage"
)

type ResultWorkers interface {
	SaveOperationWorkers(ctx context.Context, req storage.SaveWorkers) error
}

func SaveWorkersOperation(log *slog.Logger, result ResultWorkers) http.HandlerFunc {
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

		log.Info("Received assignments",
			slog.Int("total", len(req.Assignments)),
			slog.Any("sample", req.Assignments[0]),
		)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err := result.SaveOperationWorkers(ctx, req)
		if err != nil {
			log.Error("Ошибка сохранения назначении сотрудников", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Info("Assignments saved successfully",
			slog.Int("saved_count", len(req.Assignments)),
		)

		render.JSON(w, r, map[string]interface{}{
			"status":  "success",
			"saved":   len(req.Assignments),
			"details": req.Assignments,
		})
	}
}
