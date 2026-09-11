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

// GetTemplatesByCode возвращает шаблон операции
// @Summary Получить json шаблон операции
// @Tags template
// @Produce json
// @Param code query string true "Код шаблона"
// @Success 200 {object} storage.Template
// @Failure 400 {string} string "Missing required query parameter 'code'"
// @Failure 500 {string} string "Internal server error"
// @Router /api/template [get]
func GetTemplatesByCode(log *slog.Logger, template TemplateByCodeGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.GetTemplatesByCode"

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

		render.JSON(w, r, template)
	}
}

type AllTemplatesGetter interface {
	GetAllTemplates(ctx context.Context) ([]*storage.Template, error)
}

// GetAllTemplates возвращает все шаблоны операции
// @Summary Получить все json шаблоны операции
// @Tags template
// @Produce json
// @Success 200 {array} storage.Template
// @Failure 500 {string} string "Internal server error"
// @Router /api/all_templates [get]
func GetAllTemplates(log *slog.Logger, template AllTemplatesGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.GetAllTemplates"

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
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

// GetTemplatesByCodeAdmin возвращает шаблон операции для админки
// @Summary Получить json шаблон операции
// @Tags admin
// @Produce json
// @Param id query int true "ID шаблона"
// @Success 200 {object} storage.Template
// @Failure 400 {string} string "Missing or invalid query parameter 'id'"
// @Failure 500 {string} string "Internal server error"
// @Router /api/admin/template [get]
func GetTemplatesByCodeAdmin(log *slog.Logger, template TemplateByCodeAdminGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.GetTemplatesByCode"

		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.With(slog.String("op", op)).Error("Missing 'id' in query parameters")
			http.Error(w, "Missing required query parameter 'id'", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		template, err := template.GetTemplateByCodeAdmin(ctx, id)
		if err != nil {
			log.With(slog.String("op", op), slog.Int64("id", id), slog.String("error", err.Error())).Error("Failed to fetch template")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, template)
	}
}

type AllTemplatesAdminGetter interface {
	GetAllTemplatesAdmin(ctx context.Context) ([]*storage.Template, error)
}

// GetAllTemplates возвращает все шаблоны операции для админки
// @Summary Получить все json шаблоны операции
// @Tags admin
// @Produce json
// @Success 200 {array} storage.Template
// @Failure 500 {string} string "Internal server error"
// @Router /api/admin/all_templates [get]
func GetAllTemplatesAdmin(log *slog.Logger, template AllTemplatesAdminGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.GetAllTemplates"

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
