package get

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"vue-golang/internal/storage"
	"vue-golang/internal/storage/mysql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type OrderGetter interface {
	GetNormOrder(ctx context.Context, id int64) (*storage.GetOrderDetails, error)
}

// GetNormOrder возвращает нормированный заказ по ID
// @Summary Получить нормированный заказ
// @Tags norm-orders
// @Produce json
// @Param id path int true "ID нормированного заказа"
// @Success 200 {object} storage.GetOrderDetails
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Нормировка не найдена"
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router /api/orders/order/norm/{id} [get]
func GetNormOrder(log *slog.Logger, result OrderGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get.GetNormOrder"

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		norm, err := result.GetNormOrder(ctx, id)
		if err != nil {
			if strings.Contains(err.Error(), "не найдена") {
				log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Ошибка при получении нормированного заказа с операциями")
				http.Error(w, "Нормировка не найдена", http.StatusNotFound)
				return
			}
			log.Error("Ошибка получения нормировки", slog.String("op", op), slog.String("error", err.Error()))
			http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
			return
		}

		// Успешный ответ
		render.JSON(w, r, norm)
	}
}

type OrderByOrderNumGetter interface {
	GetNormOrdersByOrderNum(ctx context.Context, orderNum string, position int) ([]*storage.GetOrderDetails, error)
}

// GetNormOrdersOrderNum возвращает нормированные заказы по номеру заказа
// @Summary Получить нормированные заказы по номеру
// @Tags norm-orders
// @Produce json
// @Param order_num query string true "номер нормированного заказа"
// @Param position query int true "позиция нормированного заказа"
// @Success 200 {array} storage.GetOrderDetails
// @Failure 400 {string} string "Invalid position"
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router /api/orders/order-norm/by-order [get]
func GetNormOrdersOrderNum(log *slog.Logger, result OrderByOrderNumGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.order-norm.get.GetNormOrders"

		orderNum := r.URL.Query().Get("order_num")
		position := r.URL.Query().Get("position")

		positionInt, err := strconv.Atoi(position)
		if err != nil {
			http.Error(w, "Invalid position", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		orders, err := result.GetNormOrdersByOrderNum(ctx, orderNum, positionInt)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Ошибка при получении нормировок по номеру заказа")
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, orders)
	}
}

type OrdersGetter interface {
	GetNormOrders(ctx context.Context, orderNum, orderType string) ([]storage.GetOrderDetails, error)
}

// GetNormOrdersOrderNum возвращает все нормированные заказы
// @Summary Получить нормированные заказы
// @Tags norm-orders
// @Produce json
// @Param order_num query string false "номер заказа"
// @Param type query string false "тип заказа"
// @Success 200 {array} storage.GetOrderDetails
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router /api/orders/order/norm/all [get]
func GetNormOrders(log *slog.Logger, result OrdersGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get.GetNormOrders"

		orderNum := r.URL.Query().Get("order_num")
		orderType := r.URL.Query().Get("type")

		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()

		items, err := result.GetNormOrders(ctx, orderNum, orderType)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Ошибка при получении заказов")
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, items)
	}
}

type OrderDoubleGetter interface {
	GetNormOrderIdSub(ctx context.Context, id int64) ([]*storage.GetOrderDetails, error)
	GetMosquitoOrderDetails(ctx context.Context, requestedID int64) (*storage.GetOrderDetails, error)
	GetGutterOrderDetails(ctx context.Context, requestedID int64) (*storage.GetOrderDetails, error)
}

// DoubleReportOrder возвращает детали заказа с учётом источника данных (обычный заказ, москитные сетки, водоотлив)
// @Summary Получить детали заказа с учётом источника
// @Tags norm-orders
// @Produce json
// @Param id path int true "ID заказа"
// @Param source query string false "Источник данных: mosquito, vodootliv (по умолчанию — обычный заказ)"
// @Success 200 {array} storage.GetOrderDetails
// @Failure 400 {string} string "Invalid ID"
// @Failure 409 {string} string "REQUIRES_CALCULATOR"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/orders/order-norm/{id}/details [get]
func DoubleReportOrder(log *slog.Logger, result OrderDoubleGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get.DoubleReportOrder"

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		source := r.URL.Query().Get("source")

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		var sub []*storage.GetOrderDetails

		switch source {
		case "mosquito":
			log.Debug("loading mosquito order details", "id", id)
			item, err := result.GetMosquitoOrderDetails(ctx, id)
			if err != nil {
				log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Ошибка при получении заказов по номеру заказа")
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
			render.JSON(w, r, []*storage.GetOrderDetails{item})
			return
		case "vodootliv":
			log.Debug("loading vodootliv order details", "id", id)
			item, err := result.GetGutterOrderDetails(ctx, id)
			if err != nil {
				if strings.Contains(err.Error(), "REQUIRES_CALCULATOR") {
					// Отдаем 400 или 409 с понятным сообщением
					http.Error(w, "REQUIRES_CALCULATOR", http.StatusConflict)
					return
				}
				log.Error("daychlen", slog.String("op", op), slog.String("error", err.Error()))
				http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
				return
			}
			render.JSON(w, r, []*storage.GetOrderDetails{item})
			return
		default:
			log.Debug("loading internal order details", "id", id)
			sub, err = result.GetNormOrderIdSub(ctx, id)
		}
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Ошибка при получении заказов по номеру заказа")
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, sub)
	}
}

