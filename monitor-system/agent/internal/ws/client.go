package ws

import (
	"encoding/json"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"monitor-agent/internal/config"
	"monitor-agent/internal/monitor"
	"monitor-agent/internal/updater"
)

type Client struct {
	Config      *config.AgentConfig
	Conn        *websocket.Conn
	Send        chan []byte
	reconnectCh chan bool
	mutex       sync.RWMutex
	monitorCtrl *monitor.Controller
	updater     *updater.BinaryUpdater
}

func NewClient(cfg *config.AgentConfig) (*Client, error) {
	// Создаем контроллер монитора
	monitorCtrl := monitor.NewController(
		cfg.Paths.MonitorBin,
		cfg.Paths.MonitorConfig,
		cfg.Paths.MonitorLogFile,
	)

	// Создаем модуль обновления
	updater := updater.NewBinaryUpdater(cfg.ServerURL, cfg.Token)

	// Проверяем и скачиваем Monitor, если его нет
	err := updater.CheckAndDownloadMonitorIfNeeded(cfg.Paths.MonitorBin)
	if err != nil {
		log.Printf("Предупреждение: не удалось проверить/скачать Monitor: %v", err)
	}

	client := &Client{
		Config:      cfg,
		Send:        make(chan []byte, 256),
		reconnectCh: make(chan bool),
		monitorCtrl: monitorCtrl,
		updater:     updater,
	}

	// Подключение к серверу
	err = client.connect()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) connect() error {
	// Формирование URL для подключения
	u, err := url.Parse(c.Config.ServerURL)
	if err != nil {
		return err
	}
	u.Scheme = "ws" // Заменяем на WebSocket
	u.Path = "/ws"  // Путь к WebSocket эндпоинту

	// Добавляем токен в параметры
	q := u.Query()
	q.Add("token", c.Config.Token)
	u.RawQuery = q.Encode()

	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + c.Config.Token

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		return err
	}

	c.Conn = conn
	return nil
}

func (c *Client) Start() {
	// Запуск горутин для чтения и записи
	go c.readPump()
	go c.writePump()
	go c.metricsSender()

	log.Println("Клиент успешно подключен к серверу")
}

func (c *Client) readPump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Ошибка при чтении сообщения: %v", err)
			// Попытка переподключения
			c.reconnect()
			break
		}

		// Обработка полученной команды
		go c.handleCommand(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(time.Second * 5) // Пинг каждые 5 секунд
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Канал закрыт, отправляем закрывающее сообщение
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			// Отправляем пинг
			if err := c.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("Ошибка при отправке пинга: %v", err)
				c.reconnect()
				return
			}
		}
	}
}

func (c *Client) metricsSender() {
	ticker := time.NewTicker(time.Duration(c.Config.MetricsInterval) * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		// Отправляем метрики на сервер
		metrics := getSystemMetrics()
		c.sendMetrics(metrics)
	}
}

func (c *Client) handleCommand(message []byte) {
	// Разбор JSON команды
	var cmd map[string]interface{}
	err := json.Unmarshal(message, &cmd)
	if err != nil {
		log.Printf("Ошибка при разборе команды: %v", err)
		return
	}

	action, ok := cmd["action"].(string)
	if !ok {
		log.Printf("Не найдено поле action в команде: %s", string(message))
		return
	}

	// Обработка различных команд
	switch action {
	case "start_monitor":
		err := c.monitorCtrl.Start()
		if err != nil {
			log.Printf("Ошибка при запуске monitor: %v", err)
		} else {
			log.Println("Monitor успешно запущен")
		}
	case "stop_monitor":
		err := c.monitorCtrl.Stop()
		if err != nil {
			log.Printf("Ошибка при остановке monitor: %v", err)
		} else {
			log.Println("Monitor успешно остановлен")
		}
	case "restart_monitor":
		err := c.monitorCtrl.Restart()
		if err != nil {
			log.Printf("Ошибка при перезапуске monitor: %v", err)
		} else {
			log.Println("Monitor успешно перезапущен")
		}
	case "update_config":
		payload, ok := cmd["payload"].(map[string]interface{})
		if !ok {
			log.Printf("Неверный формат payload в команде update_config")
			return
		}
		
		configPath, ok := payload["config_path"].(string)
		if !ok {
			log.Printf("Не найден config_path в payload команды update_config")
			return
		}
		
		err := c.monitorCtrl.UpdateConfig(configPath)
		if err != nil {
			log.Printf("Ошибка при обновлении конфига monitor: %v", err)
		} else {
			log.Println("Конфиг monitor успешно обновлен")
		}
	default:
		log.Printf("Неизвестная команда: %s", action)
	}
}

func (c *Client) sendMetrics(metrics map[string]interface{}) {
	// TODO: Отправка метрик на сервер
	log.Printf("Отправка метрик: %+v", metrics)
}

func (c *Client) reconnect() {
	log.Println("Попытка переподключения к серверу...")

	for {
		time.Sleep(5 * time.Second) // Ждем 5 секунд перед попыткой переподключения

		err := c.connect()
		if err != nil {
			log.Printf("Ошибка при переподключении: %v", err)
			continue // Продолжаем попытки
		}

		log.Println("Успешно переподключено к серверу")
		
		// Перезапускаем горутины
		go c.readPump()
		go c.writePump()
		
		break
	}
}

// Вспомогательная функция для получения метрик системы
func getSystemMetrics() map[string]interface{} {
	// TODO: Реализовать получение реальных метрик
	// Временная реализация - в будущем заменить на реальные данные
	return map[string]interface{}{
		"hashrate": 0.0,
		"cpu_usage": 0.0,
		"ram_usage": 0.0,
		"temp_cpu": 0,
		"status": "stopped",
		"version": "1.0.0",
	}
}