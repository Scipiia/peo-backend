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

type TemplateByCodeGetter interface {
	GetTemplateByCode(ctx context.Context, code string) (*storage.Template, error)
}

func GetTemplatesByCode(log *slog.Logger, template TemplateByCodeGetter) http.HandlerFunc {
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

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		template, err := template.GetTemplateByCode(ctx, code)
		if err != nil {
			log.With(slog.String("op", op), slog.String("code", code), slog.String("error", err.Error())).Error("Failed to fetch template")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		//log.With(slog.String("code", code)).Info("Successfully fetched form")

		render.JSON(w, r, template)
	}
}

type AllTemplatesGetter interface {
	GetAllTemplates(ctx context.Context) ([]*storage.Template, error)
}

func GetAllTemplates(log *slog.Logger, template AllTemplatesGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.GetAllTemplates"

		log.With(slog.String("op", op)).Info("Fetching all templates")

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		templates, err := template.GetAllTemplates(ctx)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Failed to fetch templates")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, templates)
	}
}

type TemplateByCodeAdminGetter interface {
	GetTemplateByCodeAdmin(ctx context.Context, id int64) (*storage.Template, error)
}

func GetTemplatesByCodeAdmin(log *slog.Logger, template TemplateByCodeAdminGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.GetTemplatesByCode"

		//log.With(
		//	slog.String("op", op),
		//	slog.String("request_id", middleware.GetReqID(r.Context())),
		//).Info("Fetching template by code")

		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.With(slog.String("op", op)).Error("Missing 'id' in query parameters")
			http.Error(w, "Missing required query parameter 'id'", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Получаем шаблон из хранилища
		template, err := template.GetTemplateByCodeAdmin(ctx, id)
		if err != nil {
			log.With(
				slog.String("op", op),
				slog.Int64("id", id),
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

type AllTemplatesAdminGetter interface {
	GetAllTemplatesAdmin(ctx context.Context) ([]*storage.Template, error)
}

func GetAllTemplatesAdmin(log *slog.Logger, template AllTemplatesAdminGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.GetAllTemplates"

		//log.With(slog.String("op", op)).Info("Fetching all templates")

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		templates, err := template.GetAllTemplatesAdmin(ctx)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Failed to fetch templates")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, templates)
	}
}
