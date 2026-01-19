// Пакет main содержит основной исполняемый код для сервера мониторинга
package main

import (
	"fmt"         // Пакет для форматированного ввода-вывода
	"log"         // Пакет для логирования событий
	"net/http"    // Пакет для HTTP-сервера
	"os"          // Пакет для работы с операционной системой
	"path/filepath" // Пакет для работы с путями к файлам

	"github.com/gorilla/mux" // Маршрутизатор HTTP-запросов
	"pr-server/internal/api"  // Внутренний модуль API-обработчиков
	"pr-server/internal/config" // Внутренний модуль для загрузки конфигурации сервера
	"pr-server/internal/ws"     // Внутренний модуль для работы с WebSocket-хабом
	filemanager "pr-common/filemanager" // Модуль для управления файлами и конфигурациями
)

// Функция main - точка входа в приложение сервера мониторинга
// Выполняет следующие действия:
// 1. Проверяет и создает необходимые файлы (конфигурации, директории и т.д.)
// 2. Загружает конфигурационный файл
// 3. Инициализирует WebSocket-хаб
// 4. Создает маршрутизатор и инициализирует API-обработчики
// 5. Запускает HTTP-сервер
func main() {
	// Проверяем и создаем необходимые файлы
	err := ensureRequiredFiles()
	if err != nil {
		log.Printf("Предупреждение: ошибка при проверке необходимых файлов: %v", err)
	}

	// Определяем путь к конфигурационному файлу относительно местоположения исполняемого файла
	configPath := filepath.Join("configs", "config.yaml")
	
	// Также проверяем абсолютный путь или рабочую директорию
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		// Пробуем альтернативные пути для совместимости между платформами
		// Проверяем, находимся ли мы в подкаталоге, пробуем относительный путь к корню
		altConfigPath := filepath.Join("..", "configs", "config.yaml")
		cfg, err = config.LoadConfig(altConfigPath)
		if err != nil {
			// Проверяем, существует ли конфигурационный файл в текущей директории
			if _, statErr := os.Stat("config.yaml"); statErr == nil {
				cfg, err = config.LoadConfig("config.yaml")
				if err != nil {
					log.Fatalf("Не удалось загрузить конфигурацию из любого местоположения: исходный путь '%s': %v, альтернативный путь '%s': %v, локальный 'config.yaml': %v", 
						configPath, err, altConfigPath, err, err)
				}
				log.Printf("Конфигурация загружена из локальной директории")
			} else {
				log.Fatalf("Не удалось загрузить конфигурацию из любого местоположения: исходный путь '%s': %v, альтернативный путь '%s': %v, локальный 'config.yaml' не найден: %v", 
					configPath, err, altConfigPath, err, statErr)
			}
		} else {
			log.Printf("Конфигурация загружена из альтернативного пути: %s", altConfigPath)
		}
	} else {
		log.Printf("Конфигурация загружена из: %s", configPath)
	}

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

// ensureRequiredFiles проверяет наличие и создает необходимые файлы приложения
func ensureRequiredFiles() error {
	// Определяем пути к необходимым файлам
	configPath := "configs/config.yaml"
	proxyConfigPath := "configs/proxy-config.json"
	
	// Создаем карту файлов с их содержимым по умолчанию
	files := map[string][]byte{
		configPath: []byte(`# Конфигурация сервера мониторинга
server:
  # Порт, на котором будет работать сервер
  port: "8080"
  # Хост, на котором будет работать сервер (0.0.0.0 для прослушивания всех интерфейсов)
  host: "0.0.0.0"

# Токены аутентификации для агентов
# Каждый токен связан с уникальным ID агента
tokens:
  "abc123": "agent-1"      # Токен для первого агента
  "def456": "agent-2"      # Токен для второго агента
  "xyz789": "agent-3"      # Токен для третьего агента
  "web-interface-token": "web-interface"  # Специальный токен для веб-интерфейса

# Пути к различным файлам и директориям
paths:
  # Директория с бинарными файлами мониторов
  monitor_bin_dir: "./assets/bins/monitor/"
  # Путь к бинарному файлу прокси-сервера
  proxy_bin_path: "./assets/bins/proxy/xmrig-proxy"
  # Путь к конфигурационному файлу прокси-сервера
  proxy_config_path: "./configs/proxy-config.json"
  # Директория с файлами веб-интерфейса
  web_dist_dir: "./dist/"

# Интервал обновления метрик в секундах
metrics_interval: 30
`),
		proxyConfigPath: []byte(`{
  "listen": ":3333",
  "pools": [
    {
      "url": "pool.example.com:3333",
      "user": "your-wallet-address",
      "pass": "x",
      "keepalive": true,
      "enabled": true,
      "tls": false
    }
  ],
  "access-log-file": "./logs/access.log",
  "access-log-limit": 1000,
  "log-file": "./logs/proxy.log",
  "log-limit": 1000,
  "workers": 1
}
`),
	}

	// Создаем необходимые директории
	dirs := []string{"configs", "assets/bins/monitor", "assets/bins/proxy", "logs", "dist"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("не удалось создать директорию %s: %v", dir, err)
		}
	}

	// Проверяем и создаем файлы
	if err := filemanager.EnsureFiles(files); err != nil {
		return fmt.Errorf("ошибка при проверке и создании файлов: %v", err)
	}

	return nil
}