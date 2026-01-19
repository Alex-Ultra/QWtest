// Пакет api предоставляет обработчики HTTP-запросов для API сервера
package api

import (
	"encoding/json" // Пакет для работы с JSON
	"fmt"          // Пакет для форматированного вывода
	"io"           // Пакет для работы с I/O примитивами
	"log"          // Пакет для логирования
	"net/http"     // Пакет для HTTP-сервера
	"os"           // Пакет для работы с операционной системой
	"path/filepath" // Пакет для работы с путями к файлам
	"strings"      // Пакет для работы со строками
	"sync"         // Пакет для синхронизации
	"time"         // Пакет для работы со временем

	"github.com/gorilla/mux"        // Маршрутизатор HTTP-запросов
	"github.com/gorilla/websocket"  // Библиотека для работы с WebSocket
	"pr-server/internal/auth"       // Внутренний модуль аутентификации
	"pr-server/internal/config"     // Внутренний модуль конфигурации
	"pr-server/internal/models"     // Внутренний модуль моделей данных
	"pr-server/internal/proxy"      // Внутренний модуль управления прокси
	"pr-server/internal/ws"         // Внутренний модуль WebSocket-хаба
)

// Глобальный экземпляр WebSocket Upgrader для обновления HTTP-соединения до WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Разрешаем соединения из любого источника для разработки
		// В продакшене следует ограничить этот параметр
		return true
	},
}

// Структура Handlers содержит зависимости для обработчиков API
// hub - WebSocket-хаб для управления соединениями
// cfg - конфигурация сервера
// agents - карта агентов (thread-safe)
// commands - карта команд (thread-safe)
// proxyCtrl - контроллер прокси-сервиса
type Handlers struct {
	hub       *ws.Hub        // WebSocket-хаб для управления соединениями
	cfg       *config.Config // Конфигурация сервера
	agents    *sync.Map      // Карта агентов (thread-safe)
	commands  *sync.Map      // Карта команд (thread-safe)
	proxyCtrl *proxy.Controller // Контроллер прокси-сервиса
}

// Функция InitHandlers инициализирует все маршруты API
// Принимает: маршрутизатор, WebSocket-хаб и конфигурацию сервера
// Возвращает: ничего
func InitHandlers(router *mux.Router, hub *ws.Hub, cfg *config.Config) {
	// Создаем экземпляр обработчиков с зависимостями
	handlers := &Handlers{
		hub:       hub,                              // WebSocket-хаб
		cfg:       cfg,                              // Конфигурация сервера
		agents:    &sync.Map{},                      // Потокобезопасная карта агентов
		commands:  &sync.Map{},                      // Потокобезопасная карта команд
		proxyCtrl: proxy.NewController(cfg.Paths.ProxyBinPath, cfg.Paths.ProxyConfigPath), // Контроллер прокси
	}

	// Применяем middleware аутентификации к защищенным маршрутам
	protected := router.PathPrefix("/api/v1").Subrouter()
	protected.Use(auth.AuthMiddleware(cfg))

	// Маршруты для управления агентами
	protected.HandleFunc("/agents", handlers.GetAgents).Methods("GET")           // Получить список всех агентов
	protected.HandleFunc("/agent/{id}", handlers.GetAgent).Methods("GET")       // Получить информацию об одном агенте
	protected.HandleFunc("/agent/{id}/command", handlers.SendCommand).Methods("POST") // Отправить команду агенту

	// Маршрут для очереди команд
	protected.HandleFunc("/commands/queue", handlers.GetCommandQueue).Methods("GET") // Получить очередь команд

	// Маршруты для скачивания (без аутентификации)
	router.HandleFunc("/api/v1/download/monitor/{platform}", handlers.DownloadMonitor).Methods("GET") // Скачать бинарный файл монитора

	// Маршруты для управления прокси
	protected.HandleFunc("/proxy/start", handlers.StartProxy).Methods("POST")    // Запустить прокси
	protected.HandleFunc("/proxy/stop", handlers.StopProxy).Methods("POST")      // Остановить прокси
	protected.HandleFunc("/proxy/status", handlers.GetProxyStatus).Methods("GET") // Получить статус прокси

	// Точка подключения WebSocket
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handlers.WebSocketHandler(w, r)
	})

	// Обслуживание статических файлов
	serveStaticFiles(cfg, router)
}

