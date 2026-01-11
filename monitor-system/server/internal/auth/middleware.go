package auth

import (
	"monitor-server/internal/config"
)

var globalConfig *config.Config

// Функция для установки конфигурации
func SetConfig(cfg *config.Config) {
	globalConfig = cfg
}

// Проверка токена для доступа к API
func ValidateToken(token string) bool {
	if globalConfig == nil {
		return false
	}
	
	// Проверяем, есть ли такой токен в конфигурации
	for validToken := range globalConfig.Tokens {
		if token == validToken {
			return true
		}
	}
	
	return false
}

// Проверка токена для скачивания файлов
func ValidateTokenForDownload(token, platform string) bool {
	if globalConfig == nil {
		return false
	}
	
	// Для скачивания разрешаем использовать любые действительные токены
	for validToken := range globalConfig.Tokens {
		if token == validToken {
			return true
		}
	}
	
	return false
}