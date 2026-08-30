package ready

import (
	"context"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
)

type DBChecker interface {
	Ping(ctx context.Context) error
}

func Ready(log *slog.Logger, db DBChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			log.Error("database unavailable",
				slog.String("error", err.Error()),
			)

			w.WriteHeader(http.StatusServiceUnavailable)

			render.JSON(w, r, map[string]string{
				"status":   "error",
				"database": "503 Service Unavailable",
			})
			return
		}
		render.JSON(w, r, map[string]string{
			"status":   "ok",
			"database": "ok",
		})
	}
}