// Метод GetAgents возвращает список всех агентов
// Принимает: ResponseWriter и Request
// Возвращает: ничего (записывает результат в ResponseWriter)
func (h *Handlers) GetAgents(w http.ResponseWriter, r *http.Request) {
	// Создаем слайс для хранения агентов
	agents := make([]*models.Agent, 0)
	// Проходим по всем агентам в карте
	h.agents.Range(func(key, value interface{}) bool {
		// Преобразуем значение к типу *models.Agent
		agent := value.(*models.Agent)
		// Добавляем агента в слайс
		agents = append(agents, agent)
		// Продолжаем итерацию
		return true
	})

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	// Кодируем и отправляем список агентов
	json.NewEncoder(w).Encode(agents)
}

// Метод GetAgent возвращает информацию об одном агенте по ID
// Принимает: ResponseWriter и Request
// Возвращает: ничего (записывает результат в ResponseWriter)
func (h *Handlers) GetAgent(w http.ResponseWriter, r *http.Request) {
	// Получаем переменные из URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Проверяем, существует ли агент с таким ID
	if agent, ok := h.agents.Load(id); ok {
		// Устанавливаем заголовок Content-Type
		w.Header().Set("Content-Type", "application/json")
		// Кодируем и отправляем информацию об агенте
		json.NewEncoder(w).Encode(agent)
	} else {
		// Если агент не найден, отправляем ошибку 404
		http.Error(w, "Agent not found", http.StatusNotFound)
	}
}

// Метод SendCommand отправляет команду агенту
// Принимает: ResponseWriter и Request
// Возвращает: ничего (записывает результат в ResponseWriter)
func (h *Handlers) SendCommand(w http.ResponseWriter, r *http.Request) {
	// Получаем ID агента из URL
	vars := mux.Vars(r)
	agentID := vars["id"]

	// Создаем структуру для запроса команды
	var req models.CommandRequest
	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// Если произошла ошибка при чтении тела, отправляем ошибку 400
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Декодируем JSON из тела запроса
	err = json.Unmarshal(body, &req)
	if err != nil {
		// Если произошла ошибка при декодировании JSON, отправляем ошибку 400
		http.Error(w, "Error parsing request JSON", http.StatusBadRequest)
		return
	}

	// Создаем команду для отправки агенту
	command := &models.Command{
		ID:      fmt.Sprintf("%d", time.Now().UnixNano()), // Уникальный ID команды на основе времени
		AgentID: agentID,                                   // ID агента, которому направлена команда
		Action:  req.Action,                                // Действие команды
		Payload: req.Payload,                               // Дополнительные данные команды
		Created: time.Now().Unix(),                         // Время создания команды
		Sent:    false,                                     // Флаг, указывающий, была ли команда отправлена
	}

	// Сохраняем команду в карте команд
	h.commands.Store(command.ID, command)

	// Проверяем, подключен ли агент, и отправляем команду немедленно
	client, ok := h.hub.GetClient(agentID)
	if ok && client.Conn != nil {
		// Отправляем команду через WebSocket
		msgBytes, _ := json.Marshal(command)
		h.hub.SendToClient(agentID, msgBytes)
		
		// Отмечаем команду как отправленную
		command.Sent = true
		h.commands.Store(command.ID, command)
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	// Отправляем ответ об успешной постановке команды в очередь
	json.NewEncoder(w).Encode(map[string]string{"status": "command queued"})
}

// Метод GetCommandQueue возвращает очередь команд
// Принимает: ResponseWriter и Request
// Возвращает: ничего (записывает результат в ResponseWriter)
func (h *Handlers) GetCommandQueue(w http.ResponseWriter, r *http.Request) {
	// Создаем слайс для хранения команд
	commands := make([]*models.Command, 0)
	// Проходим по всем командам в карте
	h.commands.Range(func(key, value interface{}) bool {
		// Преобразуем значение к типу *models.Command
		command := value.(*models.Command)
		// Добавляем команду в слайс
		commands = append(commands, command)
		// Продолжаем итерацию
		return true
	})

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	// Кодируем и отправляем очередь команд
	json.NewEncoder(w).Encode(commands)
}

// Метод DownloadMonitor позволяет скачать бинарный файл монитора для указанной платформы
// Принимает: ResponseWriter и Request
// Возвращает: ничего (записывает результат в ResponseWriter)
func (h *Handlers) DownloadMonitor(w http.ResponseWriter, r *http.Request) {
	// Получаем платформу из URL
	vars := mux.Vars(r)
	platform := vars["platform"]

	// Определяем имя файла в зависимости от платформы
	var filename string
	switch platform {
	case "windows":
		filename = "monitor.exe" // Имя файла для Windows
	case "linux":
		filename = "monitor"     // Имя файла для Linux
	default:
		// Если платформа не поддерживается, отправляем ошибку 400
		http.Error(w, "Unsupported platform", http.StatusBadRequest)
		return
	}

	// Формируем путь к файлу
	filePath := filepath.Join(h.cfg.Paths.MonitorBinDir, filename)
	
	// Проверяем, существует ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Если файл не существует, отправляем ошибку 404
		http.Error(w, "Binary not found", http.StatusNotFound)
		return
	}

	// Устанавливаем заголовки для скачивания файла
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	// Отправляем файл пользователю
	http.ServeFile(w, r, filePath)
}

