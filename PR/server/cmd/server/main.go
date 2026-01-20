// Пакет main содержит основной исполняемый код для сервера мониторинга
package main

import (
	"log"         // Пакет для логирования событий
	"net/http"    // Пакет для HTTP-сервера
	"path/filepath" // Пакет для работы с путями к файлам

	"github.com/gorilla/mux" // Маршрутизатор HTTP-запросов
	"pr-server/internal/api"  // Внутренний модуль API-обработчиков
	"pr-server/internal/config" // Внутренний модуль для загрузки конфигурации сервера
	"pr-server/internal/ws"     // Внутренний модуль для работы с WebSocket-хабом
)

// Функция main - точка входа в приложение сервера мониторинга
// Выполняет следующие действия:
// 1. Загружает конфигурационный файл
// 2. Инициализирует WebSocket-хаб
// 3. Создает маршрутизатор и инициализирует API-обработчики
// 4. Запускает HTTP-сервер
func main() {
	// Initialize config manager
	configManager := config.NewConfigManager(filepath.Join("configs", "config.yaml"))
	
	// Ensure all required config files exist
	err := configManager.EnsureConfigs()
	if err != nil {
		log.Fatalf("Failed to ensure config files: %v", err)
	}
	
	// Load the configuration
	cfg, err := config.LoadConfig(filepath.Join("configs", "config.yaml"))
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Validate the configuration
	err = configManager.ValidateConfig(cfg)
	if err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}
	
	log.Printf("Конфигурация загружена из: configs/config.yaml")

	// Инициализируем WebSocket-хаб
	// Создает новый экземпляр хаба для управления WebSocket-соединениями
	hub := ws.NewHub()
	go hub.Run() // Запускаем хаб в отдельной горутине

	// Создаем маршрутизатор
	// Используется для маршрутизации HTTP-запросов к соответствующим обработчикам
	router := mux.NewRouter()

	// Инициализируем обработчики API
	// Устанавливает все необходимые маршруты и их обработчики
	// router - маршрутизатор HTTP-запросов
	// hub - WebSocket-хаб для обработки соединений
	// cfg - конфигурация сервера
	api.InitHandlers(router, hub, cfg)

	// Запускаем сервер
	// Формируем адрес сервера из хоста и порта из конфигурации
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Запуск сервера на %s", addr)
	// Запускаем HTTP-сервер на указанном адресе с маршрутизатором
	log.Fatal(http.ListenAndServe(addr, router))
}