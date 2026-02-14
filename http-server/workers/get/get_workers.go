package get

import (
	"context"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage"
)

type Workers interface {
	GetAllWorkers(ctx context.Context) ([]storage.GetWorkers, error)
}

func GetWorkers(log *slog.Logger, worker Workers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.order-dem-norm.get.GetWorkers"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		workers, err := worker.GetAllWorkers(ctx)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Ошибка при получении работяг")
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		//log.With(slog.Int("найдено работников", len(workers))).Info("работники найдены")

		render.JSON(w, r, workers)
	}
}
