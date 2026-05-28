package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
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
	"vue-golang/internal/middleware/auth"
)

//type Service interface {
//	recalculate.NormService
//	generate_excel.GenerateExcel
//}

func routes(app *App) *chi.Mux {
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
	router.Get("/api/orders", getorder.GetOrdersFilter(app.Log, app.Storage))

	// Маршруты для Гловяка где он внесет все данные по заказу
	router.Get("/api/orders/order/{orderNum}", getorder.GetOrderDetails(app.Log, app.Storage))

	//TODO получение шаблонов
	router.Get("/api/template", gettemplate.GetTemplatesByCode(app.Log, app.Storage))
	router.Get("/api/all_templates", gettemplate.GetAllTemplates(app.Log, app.Storage))

	//TODO сохранение нормированных нарядов
	router.Post("/api/orders/order-norm/template", save.SaveNormOrderOperation(app.Log, app.Storage))

	//TODO обновление статуса нормировки(отмена)
	router.Post("/api/orders/cancel", update.UpdateCancelStatus(app.Log, app.Storage))

	//TODO get получение нормированного наряда
	router.Get("/api/orders/order/norm/{id}", get.GetNormOrder(app.Log, app.Storage))
	//TODO получение нескольких заказов нормирования(связанных между собой)
	router.Get("/api/orders/order-norm/by-order", get.GetNormOrdersOrderNum(app.Log, app.Storage))
	router.Get("/api/orders/order-norm/{id}/details", get.DoubleReportOrder(app.Log, app.Storage))

	//TODO get получение всех нормированных нарядов
	router.Get("/api/orders/order/norm/all", get.GetNormOrders(app.Log, app.Storage))

	//TODO update обновление нормированного наряда
	router.Put("/api/orders/order/norm/update/{id}", update.UpdateNormOrderOperation(app.Log, app.Storage))

	//TODO назначение сотрудников
	router.Post("/api/workers", saveWorkers.SaveWorkersOperation(app.Log, app.Storage))
	//TODO получение всех сотрудников
	router.Get("/api/workers/all", getWorkers.GetWorkers(app.Log, app.Storage))

	//TODO финальные маршруты для всех готовых заказов и возможность провалиться в них
	router.Get("/api/allians/{order_num}", get.FinalReportNormOrder(app.Log, app.Storage))
	router.Get("/api/all_final_order", get.FinalReportNormOrders(app.Log, app.Storage))

	//TODO финальное обновление
	router.Put("/api/final/update/{id}", update.UpdateFinalOrder(app.Log, app.Storage))

	//TODO Материалы к заказу
	router.Get("/api/materials", getmaterials.GetMaterials(app.Log, app.Storage))
	router.Post("/api/materials/calculation", recalculate_norm.CalculateNormOperations(app.Log, app.Service.RecalculateService))

	// TODO генерация excel
	router.Get("/api/report/excel", generate_excel.GenerateReportExcel(app.Log, app.Service.GenerateExcelService))

	//TODO вытягивание москиток
	//router.Post("/api/sync/aa", post.SyncButton(app.Log, app.Service.MosquitoService))
	//TODO сохранение и расчет водоотливов
	router.Post("/api/orders/nashchelnik/calc", save.SaveNashchelnikCalc(app.Log, app.Storage))
	router.Get("/api/orders/nashchelnik/raw/{id}", get.GetNashchelnikRawHandler(app.Log, app.Storage))

	//TODO adminPanel
	adminRouter := chi.NewRouter()
	adminRouter.Use(auth.BasicAuth(app.Config.AdminLogin, app.Config.AdminPass))

	adminRouter.Get("/all_templates", gettemplate.GetAllTemplatesAdmin(app.Log, app.Storage))
	adminRouter.Get("/template", gettemplate.GetTemplatesByCodeAdmin(app.Log, app.Storage))
	adminRouter.Put("/template/update/{id}", uptemplate.UpdateTemplateAdmin(app.Log, app.Storage))
	adminRouter.Post("/template/new", savetemplate.SaveTemplateAdmin(app.Log, app.Storage))
	adminRouter.Get("/coefficient", getadmincoef.GetCoefficientAdmin(app.Log, app.Storage))
	adminRouter.Put("/coefficient/update", upadmincoef.UpdateCoefficientAdmin(app.Log, app.Storage))
	adminRouter.Get("/employees", getadmincoef.GetAllEmployeesAdmin(app.Log, app.Storage))
	adminRouter.Get("/employees/teams", getadmincoef.GetAllTeams(app.Log, app.Storage))
	adminRouter.Put("/employees/update/{id}", upadmincoef.UpdateEmployeesAdmin(app.Log, app.Storage))
	adminRouter.Post("/employees/save", saveadmincoef.SaveEmployerAdmin(app.Log, app.Storage))
	//
	router.Mount("/api/admin", adminRouter)
	//
	// TODO Статика, vue
	frontendDir := "./frontend-dist"
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		app.Log.Error("Папка фронтенда не найдена", "path", frontendDir)
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

	router.With(auth.BasicAuth(app.Config.AdminLogin, app.Config.AdminPass)).Handle("/admin/*",
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
