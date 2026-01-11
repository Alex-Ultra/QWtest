package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"monitor-server/internal/api"
	"monitor-server/internal/config"
	"monitor-server/internal/ws"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatal("Ошибка загрузки конфигурации:", err)
	}

	// Установка конфигурации для аутентификации
	auth.SetConfig(cfg)

	// Инициализация WebSocket хаба
	hub := ws.NewHub()
	go hub.Run()

	// Инициализация маршрутов
	router := mux.NewRouter()
	api.SetupRoutes(router, hub, cfg)

	// Настройка статических файлов
	serveStaticFiles(cfg, router)

	// Запуск сервера
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Сервер запущен на %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func serveStaticFiles(cfg *config.Config, router *mux.Router) {
	staticDir := cfg.Paths.WebDistDir
	fs := http.FileServer(http.Dir(staticDir))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Предотвращаем доступ к внутренним файлам
		if _, err := os.Stat(staticDir + r.URL.Path); os.IsNotExist(err) {
			http.ServeFile(w, r, staticDir+"/index.html")
		} else {
			http.ServeFile(w, r, staticDir+r.URL.Path)
		}
	})
}