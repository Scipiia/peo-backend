package get

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage"

	"github.com/go-chi/render"
)

type WorkersGetter interface {
	GetAllWorkers(ctx context.Context, typeIzd string) ([]storage.GetWorkers, error)
}

// GetWorkers возвращает всех сотрудников цеха
// @Summary Получить всех сотрудников
// @Tags workers
// @Produce json
// @Param type query string true "Бригада сотрудников"
// @Success 200 {array} storage.GetWorkers
// @Failure 500 {string} string "Internal server error"
// @Router /api/workers/all [get]
func GetWorkers(log *slog.Logger, worker WorkersGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.order-dem-norm.get.GetWorkers"

		typeIzd := r.URL.Query().Get("type")

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		workers, err := worker.GetAllWorkers(ctx, typeIzd)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Ошибка при получении работников")
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, workers)
	}
}
