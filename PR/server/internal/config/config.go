// Пакет config предоставляет функциональность для загрузки и обработки конфигурационных файлов сервера
package config

import (
	"os"            // Пакет для работы с файловой системой
	"gopkg.in/yaml.v3" // Библиотека для парсинга YAML-файлов
)

// Структура Config содержит общую конфигурацию сервера
// Server - конфигурация HTTP-сервера (порт, хост)
// Tokens - карта токенов для аутентификации агентов
// Paths - пути к различным файлам, используемым сервером
// MetricsInterval - интервал обновления метрик в секундах
type Config struct {
	Server           ServerConfig            `yaml:"server"`            // Конфигурация HTTP-сервера
	Tokens           map[string]string       `yaml:"tokens"`            // Карта токенов для аутентификации агентов
	Paths            PathsConfig             `yaml:"paths"`             // Конфигурация путей к файлам
	MetricsInterval  int                     `yaml:"metrics_interval"`  // Интервал обновления метрик в секундах
}

// Структура ServerConfig содержит конфигурацию HTTP-сервера
// Port - номер порта, на котором будет работать сервер
// Host - хост, на котором будет работать сервер
type ServerConfig struct {
	Port string `yaml:"port"`  // Порт сервера
	Host string `yaml:"host"`  // Хост сервера
}

// Структура PathsConfig содержит пути к различным файлам, используемым сервером
// MonitorBinDir - директория с бинарными файлами мониторов
// ProxyBinPath - путь к бинарному файлу прокси
// ProxyConfigPath - путь к конфигурационному файлу прокси
// WebDistDir - директория с файлами веб-интерфейса
type PathsConfig struct {
	MonitorBinDir     string `yaml:"monitor_bin_dir"`    // Директория с бинарными файлами мониторов
	ProxyBinPath      string `yaml:"proxy_bin_path"`     // Путь к бинарному файлу прокси
	ProxyConfigPath   string `yaml:"proxy_config_path"`  // Путь к конфигурационному файлу прокси
	WebDistDir        string `yaml:"web_dist_dir"`       // Директория с файлами веб-интерфейса
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