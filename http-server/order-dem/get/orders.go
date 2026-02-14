package get

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"vue-golang/internal/storage"
)

type ResponseOrders struct {
	Orders []*storage.Order `json:"orders"`
	Status string           `json:"status"`
	Error  string           `json:"error"`
}

type GetOrders interface {
	GetOrdersMonth(ctx context.Context, year int, month int, search string) ([]*storage.Order, error)
}

func GetOrdersFilter(log *slog.Logger, getOrders GetOrders) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.orders.orders.GetOrdersFilter"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Получаем параметры
		yearStr := r.URL.Query().Get("year")
		monthStr := r.URL.Query().Get("month")
		search := r.URL.Query().Get("search")

		var year, month int
		var err error

		// Если поиск не указан — year и month обязательны
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

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Передаём в storage
		orders, err := getOrders.GetOrdersMonth(ctx, year, month, search)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Ошибка при получении заказов из дема")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, ResponseOrders{Error: "В базе не найдено заказов"})
			return
		}

		render.JSON(w, r, ResponseOrders{
			Orders: orders,
			Status: strconv.Itoa(http.StatusOK),
		})
	}
}
