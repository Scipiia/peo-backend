package save

import (
	"context"
	"encoding/json"
	"fmt"
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
	SaveNashchelnikNorm(ctx context.Context, legacyID int64, orderNum string, a, b, c, d, sqr, count float64) (*storage.GetOrderDetails, error)
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

func SaveNashchelnikCalc(log *slog.Logger, storage ResultNorm) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.post.SaveNashchelnikCalc"

		// 1. Парсим входные данные
		var req struct {
			LegacyID int64   `json:"legacy_id"` // ID из старой базы (dem_orders.idorders)
			OrderNum string  `json:"order_num"` // Номер заказа (для связи)
			A        float64 `json:"a"`
			B        float64 `json:"b"`
			C        float64 `json:"c"`
			D        float64 `json:"d"`
			Count    float64 `json:"count"`
			Sqr      float64 `json:"sqr"`
		}

		fmt.Println(req)

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// 2. Вызываем функцию сохранения и расчета
		item, err := storage.SaveNashchelnikNorm(ctx, req.LegacyID, req.OrderNum, req.A, req.B, req.C, req.D, req.Sqr, req.Count)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Ошибка при сохранении нащельника")

			// Если ошибка "HAS_NASHCHELNIK" тут не нужна, так как мы уже в калькуляторе.
			// Но если какая-то другая ошибка БД:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// 3. Возвращаем сохраненный заказ с операциями
		// Фронтенд получит этот JSON и сможет сразу показать нормы или перейти к назначению
		render.JSON(w, r, item)
	}
}
