package post

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type MosquitoHandler interface {
	Sync(ctx context.Context, since time.Time) error
}

func SyncButton(log *slog.Logger, ms MosquitoHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.norm.SyncButton"

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		//specificTime := time.Date(2026, 4, 7, 14, 0, 0, 0, time.UTC)
		since := time.Now().AddDate(0, -6, 0)

		err := ms.Sync(ctx, since)
		if err != nil {
			log.Error("failed to sync ", "op", op, "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("sync completed"))
	}
}
