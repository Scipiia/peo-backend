package generate_excel

import (
	"fmt"
	"golang.org/x/net/context"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage/mysql"
)

type GenerateExcelHandler interface {
	GenerateExcel(ctx context.Context, filter mysql.ProductFilter) ([]byte, error)
}

func GenerateReportExcel(log *slog.Logger, gen GenerateExcelHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.norm.GenerateReportExcel"

		// Парсинг параметров (как у тебя)
		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")
		orderNum := r.URL.Query().Get("order_num")
		typeIzd := r.URL.Query()["type"]

		// Логика с датами (оставляем твою, она хорошая)
		now := time.Now()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

		// Функция парсинга внутри (ок, но можно вынести)
		fDate, err := time.Parse("2006-01-02", fromStr)
		if err != nil && fromStr != "" {
			http.Error(w, "invalid from date", http.StatusBadRequest)
			return
		}
		if fromStr == "" {
			fDate = startOfMonth
		}

		tDate, err := time.Parse("2006-01-02", toStr)
		if err != nil && toStr != "" {
			http.Error(w, "invalid to date", http.StatusBadRequest)
			return
		}
		if toStr == "" {
			tDate = now
		}

		filter := mysql.ProductFilter{
			From:     fDate,
			To:       tDate,
			OrderNum: orderNum,
			Type:     typeIzd,
		}

		//fmt.Printf("TYPE handle %s", filter.Type)

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second) // На Excel можно побольше времени
		defer cancel()

		excelBytes, err := gen.GenerateExcel(ctx, filter)
		if err != nil {
			log.Error("failed to generate excel", "op", op, "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// ФОРМИРУЕМ ОТВЕТ
		fileName := fmt.Sprintf("PEO_Report_%s.xlsx", time.Now().Format("2006-01-02_150405"))

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		w.Write(excelBytes)
	}
}
