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

type NormOrderUpdater interface {
	UpdateNormOrder(ctx context.Context, ID int64, update storage.UpdateOrderDetails) error
}

// UpdateNormOrderOperation обновляет нормы операции
// @Summary Обновить нормы по операциям
// @Tags norm-orders
// @Accept json
// @Produce json
// @Param id path int true "ID заказа для обновления"
// @Param request body storage.UpdateOrderDetails true "Обновлённые данные по операциям"
// @Success 200 {integer} int64 "ID обновлённого заказа"
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/orders/order/norm/update/{id} [put]
func UpdateNormOrderOperation(log *slog.Logger, update NormOrderUpdater) http.HandlerFunc {
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

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		err = update.UpdateNormOrder(ctx, id, req)
		if err != nil {
			log.Error("Ошибка обновления", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, id)
	}
}

type FinalOrderUpdater interface {
	UpdateFinalOrder(ctx context.Context, ID int64, update storage.UpdateFinalOrderDetails) error
}

// UpdateFinalOrder обновляет детали по заказу
// @Summary Обновить детали заказа
// @Tags norm-orders
// @Accept json
// @Produce json
// @Param id path int true "ID заказа для обновления"
// @Param request body storage.UpdateFinalOrderDetails true "Обновлённые данные по заказу"
// @Success 204 "Успешно обновлено"
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Ошибка обновления"
// @Router /api/final/update/{id} [put]
func UpdateFinalOrder(log *slog.Logger, update FinalOrderUpdater) http.HandlerFunc {
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

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = update.UpdateFinalOrder(ctx, id, req)
		if err != nil {
			log.Error("Ошибка обновления", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
			return
		}

		render.NoContent(w, r)
	}
}

type CancelStatusUpdater interface {
	UpdateStatus(ctx context.Context, rootProductID int64, status string) error
}

type CancelStatusRequest struct {
	RootProductID int64 `json:"root_product_id"`
}

// UpdateCancelStatus отменяет статус заказа по корневому изделию
// @Summary Отменить статус заказа
// @Tags norm-orders
// @Accept json
// @Produce json
// @Param request body CancelStatusRequest true "ID корневого изделия"
// @Success 204 "Успешно отменено"
// @Failure 400 {string} string "Invalid request payload"
// @Failure 500 {string} string "Failed to cancel order"
// @Router /api/orders/cancel [post]
func UpdateCancelStatus(log *slog.Logger, update CancelStatusUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.norm.UpdateCancelStatus"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var req CancelStatusRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("ошибка декодирования JSON")
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		err := update.UpdateStatus(ctx, req.RootProductID, "cancel")
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Failed to update status to 'cancelled'")
			http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
			return
		}

		render.NoContent(w, r)
	}
}
