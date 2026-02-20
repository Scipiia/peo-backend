package get

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"vue-golang/internal/storage"
)

// pkg/http-server/handlers/template/get.go

type TemplateJSON interface {
	GetTemplateByCode(ctx context.Context, code string) (*storage.Template, error)
	GetAllTemplates(ctx context.Context) ([]*storage.Template, error)

	GetTemplateByCodeAdmin(ctx context.Context, code string) (*storage.Template, error)
	GetAllTemplatesAdmin(ctx context.Context) ([]*storage.Template, error)
}

type ResponseForm struct {
	ID         int                 `json:"ID"`
	Code       string              `json:"code"`
	Name       string              `json:"name"`
	Category   string              `json:"category"`
	Systema    *string             `json:"systema"`
	TypeIzd    *string             `json:"type_izd"`
	Profile    *string             `json:"profile"`
	Operations []storage.Operation `json:"operations"`
}

func GetTemplatesByCode(log *slog.Logger, template TemplateJSON) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.GetTemplatesByCode"

		//log.With(
		//	slog.String("op", op),
		//	slog.String("request_id", middleware.GetReqID(r.Context())),
		//).Info("Fetching template by code")

		code := r.URL.Query().Get("code")
		if code == "" {
			log.With(slog.String("op", op)).Error("Missing 'code' in query parameters")
			http.Error(w, "Missing required query parameter 'code'", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Получаем шаблон из хранилища
		template, err := template.GetTemplateByCode(ctx, code)
		if err != nil {
			if strings.Contains(err.Error(), "не найден") || errors.Is(err, sql.ErrNoRows) {
				log.With(slog.String("op", op), slog.String("code", code)).Warn("Form not found")
				http.Error(w, "Form not found", http.StatusNotFound)
				return
			}

			log.With(
				slog.String("op", op),
				slog.String("code", code),
				slog.String("error", err.Error()),
			).Error("Failed to fetch template")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Формируем ответ
		response := ResponseForm{
			ID:         template.ID,
			Code:       template.Code,
			Name:       template.Name,
			Category:   template.Category,
			Systema:    template.Systema,
			TypeIzd:    template.TypeIzd,
			Profile:    template.Profile,
			Operations: template.Operations,
		}

		//log.With(slog.String("code", code)).Info("Successfully fetched form")

		// Отправляем JSON
		render.JSON(w, r, response)
	}
}

type ResponseAllForm struct {
	Template []*storage.Template
	Error    string
}

func GetAllTemplates(log *slog.Logger, template TemplateJSON) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.GetAllTemplates"

		//log.With(slog.String("op", op)).Info("Fetching all templates")

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		templates, err := template.GetAllTemplates(ctx)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Failed to fetch templates")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response := ResponseAllForm{
			Template: templates,
			Error:    "",
		}

		render.JSON(w, r, response)
	}
}

func GetTemplatesByCodeAdmin(log *slog.Logger, template TemplateJSON) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.GetTemplatesByCode"

		//log.With(
		//	slog.String("op", op),
		//	slog.String("request_id", middleware.GetReqID(r.Context())),
		//).Info("Fetching template by code")

		code := r.URL.Query().Get("code")
		if code == "" {
			log.With(slog.String("op", op)).Error("Missing 'code' in query parameters")
			http.Error(w, "Missing required query parameter 'code'", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Получаем шаблон из хранилища
		template, err := template.GetTemplateByCodeAdmin(ctx, code)
		if err != nil {
			if strings.Contains(err.Error(), "не найден") || errors.Is(err, sql.ErrNoRows) {
				log.With(slog.String("op", op), slog.String("code", code)).Warn("Form not found")
				http.Error(w, "Form not found", http.StatusNotFound)
				return
			}

			log.With(
				slog.String("op", op),
				slog.String("code", code),
				slog.String("error", err.Error()),
			).Error("Failed to fetch template")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		//log.With(slog.String("code", code)).Info("Successfully fetched form")

		// Отправляем JSON
		render.JSON(w, r, template)
	}
}

type ResponseAllFormAdmin struct {
	Template []*storage.Template
	Error    string
}

func GetAllTemplatesAdmin(log *slog.Logger, template TemplateJSON) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.GetAllTemplates"

		//log.With(slog.String("op", op)).Info("Fetching all templates")

		ctx, cancel := context.WithTimeout(r.Context(), 50*time.Millisecond)
		defer cancel()

		templates, err := template.GetAllTemplatesAdmin(ctx)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Failed to fetch templates")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		response := ResponseAllForm{
			Template: templates,
			Error:    "",
		}

		render.JSON(w, r, response)
	}
}
