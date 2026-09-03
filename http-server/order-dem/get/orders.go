package get

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"vue-golang/internal/storage"

	"github.com/go-chi/render"
)

type OrdersGetter interface {
	GetOrdersMonth(ctx context.Context, year int, month int, search string) ([]*storage.Order, error)
}

func GetOrdersFilter(log *slog.Logger, getOrders OrdersGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.orders.orders.GetOrdersFilter"

		yearStr := r.URL.Query().Get("year")
		monthStr := r.URL.Query().Get("month")
		search := r.URL.Query().Get("search")

		var year, month int
		var err error

		if search == "" {
			if yearStr == "" || monthStr == "" {
				log.Error("Missing year or month in query parameters", slog.Bool("has_search", search != ""))
				http.Error(w, "Missing year or month", http.StatusBadRequest)
				return
			}

			year, err = strconv.Atoi(yearStr)
			if err != nil {
				log.Error("Invalid year", slog.String("error", err.Error()))
				http.Error(w, "Invalid year", http.StatusBadRequest)
				return
			}

			month, err = strconv.Atoi(monthStr)
			if err != nil {
				log.Error("Invalid month", slog.String("error", err.Error()))
				http.Error(w, "Invalid month", http.StatusBadRequest)
				return
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Передаём в storage
		orders, err := getOrders.GetOrdersMonth(ctx, year, month, search)
		if err != nil {
			log.Error("не удалось получить заказы из дем", slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, orders)
	}
}
