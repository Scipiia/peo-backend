package get

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type OrderDetailsGetter interface {
	GetOrderDetails(ctx context.Context, orderNum string) ([]*storage.ResultOrderDetails, error)
}

// GetOrderDetails возвращает заказ с изделиями по номеру
// @Summary Получить детали заказа по номеру
// @Tags orders
// @Produce json
// @Param orderNum path string true "Номер заказа"
// @Success 200 {array} storage.ResultOrderDetails
// @Failure 500 {string} string "Internal server error"
// @Router /api/orders/{orderNum} [get]
func GetOrderDetails(log *slog.Logger, order OrderDetailsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.get_orders.GetOrderDetails"

		orderNum := chi.URLParam(r, "orderNum")

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		details, err := order.GetOrderDetails(ctx, orderNum)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("ошибка получения заказов из дем")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, details)
	}
}
