package save

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage"
)

type EmployeesProvider interface {
	CreateEmployerAdmin(ctx context.Context, emp storage.EmployeesAdmin) error
}

func SaveEmployerAdmin(log *slog.Logger, emp EmployeesProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.SaveEmployerAdmin"

		if r.Method != http.MethodPost {
			http.Error(w, "Метод запрещен", http.StatusMethodNotAllowed)
			return
		}

		var employer storage.EmployeesAdmin
		err := json.NewDecoder(r.Body).Decode(&employer)
		if err != nil {
			http.Error(w, "Неверный JSON", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 50*time.Millisecond)
		defer cancel()

		err = emp.CreateEmployerAdmin(ctx, employer)
		if err != nil {
			log.Error("Ошибка добавления сотрудника", "error", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
