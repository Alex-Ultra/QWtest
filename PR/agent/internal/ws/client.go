// Пакет ws предоставляет функциональность для работы с WebSocket-соединением агента
package ws

import (
	"encoding/json" // Пакет для работы с JSON
	"fmt"          // Пакет для форматированного вывода
	"log"          // Пакет для логирования
	"net/http"     // Пакет для HTTP-запросов
	"net/url"      // Пакет для работы с URL
	"runtime"      // Пакет для получения информации о среде выполнения
	"time"         // Пакет для работы со временем

	"github.com/gorilla/websocket" // Библиотека для работы с WebSocket
)

// Структура Client представляет собой клиент WebSocket-соединения
// URL - URL-адрес сервера для подключения
// Token - токен аутентификации
// Connection - активное WebSocket-соединение
// Send - канал для отправки сообщений
// Receive - канал для получения сообщений
// Connected - флаг состояния подключения
// OnConnect - функция-обработчик события подключения
// OnDisconnect - функция-обработчик события отключения
// OnCommand - функция-обработчик команд от сервера
type Client struct {
	URL           string                             // URL-адрес сервера для подключения
	Token         string                             // Токен аутентификации
	Connection    *websocket.Conn                    // Активное WebSocket-соединение
	Send          chan []byte                        // Канал для отправки сообщений
	Receive       chan []byte                        // Канал для получения сообщений
	Connected     bool                               // Флаг состояния подключения
	OnConnect     func()                            // Функция-обработчик события подключения
	OnDisconnect  func()                            // Функция-обработчик события отключения
	OnCommand     func(map[string]interface{})      // Функция-обработчик команд от сервера
}

// Функция NewClient создает новый экземпляр WebSocket-клиента
// Принимает: URL сервера и токен аутентификации
// Возвращает: указатель на новый экземпляр Client
func NewClient(serverURL, token string) *Client {
	// Преобразуем HTTP URL в WebSocket URL
	// Если URL начинается с "http", заменяем на "ws"
	wsURL := serverURL
	if serverURL[:4] == "http" {
		wsURL = "ws" + serverURL[4:]
	}
	
	// Создаем и возвращаем новый экземпляр клиента
	// Устанавливаем начальное состояние клиента
	return &Client{
		URL:       wsURL,                // URL для WebSocket-подключения
		Token:     token,                // Токен аутентификации
		Send:      make(chan []byte, 256), // Канал для отправки сообщений с буфером 256
		Receive:   make(chan []byte, 256), // Канал для получения сообщений с буфером 256
		Connected: false,                 // Изначально клиент не подключен
	}
}

// Метод Connect устанавливает WebSocket-соединение с сервером
// Принимает: ничего
// Возвращает: ошибку при неудаче
func (c *Client) Connect() error {
	// Разбираем URL для подключения
	// Добавляем путь "/ws" к основному URL
	u, err := url.Parse(c.URL + "/ws")
	if err != nil {
		// Если возникла ошибка при разборе URL, возвращаем ошибку
		return fmt.Errorf("ошибка разбора URL: %v", err)
	}
	
	// Добавляем токен как параметр запроса
	q := u.Query()
	q.Set("token", c.Token)
	u.RawQuery = q.Encode()
	
	// Устанавливаем заголовки запроса
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+c.Token)
	
	// Подключаемся к WebSocket
	// u.String() - полный URL для подключения
	// headers - заголовки запроса
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		// Если возникла ошибка при подключении, возвращаем ошибку
		return fmt.Errorf("ошибка подключения к WebSocket: %v", err)
	}
	
	// Сохраняем соединение и устанавливаем флаг подключения
	c.Connection = conn
	c.Connected = true
	
	// Запускаем горутины для обработки сообщений
	// readPump - читает сообщения из соединения
	// writePump - отправляет сообщения в соединение
	go c.readPump()
	go c.writePump()
	
	// Вызываем обработчик подключения, если он установлен
	if c.OnConnect != nil {
		c.OnConnect()
	}
	
	// Возвращаем nil как признак успешного подключения
	return nil
}

// Метод Disconnect закрывает WebSocket-соединение
// Принимает: ничего
// Возвращает: ничего
func (c *Client) Disconnect() {
	// Устанавливаем флаг неподключения
	c.Connected = false
	// Если соединение активно, закрываем его
	if c.Connection != nil {
		c.Connection.Close()
	}
	
	// Вызываем обработчик отключения, если он установлен
	if c.OnDisconnect != nil {
		c.OnDisconnect()
	}
}

