package generate_excel

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"vue-golang/internal/storage/mysql"

	"context"
)

type GenerateExcelHandler interface {
	GenerateExcel(ctx context.Context, filter mysql.ProductFilter) ([]byte, error)
}

// GenerateReportExcel генерирует Excel-отчёт по изделиям за период
// @Summary Скачать отчёт в формате Excel
// @Tags report-excel
// @Produce application/octet-stream
// @Param from query string false "Дата начала периода (YYYY-MM-DD), по умолчанию — начало текущего месяца"
// @Param to query string false "Дата окончания периода (YYYY-MM-DD), по умолчанию — текущая дата"
// @Param order_num query string false "Номер заказа"
// @Param type query []string false "Типы изделий (можно указать несколько)" collectionFormat(multi)
// @Success 200 {file} binary "Excel-файл с отчётом"
// @Failure 400 {string} string "invalid from date / invalid to date"
// @Failure 500 {string} string "Internal error"
// @Router /api/report/excel [get]
func GenerateReportExcel(log *slog.Logger, gen GenerateExcelHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.norm.GenerateReportExcel"

		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")
		orderNum := r.URL.Query().Get("order_num")
		typeIzd := r.URL.Query()["type"]

		now := time.Now()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

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

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		excelBytes, err := gen.GenerateExcel(ctx, filter)
		if err != nil {
			log.Error("failed to generate excel", "op", op, "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		fileName := fmt.Sprintf("%s.xlsx", time.Now().Format("2006-01-02_150405"))

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		_, err = w.Write(excelBytes)
		if err != nil {
			log.Error("failed to generate excel", "op", op, "err", err)
			return
		}
	}
}
