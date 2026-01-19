// Пакет models предоставляет структуры данных для представления агентов, команд и метрик
package models

// Структура Agent представляет данные агента мониторинга
// ID - уникальный идентификатор агента
// Token - токен аутентификации агента (не сериализуется в JSON)
// IP - IP-адрес агента
// Platform - платформа агента (например, "windows", "linux")
// LastSeen - время последнего подключения агента (в формате Unix timestamp)
// Hashrate - хешрейт агента
// CPUUsage - использование CPU в процентах
// RAMUsage - использование RAM в процентах
// TempCPU - температура процессора в градусах Цельсия
// Status - статус агента ("online", "mining", "stopped", "offline")
// Version - версия агента
type Agent struct {
	ID          string  `json:"id"`              // Уникальный идентификатор агента
	Token       string  `json:"-"`               // Токен аутентификации агента (не сериализуется в JSON)
	IP          string  `json:"ip"`              // IP-адрес агента
	Platform    string  `json:"platform"`        // Платформа агента (например, "windows", "linux")
	LastSeen    int64   `json:"last_seen"`       // Время последнего подключения агента (Unix timestamp)
	Hashrate    float64 `json:"hashrate"`        // Хешрейт агента
	CPUUsage    float64 `json:"cpu_usage"`       // Использование CPU в процентах
	RAMUsage    float64 `json:"ram_usage"`       // Использование RAM в процентах
	TempCPU     int     `json:"temp_cpu"`        // Температура процессора в градусах Цельсия
	Status      string  `json:"status"`          // Статус агента ("online", "mining", "stopped", "offline")
	Version     string  `json:"version"`         // Версия агента
}

// Структура Command представляет команду, отправляемую агенту
// ID - уникальный идентификатор команды
// AgentID - идентификатор агента, которому направлена команда
// Action - действие команды ("start_monitor", "stop_monitor", "update_config", "reboot")
// Payload - дополнительные данные команды
// Created - время создания команды (в формате Unix timestamp)
// Sent - флаг, указывающий, была ли команда отправлена агенту
type Command struct {
	ID       string      `json:"id"`              // Уникальный идентификатор команды
	AgentID  string      `json:"agent_id"`        // Идентификатор агента, которому направлена команда
	Action   string      `json:"action"`          // Действие команды ("start_monitor", "stop_monitor", "update_config", "reboot")
	Payload  interface{} `json:"payload,omitempty"` // Дополнительные данные команды
	Created  int64       `json:"created"`         // Время создания команды (Unix timestamp)
	Sent     bool        `json:"sent"`            // Флаг отправки команды агенту
}

// Структура Metrics представляет метрики производительности системы
// Hashrate - хешрейт системы
// CPUUsage - использование CPU в процентах
// RAMUsage - использование RAM в процентах
// TempCPU - температура процессора в градусах Цельсия
// Status - статус системы
type Metrics struct {
	Hashrate float64 `json:"hashrate"`        // Хешрейт системы
	CPUUsage float64 `json:"cpu_usage"`       // Использование CPU в процентах
	RAMUsage float64 `json:"ram_usage"`       // Использование RAM в процентах
	TempCPU  int     `json:"temp_cpu"`        // Температура процессора в градусах Цельсия
	Status   string  `json:"status"`          // Статус системы
}

// Структура CommandRequest представляет запрос команды от API
// Action - действие команды
// Payload - дополнительные данные команды
type CommandRequest struct {
	Action  string      `json:"action"`          // Действие команды
	Payload interface{} `json:"payload,omitempty"` // Дополнительные данные команды
}