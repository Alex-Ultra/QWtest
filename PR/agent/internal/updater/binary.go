// Пакет updater предоставляет функциональность для обновления бинарных файлов агента
package updater

import (
	"fmt"      // Пакет для форматированного вывода
	"io"       // Пакет для работы с I/O примитивами
	"net/http" // Пакет для HTTP-запросов
	"os"       // Пакет для работы с операционной системой
	"runtime"  // Пакет для получения информации о среде выполнения
	"strings"  // Пакет для работы со строками
)

// Структура BinaryUpdater управляет процессом обновления бинарных файлов
// serverURL - URL сервера для загрузки обновлений
// token - токен аутентификации для доступа к серверу
type BinaryUpdater struct {
	serverURL string  // URL сервера для загрузки обновлений
	token     string  // Токен аутентификации для доступа к серверу
}

// Функция NewBinaryUpdater создает новый экземпляр обновляльщика бинарных файлов
// Принимает: URL сервера и токен аутентификации
// Возвращает: указатель на новый экземпляр BinaryUpdater
func NewBinaryUpdater(serverURL, token string) *BinaryUpdater {
	// Создаем и возвращаем новый экземпляр обновляльщика
	// Устанавливаем URL сервера и токен аутентификации
	return &BinaryUpdater{
		serverURL: serverURL,  // URL сервера для загрузки обновлений
		token:     token,      // Токен аутентификации
	}
}

// Метод DownloadMonitor загружает бинарный файл монитора с сервера
// Принимает: путь для сохранения бинарного файла
// Возвращает: ошибку при неудаче
func (bu *BinaryUpdater) DownloadMonitor(binaryPath string) error {
	// Определяем платформу операционной системы
	platform := runtime.GOOS
	var downloadURL string
	
	// Формируем URL для загрузки в зависимости от платформы
	switch platform {
	case "windows":
		// Для Windows формируем URL с расширением .exe
		downloadURL = fmt.Sprintf("%s/api/v1/download/monitor/windows", strings.TrimSuffix(bu.serverURL, "/"))
	case "linux":
		// Для Linux формируем URL без расширения
		downloadURL = fmt.Sprintf("%s/api/v1/download/monitor/linux", strings.TrimSuffix(bu.serverURL, "/"))
	default:
		// Если платформа не поддерживается, возвращаем ошибку
		return fmt.Errorf("unsupported platform: %s", platform)
	}
	
	// Создаем HTTP-запрос для загрузки файла
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		// Если не удалось создать запрос, возвращаем ошибку
		return err
	}
	
	// Устанавливаем заголовок аутентификации
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bu.token))
	
	// Создаем HTTP-клиент и выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Если не удалось выполнить запрос, возвращаем ошибку
		return err
	}
	// Отложенное закрытие тела ответа
	defer resp.Body.Close()
	
	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		// Если статус не OK, возвращаем ошибку
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}
	
	// Создаем файл для сохранения бинарного файла
	out, err := os.Create(binaryPath)
	if err != nil {
		// Если не удалось создать файл, возвращаем ошибку
		return err
	}
	// Отложенное закрытие файла
	defer out.Close()
	
	// Копируем содержимое ответа в файл
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		// Если не удалось скопировать данные, возвращаем ошибку
		return err
	}
	
	// Делаем файл исполняемым на Unix-системах
	if platform != "windows" {
		// Устанавливаем права доступа 0755 для Unix-систем
		err = os.Chmod(binaryPath, 0755)
		if err != nil {
			// Если не удалось изменить права доступа, возвращаем ошибку
			return err
		}
	}
	
	// Возвращаем nil как признак успешной загрузки
	return nil
}

// Метод CheckAndDownloadMonitor проверяет наличие бинарного файла и загружает его при необходимости
// Принимает: путь к бинарному файлу
// Возвращает: ошибку при неудаче
func (bu *BinaryUpdater) CheckAndDownloadMonitor(binaryPath string) error {
	// Проверяем, существует ли бинарный файл
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Если файл не существует, выводим сообщение и загружаем его
		fmt.Printf("Monitor binary not found, downloading from server...\n")
		return bu.DownloadMonitor(binaryPath)
	}
	
	// Если файл существует, возвращаем nil
	return nil
}