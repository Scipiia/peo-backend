package update

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
	"vue-golang/internal/storage"
)

type TemplateUpdateProvider interface {
	UpdateTemplateAdmin(ctx context.Context, code int, update storage.TemplateAdmin) error
}

func UpdateTemplateAdmin(log *slog.Logger, temp TemplateUpdateProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.UpdateTemplateAdmin"

		code := chi.URLParam(r, "code")
		templCode, err := strconv.Atoi(code)
		if err != nil {
			http.Error(w, "неверный ID шаблона", http.StatusBadRequest)
			return
		}

		// Распарсить тело запроса
		var req struct {
			Category   string              `json:"category"`
			IsActive   bool                `json:"is_active"`
			Name       string              `json:"name"`
			Profile    string              `json:"profile"`
			Systema    string              `json:"systema"`
			TypeIzd    string              `json:"type_izd"`
			Operations []storage.Operation `json:"operations"`
			Rules      []storage.Rule      `json:"rules"`
			HeadName   string              `json:"head_name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "ошибка парсинга JSON", http.StatusBadRequest)
			return
		}

		// 3. Сериализовать operations в JSON
		opsJSON, err := json.Marshal(req.Operations)
		if err != nil {
			log.Error(fmt.Sprintf("%s: ошибка сериализации operations: %v", op, err))
			http.Error(w, "ошибка обработки операций", http.StatusInternalServerError)
			return
		}

		err = temp.UpdateTemplateAdmin(r.Context(), templCode, storage.TemplateAdmin{
			Category:  req.Category,
			IsActive:  req.IsActive,
			Name:      req.Name,
			Profile:   req.Profile,
			Systema:   req.Systema,
			TypeIzd:   req.TypeIzd,
			Operation: string(opsJSON),
			HeadName:  req.HeadName,
		})
		if err != nil {
			log.Error(fmt.Sprintf("% s: %v", op, err))
			http.Error(w, "ошибка обновления шаблона", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