type OrderFinalGetter interface {
	GetSimpleOrderReport(ctx context.Context, orderNum string) (*storage.OrderFinalReport, error)
}

func FinalReportNormOrder(log *slog.Logger, result OrderFinalGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get.FinalReportNormOrder"

		orderNum := chi.URLParam(r, "order_num")

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		report, err := result.GetSimpleOrderReport(ctx, orderNum)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Ошибка при получении заказов по номеру заказа")
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, report)
	}
}

type OrdersPEOGetter interface {
	GetPEOProductsByCategory(ctx context.Context, filter mysql.ProductFilter) ([]storage.PEOProduct, []storage.GetWorkers, error)
}

type FinalReportResponse struct {
	Employees []storage.GetWorkers `json:"employees"`
	Products  []storage.PEOProduct `json:"products"`
}

// FinalReportNormOrders возвращает изделия и сотрудников для финального отчёта ПЭО за период
// @Summary Получить данные финального отчёта ПЭО
// @Tags norm-orders
// @Produce json
// @Param from query string false "Дата начала периода (YYYY-MM-DD), по умолчанию — начало текущего месяца"
// @Param to query string false "Дата окончания периода (YYYY-MM-DD), по умолчанию — конец текущего месяца"
// @Param order_num query string false "Номер заказа"
// @Param type query []string false "Типы изделий (можно указать несколько)" collectionFormat(multi)
// @Success 200 {object} FinalReportResponse
// @Failure 400 {string} string "Неверный формат даты 'from' или 'to'"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/all_final_order [get]
func FinalReportNormOrders(log *slog.Logger, result OrdersPEOGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.order-norm.get.FinalReportNormOrders"

		fromStr := r.URL.Query().Get("from")
		toStr := r.URL.Query().Get("to")
		orderNum := r.URL.Query().Get("order_num")
		typeIzd := r.URL.Query()["type"]

		var from, to time.Time

		parseDate := func(dateStr string, defaultTime time.Time) (time.Time, error) {
			if dateStr == "" {
				return defaultTime, nil
			}
			return time.Parse("2006-01-02", dateStr)
		}

		now := time.Now()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

		from, err := parseDate(fromStr, startOfMonth)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Warn("Неверный формат from")
			http.Error(w, "Неверный формат даты 'from'", http.StatusBadRequest)
			return
		}

		to, err = parseDate(toStr, endOfMonth)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Warn("Неверный формат to")
			http.Error(w, "Неверный формат даты 'to'", http.StatusBadRequest)
			return
		}

		filter := mysql.ProductFilter{
			From:     from,
			To:       to,
			OrderNum: orderNum,
			Type:     typeIzd,
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		products, employees, err := result.GetPEOProductsByCategory(ctx, filter)
		if err != nil {
			log.With(slog.String("op", op), slog.Any("error", err)).Error("Ошибка при получении изделий")
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		response := FinalReportResponse{
			Employees: employees,
			Products:  products,
		}

		render.JSON(w, r, response)
	}
}

type OrderNashelGetter interface {
	GetNashchelnikRawData(ctx context.Context, legacyID int64) (*storage.NashchelnikRawData, error)
}

// GetNashchelnikRawHandler возвращает данные по нащельникам для калькулятора
// @Summary Получить данные по нащельникам
// @Tags norm-orders
// @Produce json
// @Param id path int true "ID заказа (legacy)"
// @Success 200 {object} storage.NashchelnikRawData
// @Failure 400 {string} string "Invalid ID"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/orders/nashchelnik/raw/{id} [get]
func GetNashchelnikRawHandler(log *slog.Logger, storage OrderNashelGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get.GetNashchelnikRawHandler"

		idStr := chi.URLParam(r, "id")
		legacyID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		data, err := storage.GetNashchelnikRawData(ctx, legacyID)
		if err != nil {
			log.With(slog.String("op", op), slog.String("error", err.Error())).Error("Failed to get raw nashchelnik data")
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, data)
	}
}

type OrderVitrageGetter interface {
	GetNormOrderVitrage(ctx context.Context, id int64) ([]storage.GetWorkersVitrage, error)
}

// GetVitrageAssignments возвращает данные по витражам
// @Summary Получить данные по витражам
// @Tags norm-orders
// @Produce json
// @Param id path int true "ID заказа"
// @Success 200 {array} storage.GetWorkersVitrage
// @Failure 400 {string} string "Invalid ID"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/orders/{id}/vitr-assign [get]
func GetVitrageAssignments(log *slog.Logger, storage OrderVitrageGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.get.GetVitrageAssignments"

		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "неверный id заказа", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		vitrage, err := storage.GetNormOrderVitrage(ctx, id)
		if err != nil {
			log.Error("ошибка получения назначений", slog.String("op", op), slog.Any("error", err))
			http.Error(w, "ошибка получения назначений", http.StatusInternalServerError)
			return
		}

		render.JSON(w, r, vitrage)
	}
}
