// Пакет filemanager предоставляет функциональность для управления файлами и конфигурациями
package filemanager

import (
	"fmt"
	"os"
	"path/filepath"
	"log"
)

// ConfigManager структура для управления конфигурационными файлами и другими необходимыми файлами
type ConfigManager struct {
	configPath string
	defaultConfigs map[string][]byte
}

// NewConfigManager создает новый менеджер конфигураций
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
		defaultConfigs: make(map[string][]byte),
	}
}

// RegisterDefaultConfig регистрирует конфигурацию по умолчанию для файла
func (cm *ConfigManager) RegisterDefaultConfig(filePath string, defaultContent []byte) {
	cm.defaultConfigs[filePath] = defaultContent
}

// CheckAndCreateFile проверяет существование файла и создает его при необходимости
func (cm *ConfigManager) CheckAndCreateFile(filePath string, defaultContent []byte) error {
	// Проверяем, существует ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Файл не существует, создаем директорию если нужно
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("не удалось создать директорию для файла %s: %v", filePath, err)
		}

		// Создаем файл с содержимым по умолчанию
		if err := os.WriteFile(filePath, defaultContent, 0644); err != nil {
			return fmt.Errorf("не удалось создать файл %s: %v", filePath, err)
		}

		log.Printf("Файл создан: %s", filePath)
		return nil
	} else if err != nil {
		// Произошла другая ошибка при проверке файла
		return fmt.Errorf("ошибка при проверке файла %s: %v", filePath, err)
	}

	// Файл существует, проверим его содержимое
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ошибка при чтении файла %s: %v", filePath, err)
	}

	// Здесь можно добавить дополнительную проверку содержимого файла
	// Пока просто проверим, что он не пустой
	if len(content) == 0 {
		log.Printf("Файл %s пустой, перезаписываем содержимым по умолчанию", filePath)
		if err := os.WriteFile(filePath, defaultContent, 0644); err != nil {
			return fmt.Errorf("не удалось перезаписать файл %s: %v", filePath, err)
		}
	}

	log.Printf("Файл уже существует: %s", filePath)
	return nil
}

// EnsureConfigurations проверяет и создает все зарегистрированные конфигурации
func (cm *ConfigManager) EnsureConfigurations() error {
	for filePath, content := range cm.defaultConfigs {
		if err := cm.CheckAndCreateFile(filePath, content); err != nil {
			return fmt.Errorf("ошибка при проверке файла %s: %v", filePath, err)
		}
	}
	return nil
}

// EnsureFiles проверяет и создает список файлов с указанными содержимым
func EnsureFiles(fileList map[string][]byte) error {
	for filePath, content := range fileList {
		// Проверяем, существует ли файл
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// Файл не существует, создаем директорию если нужно
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("не удалось создать директорию для файла %s: %v", filePath, err)
			}

			// Создаем файл с содержимым
			if err := os.WriteFile(filePath, content, 0644); err != nil {
				return fmt.Errorf("не удалось создать файл %s: %v", filePath, err)
			}

			log.Printf("Файл создан: %s", filePath)
		} else if err != nil {
			// Произошла другая ошибка при проверке файла
			return fmt.Errorf("ошибка при проверке файла %s: %v", filePath, err)
		} else {
			log.Printf("Файл уже существует: %s", filePath)
			
			// Проверим, что все необходимые параметры присутствуют в файле
			existingContent, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("ошибка при чтении файла %s: %v", filePath, err)
			}
			
			// Если файл пустой, запишем в него содержимое по умолчанию
			if len(existingContent) == 0 {
				log.Printf("Файл %s пустой, перезаписываем содержимым по умолчанию", filePath)
				if err := os.WriteFile(filePath, content, 0644); err != nil {
					return fmt.Errorf("не удалось перезаписать файл %s: %v", filePath, err)
				}
			}
		}
	}
	return nil
}