// Метод StartProxy запускает прокси-сервис
// Принимает: ResponseWriter и Request
// Возвращает: ничего (записывает результат в ResponseWriter)
func (h *Handlers) StartProxy(w http.ResponseWriter, r *http.Request) {
	// Запускаем прокси-сервис
	err := h.proxyCtrl.Start()
	if err != nil {
		// Если произошла ошибка при запуске, отправляем ошибку 500
		http.Error(w, fmt.Sprintf("Error starting proxy: %v", err), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	// Отправляем ответ об успешном запуске прокси
	json.NewEncoder(w).Encode(map[string]string{"status": "proxy started"})
}

// Метод StopProxy останавливает прокси-сервис
// Принимает: ResponseWriter и Request
// Возвращает: ничего (записывает результат в ResponseWriter)
func (h *Handlers) StopProxy(w http.ResponseWriter, r *http.Request) {
	// Останавливаем прокси-сервис
	err := h.proxyCtrl.Stop()
	if err != nil {
		// Если произошла ошибка при остановке, отправляем ошибку 500
		http.Error(w, fmt.Sprintf("Error stopping proxy: %v", err), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	// Отправляем ответ об успешной остановке прокси
	json.NewEncoder(w).Encode(map[string]string{"status": "proxy stopped"})
}

// Метод GetProxyStatus возвращает статус прокси-сервиса
// Принимает: ResponseWriter и Request
// Возвращает: ничего (записывает результат в ResponseWriter)
func (h *Handlers) GetProxyStatus(w http.ResponseWriter, r *http.Request) {
	// Получаем статус прокси-сервиса
	status := h.proxyCtrl.Status()
	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	// Кодируем и отправляем статус
	json.NewEncoder(w).Encode(status)
}

// Метод WebSocketHandler обрабатывает WebSocket-соединения от агентов
// Принимает: ResponseWriter и Request
// Возвращает: ничего
func (h *Handlers) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Обновляем HTTP-соединение до WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Если произошла ошибка при обновлении соединения, логируем и выходим
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	// Откладываем закрытие соединения
	defer conn.Close()

	// Извлекаем токен из параметра запроса
	token := r.URL.Query().Get("token")
	if token == "" {
		// Если токен не предоставлен, логируем и выходим
		log.Println("No token provided in WebSocket connection")
		return
	}

	// Находим ID агента, связанного с токеном
	var agentID string
	var found bool
	for tkn, id := range h.cfg.Tokens {
		if tkn == token {
			agentID = id
			found = true
			break
		}
	}

	if !found {
		// Если токен недействителен, логируем и выходим
		log.Printf("Invalid token in WebSocket connection: %s", token)
		return
	}

	// Создаем клиента WebSocket
	client := &ws.Client{
		ID:  agentID,                           // ID агента
		Hub: h.hub,                            // Ссылка на хаб
		Conn: conn,                            // Активное соединение
		Send: make(chan []byte, 256),          // Канал для отправки сообщений
		Agent: &models.Agent{                  // Модель данных агента
			ID: agentID,                       // ID агента
			Token: token,                      // Токен агента
		},
	}

	// Регистрируем клиента в хабе
	client.Hub.Register(client)

	// Обрабатываем входящие сообщения от клиента
	for {
		// Читаем сообщение из WebSocket
		_, message, err := conn.ReadMessage()
		if err != nil {
			// Если произошла ошибка при чтении, логируем и отменяем регистрацию клиента
			log.Printf("WebSocket read error: %v", err)
			client.Hub.Unregister(client)
			break
		}

		// Обрабатываем различные типы сообщений
		var msgMap map[string]interface{}
		err = json.Unmarshal(message, &msgMap)
		if err != nil {
			// Если произошла ошибка при разборе JSON, логируем и продолжаем
			log.Printf("Error unmarshaling WebSocket message: %v", err)
			continue
		}

		// Проверяем, являются ли данные метриками
		if _, hasHashrate := msgMap["hashrate"]; hasHashrate {
			// Это данные метрик
			var metrics models.Metrics
			metricsData, _ := json.Marshal(msgMap)
			json.Unmarshal(metricsData, &metrics)

			// Обновляем метрики агента
			if agent, ok := h.agents.Load(agentID); ok {
				// Если агент уже зарегистрирован, обновляем его данные
				a := agent.(*models.Agent)
				a.Hashrate = metrics.Hashrate
				a.CPUUsage = metrics.CPUUsage
				a.RAMUsage = metrics.RAMUsage
				a.TempCPU = metrics.TempCPU
				a.Status = metrics.Status
				a.LastSeen = time.Now().Unix()
			} else {
				// Если агент еще не зарегистрирован, создаем нового
				newAgent := &models.Agent{
					ID:       agentID,                    // ID агента
					Token:    token,                       // Токен агента
					IP:       getClientIP(r),              // IP-адрес клиента
					Platform: getStringValue(msgMap, "platform", "unknown"), // Платформа агента
					LastSeen: time.Now().Unix(),           // Время последнего подключения
					Hashrate: metrics.Hashrate,            // Хешрейт агента
					CPUUsage: metrics.CPUUsage,            // Использование CPU
					RAMUsage: metrics.RAMUsage,            // Использование RAM
					TempCPU:  metrics.TempCPU,             // Температура CPU
					Status:   metrics.Status,              // Статус агента
					Version:  getStringValue(msgMap, "version", "1.0.0"), // Версия агента
				}
				// Сохраняем нового агента в карте
				h.agents.Store(agentID, newAgent)
			}
		}
	}
}


// Функция serveStaticFiles обслуживает статические файлы веб-интерфейса
// Принимает: конфигурацию и маршрутизатор
// Возвращает: ничего
func serveStaticFiles(cfg *config.Config, router *mux.Router) {
	// Получаем директорию со статическими файлами
	staticDir := cfg.Paths.WebDistDir
	// Создаем файловый сервер
	fs := http.FileServer(http.Dir(staticDir))
	
	// Обрабатываем конкретные запросы статических файлов
	router.PathPrefix("/styles.css").Handler(http.StripPrefix("", fs))
	router.PathPrefix("/app.js").Handler(http.StripPrefix("", fs))
	router.PathPrefix("/favicon.ico").Handler(http.StripPrefix("", fs))
	router.PathPrefix("/assets/").Handler(http.StripPrefix("", fs))
	
	// Обработчик для маршрутизации одностраничного приложения (SPA) (должен быть последним)
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Не обслуживаем index.html для маршрутов API или других специфических маршрутов
		if strings.HasPrefix(r.URL.Path, "/api/") || 
		   strings.HasPrefix(r.URL.Path, "/ws") {
			http.NotFound(w, r)
			return
		}
		
		// Пытаемся сначала обслужить запрашиваемый файл
		filePath := filepath.Join(staticDir, r.URL.Path)
		if _, err := os.Stat(filePath); err == nil {
			// Файл существует, обслуживаем его
			http.ServeFile(w, r, filePath)
			return
		}
		
		// Файл не существует, обслуживаем index.html для маршрутизации SPA
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})
}

// Функция getClientIP получает реальный IP-адрес клиента за прокси
// Принимает: HTTP-запрос
// Возвращает: строку с IP-адресом клиента
func getClientIP(r *http.Request) string {
	// Получаем реальный IP, если клиент за прокси
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// Если X-Forwarded-For содержит несколько IP, берем первый
		return strings.Split(ip, ",")[0]
	}
	
	// Проверяем заголовок X-Real-IP
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	
	// Если заголовки отсутствуют, используем RemoteAddr
	host := r.RemoteAddr
	colonIndex := strings.LastIndex(host, ":")
	if colonIndex != -1 {
		// Убираем порт из адреса
		return host[:colonIndex]
	}
	
	// Возвращаем весь адрес, если не можем извлечь IP
	return host
}

// Функция getStringValue безопасно извлекает строковое значение из map
// Принимает: map, ключ и значение по умолчанию
// Возвращает: строковое значение по ключу или значение по умолчанию
func getStringValue(m map[string]interface{}, key, defaultValue string) string {
	// Проверяем, существует ли ключ в map
	if val, ok := m[key]; ok {
		// Проверяем, является ли значение строкой
		if str, ok := val.(string); ok {
			// Возвращаем строковое значение
			return str
		}
	}
	// Возвращаем значение по умолчанию
	return defaultValue
}