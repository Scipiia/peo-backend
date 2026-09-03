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

func GetOrderDetails(log *slog.Logger, order OrderDetailsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.get_orders.GetOrderDetails1"

		orderNum := chi.URLParam(r, "orderNum")

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		details, err := order.GetOrderDetails(ctx, orderNum)
		if err != nil {
			log.Error("не удалось получить детали заказа из дема", slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, details)
	}
}