// Метод readPump читает сообщения из WebSocket-соединения
// Принимает: ничего
// Возвращает: ничего
// Работает в отдельной горутине, постоянно читая сообщения
func (c *Client) readPump() {
	// Отложенное выполнение при завершении функции
	// Закрываем соединение и устанавливаем флаг отключения
	defer func() {
		c.Connection.Close()
		c.Connected = false
	}()
	
	// Бесконечный цикл для чтения сообщений
	for {
		// Проверяем, подключен ли клиент
		if !c.Connected {
			break
		}
		
		// Читаем сообщение из соединения
		// Возвращает тип сообщения, само сообщение и ошибку
		_, message, err := c.Connection.ReadMessage()
		if err != nil {
			// Проверяем, является ли ошибка неожиданным закрытием
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Ошибка WebSocket: %v", err)
			}
			break
		}
		
		// Разбираем полученное сообщение как JSON
		// message - байтовый срез с JSON-данными
		var msgMap map[string]interface{}
		err = json.Unmarshal(message, &msgMap)
		if err != nil {
			log.Printf("Ошибка разбора сообщения: %v", err)
			continue
		}
		
		// Отправляем сообщение в канал получения
		// Используем select с default, чтобы избежать блокировки
		select {
		case c.Receive <- message:
		default:
			// Канал полон, пропускаем сообщение
		}
		
		// Если сообщение содержит действие (command), вызываем обработчик команд
		if _, ok := msgMap["action"]; ok {
			if c.OnCommand != nil {
				c.OnCommand(msgMap)
			}
		}
	}
}

// Метод writePump отправляет сообщения в WebSocket-соединение
// Принимает: ничего
// Возвращает: ничего
// Работает в отдельной горутине, постоянно отправляя сообщения
func (c *Client) writePump() {
	// Создаем таймер для отправки пингов
	// Пинг отправляется каждые 30 секунд
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	
	// Бесконечный цикл для отправки сообщений
	for {
		// Проверяем, подключен ли клиент
		if !c.Connected {
			break
		}
		
		// Используем select для ожидания сообщений или таймера
		select {
		case message, ok := <-c.Send:
			// Проверяем, открыт ли канал
			if !ok {
				// Канал закрыт, закрываем соединение
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			// Отправляем сообщение в WebSocket
			if err := c.Connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Ошибка записи: %v", err)
				return
			}
		case <-ticker.C:
			// Отправляем пинг для поддержания соединения
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ошибка пинга: %v", err)
				return
			}
		}
	}
}

// Метод SendMetrics отправляет метрики на сервер
// Принимает: хешрейт, использование CPU, использование RAM, температуру CPU и статус
// Возвращает: ошибку при неудаче
func (c *Client) SendMetrics(hashrate float64, cpuUsage float64, ramUsage float64, tempCPU int, status string) error {
	// Создаем карту метрик для отправки
	metrics := map[string]interface{}{
		"hashrate":  hashrate,   // Хешрейт системы
		"cpu_usage": cpuUsage,   // Использование CPU
		"ram_usage": ramUsage,   // Использование RAM
		"temp_cpu":  tempCPU,    // Температура CPU
		"status":    status,     // Статус системы
		"platform":  getPlatformInfo(), // Информация о платформе (ОС/архитектура)
		"version":   getVersionInfo(),  // Версия агента
	}
	
	// Маршалим метрики в JSON
	// metrics - карта данных для отправки
	// data - байтовый срез с JSON-представлением метрик
	data, err := json.Marshal(metrics)
	if err != nil {
		// Если возникла ошибка при маршалинге, возвращаем ошибку
		return err
	}
	
	// Отправляем метрики в канал отправки
	// Используем select с default, чтобы избежать блокировки
	select {
	case c.Send <- data:
	default:
		// Канал отправки полон, возвращаем ошибку
		return fmt.Errorf("канал отправки заполнен")
	}
	
	// Возвращаем nil как признак успешной отправки
	return nil
}

// Метод Reconnect переподключается к серверу
// Принимает: ничего
// Возвращает: ошибку при неудаче
func (c *Client) Reconnect() error {
	// Отключаемся от текущего сервера
	c.Disconnect()
	// Ждем 5 секунд перед повторным подключением
	time.Sleep(5 * time.Second)
	// Подключаемся снова
	return c.Connect()
}

// Вспомогательные функции для получения динамической информации

// Функция getPlatformInfo возвращает информацию о платформе (операционная система и архитектура)
// Принимает: ничего
// Возвращает: строку с информацией о платформе
func getPlatformInfo() string {
	// Получаем операционную систему и архитектуру
	os := runtime.GOOS    // Операционная система (например, "linux", "windows", "darwin")
	arch := runtime.GOARCH // Архитектура (например, "amd64", "arm64")
	// Формируем и возвращаем строку в формате "ОС/архитектура"
	return fmt.Sprintf("%s/%s", os, arch)
}

// Функция getVersionInfo возвращает информацию о версии агента
// Принимает: ничего
// Возвращает: строку с информацией о версии
func getVersionInfo() string {
	// В реальной реализации это возвращало бы актуальную версию сборки
	// Сейчас возвращаем заглушку, которая может быть заменена во время сборки
	return "dynamic_version" // Это обычно устанавливается через флаги сборки
}