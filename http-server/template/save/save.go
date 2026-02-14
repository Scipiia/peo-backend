package save

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"vue-golang/internal/storage"
)

type TemplateCreateProvider interface {
	CreateTemplateAdmin(ctx context.Context, res storage.TemplateAdmin) error
}

func SaveTemplateAdmin(log *slog.Logger, temp TemplateCreateProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.SaveTemplateAdmin"

		var req struct {
			Code       string              `json:"code"`
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

		opsJSON, err := json.Marshal(req.Operations)
		if err != nil {
			log.Error(fmt.Sprintf("%s: ошибка сериализации operations: %v", op, err))
			http.Error(w, "ошибка обработки операций", http.StatusInternalServerError)
			return
		}

		// Если Rules == nil, заменяем на пустой срез
		if req.Rules == nil {
			req.Rules = []storage.Rule{}
		}

		rulesJSON, err := json.Marshal(req.Rules)
		if err != nil {
			log.Error(fmt.Sprintf("%s: ошибка сериализации правил: %v", op, err))
			http.Error(w, "ошибка обработки правил шаблона", http.StatusInternalServerError)
			return
		}

		log.Info("FFFFF", rulesJSON)

		err = temp.CreateTemplateAdmin(r.Context(), storage.TemplateAdmin{
			Code:      req.Code,
			Category:  req.Category,
			IsActive:  req.IsActive,
			Name:      req.Name,
			Profile:   req.Profile,
			Systema:   req.Systema,
			TypeIzd:   req.TypeIzd,
			Operation: string(opsJSON),
			Rules:     string(rulesJSON),
			HeadName:  req.HeadName,
		})
		if err != nil {
			log.Error(fmt.Sprintf("%s: %v", op, err))
			http.Error(w, "ошибка создания шаблона", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "created"})
	}
}
