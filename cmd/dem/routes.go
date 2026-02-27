package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	getadmincoef "vue-golang/http-server/admin/get"
	saveadmincoef "vue-golang/http-server/admin/save"
	upadmincoef "vue-golang/http-server/admin/update"
	generate_excel "vue-golang/http-server/generate-report/generate-excel"
	getmaterials "vue-golang/http-server/materials/get"
	getorder "vue-golang/http-server/order-dem/get"
	"vue-golang/http-server/order-norm/get"
	"vue-golang/http-server/order-norm/save"
	"vue-golang/http-server/order-norm/update"
	recalculate_norm "vue-golang/http-server/recalculate-norm"
	gettemplate "vue-golang/http-server/template/get"
	savetemplate "vue-golang/http-server/template/save"
	uptemplate "vue-golang/http-server/template/update"
	getWorkers "vue-golang/http-server/workers/get"
	saveWorkers "vue-golang/http-server/workers/save"
	"vue-golang/internal/config"
	"vue-golang/internal/middleware/auth"
	generate_excel2 "vue-golang/internal/service/generate-excel"
	"vue-golang/internal/service/recalculate"
	"vue-golang/internal/storage/mysql"
)

//type Service interface {
//	recalculate.NormService
//	generate_excel.GenerateExcel
//}

func routes(cfg config.Config, log *slog.Logger, storage *mysql.Storage, service *recalculate.NormService, genSevice *generate_excel2.GenerateExcelService) *chi.Mux {
	router := chi.NewRouter()

	//adminUser := "admin"
	//adminPass := "your-secure-password"

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081", "http://localhost:5173"}, // Разрешаем запросы с фронтенда
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	router.Use(corsHandler.Handler)

	router.Use(middleware.RequestID)
	//ip пользователя
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	//router.Use(middleware.URLFormat)

	//TODO массив со всеми заказами из дема
	router.Get("/api/orders", getorder.GetOrdersFilter(log, storage))

	// Маршруты для Гловяка где он внесет все данные по заказу
	router.Get("/api/orders/order/{orderNum}", getorder.GetOrderDetails(log, storage))

	//TODO получение шаблонов
	router.Get("/api/template", gettemplate.GetTemplatesByCode(log, storage))
	router.Get("/api/all_templates", gettemplate.GetAllTemplates(log, storage))

	//TODO сохранение нормированных нарядов
	router.Post("/api/orders/order-norm/template", save.SaveNormOrderOperation(log, storage))

	//TODO обновление статуса нормировки(отмена)
	router.Post("/api/orders/cancel", update.UpdateCancelStatus(log, storage))

	//TODO get получение нормированного наряда
	router.Get("/api/orders/order/norm/{id}", get.GetNormOrder(log, storage))
	//TODO получение нескольких заказов нормирования(связанных между собой)
	router.Get("/api/orders/order-norm/by-order", get.GetNormOrdersOrderNum(log, storage))
	router.Get("/api/orders/order-norm/{id}", get.DoubleReportOrder(log, storage))

	//TODO get получение всех нормированных нарядов
	router.Get("/api/orders/order/norm/all", get.GetNormOrders(log, storage))

	//TODO update обновление нормированного наряда
	router.Put("/api/orders/order/norm/update/{id}", update.UpdateNormOrderOperation(log, storage))

	//TODO назначение сотрудников
	router.Post("/api/workers", saveWorkers.SaveWorkersOperation(log, storage))
	//TODO получение всех сотрудников
	router.Get("/api/workers/all", getWorkers.GetWorkers(log, storage))

	//TODO финальные маршруты для всех готовых заказов и возможность провалиться в них
	router.Get("/api/allians/{order_num}", get.FinalReportNormOrder(log, storage))
	router.Get("/api/all_final_order", get.FinalReportNormOrders(log, storage))

	//TODO финальное обновление
	router.Put("/api/final/update/{id}", update.UpdateFinalOrder(log, storage))

	//Материалы к заказу
	router.Get("/api/materials", getmaterials.GetMaterials(log, storage))
	router.Post("/api/materials/calculation", recalculate_norm.CalculateNormOperations(log, service))

	// TODO генерация excel
	router.Get("/api/report/excel", generate_excel.GenerateReportExcel(log, genSevice))

	//TODO adminPanel

	adminRouter := chi.NewRouter()
	adminRouter.Use(auth.BasicAuth(cfg.AdminLogin, cfg.AdminPass))

	adminRouter.Get("/all_templates", gettemplate.GetAllTemplatesAdmin(log, storage))
	adminRouter.Get("/template", gettemplate.GetTemplatesByCodeAdmin(log, storage))
	adminRouter.Put("/template/update/{code}", uptemplate.UpdateTemplateAdmin(log, storage))
	adminRouter.Post("/template/new", savetemplate.SaveTemplateAdmin(log, storage))
	adminRouter.Get("/coefficient", getadmincoef.GetCoefficientAdmin(log, storage))
	adminRouter.Put("/coefficient/update", upadmincoef.UpdateCoefficientAdmin(log, storage))
	adminRouter.Get("/employees", getadmincoef.GetAllEmployeesAdmin(log, storage))
	adminRouter.Put("/employees/update", upadmincoef.UpdateEmployeesAdmin(log, storage))
	adminRouter.Post("/employees/save", saveadmincoef.SaveEmployerAdmin(log, storage))
	//
	router.Mount("/api/admin", adminRouter)
	//
	// TODO Статика, vue
	frontendDir := "./frontend-dist"
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		log.Error("Папка фронтенда не найдена", "path", frontendDir)
		os.Exit(1) // или panic — лучше упасть при старте
	}

	//Отдаём статические файлы: assets/, js/, css/, img/, favicon.ico и т.д.
	fileServer := http.StripPrefix("/", http.FileServer(http.Dir(frontendDir)))

	// Регистрируем точные префиксы для ассетов
	router.Handle("/assets/*", fileServer)
	router.Handle("/js/*", fileServer)
	router.Handle("/css/*", fileServer)
	router.Handle("/img/*", fileServer)
	//router.Handle("/favicon.ico", fileServer)

	router.With(auth.BasicAuth(cfg.AdminLogin, cfg.AdminPass)).Handle("/admin/*",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join("./frontend-dist", "index.html"))
		}),
	)

	//SPA fallback: любой другой путь → index.html
	router.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, существует ли файл — если да, отдаем его
		path := filepath.Join(frontendDir, r.URL.Path)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}
		// Иначе — SPA
		http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
	})

	return router
}
