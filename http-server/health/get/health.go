package get

import (
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

func Health(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.health.Health"

		log.Info("health check", slog.String("op", op))

		render.JSON(w, r, map[string]int{
			"status": http.StatusOK,
		})
	}
}
