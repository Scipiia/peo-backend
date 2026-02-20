package recalculate_norm

import (
	"context"
	"encoding/json"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/service"
	"vue-golang/internal/storage"
)

type NormCalculator interface {
	CalculateNorm(ctx context.Context, orderNum string, pos int, typeIzd string, templateCode string, itemCount int) ([]storage.Operation, service.Context, error)
}

type Resp struct {
	Operation []storage.Operation `json:"operation"`
	Context   service.Context     `json:"context"`
}

func CalculateNormOperations(log *slog.Logger, calc NormCalculator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.norm.CalculateNormOperations"

		var req struct {
			OrderNum     string `json:"order_num"`
			Position     int    `json:"position"`
			TypeIzd      string `json:"type"`
			TemplateCode string `json:"template"`
			ItemCount    int    `json:"count"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Некорректный JSON", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		norm, ctxData, err := calc.CalculateNorm(ctx, req.OrderNum, req.Position, req.TypeIzd, req.TemplateCode, req.ItemCount)
		if err != nil {
			log.Error("Failed to calculate norm", slog.String("error", err.Error()))
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, Resp{
			Operation: norm,
			Context:   ctxData,
		})
	}
}
