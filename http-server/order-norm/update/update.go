package update

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"vue-golang/internal/storage"
)

type ResultUpdateNorm interface {
	UpdateNormOrder(ctx context.Context, ID int64, update storage.UpdateOrderDetails) error
	UpdateFinalOrder(ctx context.Context, ID int64, update storage.UpdateFinalOrderDetails) error
	UpdateStatus(ctx context.Context, rootProductID int64, status string) error
}

func UpdateNormOrderOperation(log *slog.Logger, update ResultUpdateNorm) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.norm.UpdateNormHandler"

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var req storage.UpdateOrderDetails
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("Invalid JSON", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Invalid data", http.StatusBadRequest)
			return
		}

		log.Info("Обновление нормировки", slog.Int64("id", id))

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = update.UpdateNormOrder(ctx, id, req)
		if err != nil {
			log.Error("Ошибка обновления", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
			return
		}

		log.Info("Нормировка обновлена", slog.Int64("id", id))

		render.JSON(w, r, map[string]interface{}{
			"status":  strconv.Itoa(http.StatusOK),
			"norm_id": id,
		})
	}
}

func UpdateFinalOrder(log *slog.Logger, update ResultUpdateNorm) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.norm.UpdateFinalOrder"

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var req storage.UpdateFinalOrderDetails
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("Invalid JSON", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Invalid data", http.StatusBadRequest)
			return
		}

		log.Info("Обновление финальной нормировки", slog.Int64("id", id))

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = update.UpdateFinalOrder(ctx, id, req)
		if err != nil {
			log.Error("Ошибка обновления", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"status": "success",
		})
	}
}

func UpdateCancelStatus(log *slog.Logger, update ResultUpdateNorm) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.norm.UpdateCancelStatus"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if r.Method != http.MethodPost {
			log.Warn("Invalid method", "method", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			RootProductID int64 `json:"root_product_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("Failed to decode request body", "error", err)
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		err := update.UpdateStatus(ctx, req.RootProductID, "cancel")
		if err != nil {
			log.Error("Failed to update status to 'cancelled'", "error", err, "root_product_id", req.RootProductID)
			http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
			return
		}

		log.Info("Order successfully cancelled", "root_product_id", req.RootProductID)
	}
}
