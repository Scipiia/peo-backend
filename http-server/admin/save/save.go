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
	CreateEmployeeAdmin(ctx context.Context, input storage.CreateEmployeeInput) error
}

func SaveEmployerAdmin(log *slog.Logger, emp EmployeesProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.template.SaveEmployerAdmin"

		if r.Method != http.MethodPost {
			http.Error(w, "Метод запрещен", http.StatusMethodNotAllowed)
			return
		}

		var employer storage.CreateEmployeeInput

		err := json.NewDecoder(r.Body).Decode(&employer)
		if err != nil {
			http.Error(w, "Неверный JSON", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = emp.CreateEmployeeAdmin(ctx, employer)
		if err != nil {
			log.Error("Ошибка добавления сотрудника", "error", err)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
