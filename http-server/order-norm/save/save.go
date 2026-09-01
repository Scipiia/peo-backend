package save

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage"

	"github.com/go-chi/render"
)

type ResultNormSaver interface {
	SaveNormOrder(ctx context.Context, result storage.OrderNormDetails) (int64, error)
	SaveNormOperation(ctx context.Context, OrderID int64, operations []storage.NormOperation) error
}

type Response struct {
	OrderID int64 `json:"order_id"`
}

func SaveNormOrderOperation(log *slog.Logger, res ResultNormSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.save.SaveNormOrderOperation"

		var req storage.OrderNormDetails
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("Неверный JSON", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Неверные данные", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		orderID, err := res.SaveNormOrder(ctx, req)
		if err != nil {
			log.Error("Ошибка при сохранения нормированного наряда", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "не удалось сохранить нормировку", http.StatusInternalServerError)
			return
		}

		err = res.SaveNormOperation(ctx, orderID, req.Operations)
		if err != nil {
			log.Error("Ошибка при сохранении операции нормированного наряда", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "не удалось сохранить операции", http.StatusInternalServerError)
			return
		}

		//log.Info("message added", slog.Int64("id", orderID))

		render.JSON(w, r, Response{OrderID: orderID})
	}
}

type SaveNashchelnikSaver interface {
	SaveNashchelnikNorm(ctx context.Context, legacyID int64, orderNum string, a, b, c, d, sqr, count float64, opsFromFront []storage.NormOperation) (*storage.GetOrderDetails, error)
}

func SaveNashchelnikCalc(log *slog.Logger, res SaveNashchelnikSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.post.SaveNashchelnikCalc"

		var req struct {
			LegacyID   int64                   `json:"legacy_id"`
			OrderNum   string                  `json:"order_num"`
			A          float64                 `json:"a"`
			B          float64                 `json:"b"`
			C          float64                 `json:"c"`
			D          float64                 `json:"d"`
			Count      float64                 `json:"count"`
			Sqr        float64                 `json:"sqr"`
			Operations []storage.NormOperation `json:"operations"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		item, err := res.SaveNashchelnikNorm(ctx, req.LegacyID, req.OrderNum, req.A, req.B, req.C, req.D, req.Sqr, req.Count, req.Operations)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Ошибка при сохранении нащельника")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, item)
	}
}
