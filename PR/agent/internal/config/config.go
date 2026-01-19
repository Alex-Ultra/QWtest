// Пакет config предоставляет функциональность для загрузки и обработки конфигурационных файлов агента
package config

import (
	"os"            // Пакет для работы с файловой системой
	"gopkg.in/yaml.v3" // Библиотека для парсинга YAML-файлов
)

// Структура Config содержит общую конфигурацию агента
// ServerURL - URL-адрес сервера, к которому подключается агент
// Token - токен аутентификации для подключения к серверу
// Paths - пути к различным файлам, используемым агентом
// MetricsInterval - интервал отправки метрик в секундах
type Config struct {
	ServerURL       string       `yaml:"server_url"`     // URL сервера для подключения
	Token           string       `yaml:"token"`          // Токен аутентификации
	Paths           PathsConfig  `yaml:"paths"`          // Конфигурация путей к файлам
	MetricsInterval int          `yaml:"metrics_interval"` // Интервал отправки метрик в секундах
}

// Структура PathsConfig содержит пути к различным файлам, используемым агентом
// MonitorBin - путь к исполняемому файлу монитора
// MonitorConfig - путь к конфигурационному файлу монитора
// MonitorLogFile - путь к лог-файлу монитора
type PathsConfig struct {
	MonitorBin      string `yaml:"monitor_bin"`       // Путь к исполняемому файлу монитора
	MonitorConfig   string `yaml:"monitor_config"`    // Путь к конфигурационному файлу монитора
	MonitorLogFile  string `yaml:"monitor_log_file"`  // Путь к лог-файлу монитора
}

// Функция LoadConfig загружает конфигурацию из YAML-файла
// Принимает: путь к конфигурационному файлу
// Возвращает: указатель на структуру Config и ошибку, если произошла проблема при загрузке
func LoadConfig(path string) (*Config, error) {
	// Читаем содержимое файла конфигурации
	// path - путь к YAML-файлу конфигурации
	data, err := os.ReadFile(path)
	if err != nil {
		// Если возникла ошибка при чтении файла, возвращаем nil и ошибку
		return nil, err
	}

	// Создаем переменную для хранения конфигурации
	var config Config
	// Десериализуем YAML-данные в структуру Config
	// data - байтовый срез с содержимым YAML-файла
	// &config - указатель на структуру, в которую будут записаны данные
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		// Если возникла ошибка при разборе YAML, возвращаем nil и ошибку
		return nil, err
	}

	// Возвращаем указатель на заполненную структуру Config и nil как ошибку
	return &config, nil
}