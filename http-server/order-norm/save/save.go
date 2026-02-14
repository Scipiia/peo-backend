package save

import (
	"context"
	"encoding/json"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"vue-golang/internal/storage"
)

type ResultNorm interface {
	SaveNormOrder(ctx context.Context, result storage.OrderNormDetails) (int64, error)
	SaveNormOperation(ctx context.Context, OrderID int64, operations []storage.NormOperation) error
}

type Response struct {
	OrderID int64  `json:"order_id"`
	Status  string `json:"status"`
	Error   string `json:"error"`
}

func SaveNormOrderOperation(log *slog.Logger, res ResultNorm) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.save.SaveNormOrderOperation"

		//var req RequestNormData
		var req storage.OrderNormDetails
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("Неверный JSON", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Неверные данные", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		orderID, err := res.SaveNormOrder(ctx, req)
		if err != nil {
			log.Error("Ошибка при сохранения нормированного наряда", slog.String("op", op), slog.String("error", err.Error()))
			render.JSON(w, r, Response{Error: "не удалось сохранить нормировку"})
			return
		}

		// Сохраняем операции
		err = res.SaveNormOperation(ctx, orderID, req.Operations)
		if err != nil {
			log.Error("Ошибка при сохранении операции нормированного наряда", slog.String("op", op), slog.String("error", err.Error()))
			render.JSON(w, r, Response{Error: "не удалось сохранить нормировку"})
			return
		}

		//log.Info("message added", slog.Int64("id", orderID))

		render.JSON(w, r, Response{
			OrderID: orderID,
			Status:  strconv.Itoa(http.StatusOK),
			Error:   "",
		})
	}
}
