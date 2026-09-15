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

// SaveNormOrderOperation сохранение нормированных операций к заказу
// @Summary Сохранить новые операции
// @Tags norm-orders
// @Accept json
// @Produce json
// @Param request body storage.OrderNormDetails true "Нормированные операции к наряду"
// @Success 200 {object} Response
// @Failure 400 {string} string "Неверные данные"
// @Failure 500 {string} string "не удалось сохранить нормировку / не удалось сохранить операции"
// @Router /api/orders/order-norm/operations [post]
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

		render.JSON(w, r, Response{OrderID: orderID})
	}
}

type SaveNashchelnikSaver interface {
	SaveNashchelnikNorm(ctx context.Context, legacyID int64, orderNum string, a, b, c, d, sqr, count float64, opsFromFront []storage.NormOperation) (*storage.GetOrderDetails, error)
}

type SaveNashchelnikCalcRequest struct {
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

// SaveNashchelnikCalc сохраняет расчёт нащельника
// @Summary Сохранить расчёт нащельника
// @Tags norm-orders
// @Accept json
// @Produce json
// @Param request body SaveNashchelnikCalcRequest true "Данные расчёта нащельника"
// @Success 200 {object} storage.GetOrderDetails
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/orders/nashchelnik/calc [post]
func SaveNashchelnikCalc(log *slog.Logger, res SaveNashchelnikSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.post.SaveNashchelnikCalc"

		var req SaveNashchelnikCalcRequest

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
