package updater

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

type BinaryUpdater struct {
	ServerURL string
	Token     string
}

func NewBinaryUpdater(serverURL, token string) *BinaryUpdater {
	return &BinaryUpdater{
		ServerURL: serverURL,
		Token:     token,
	}
}

func (bu *BinaryUpdater) DownloadMonitor(destPath string) error {
	// Определяем платформу
	platform := runtime.GOOS

	// Формируем URL для скачивания
	url := fmt.Sprintf("%s/api/v1/download/monitor/%s", bu.ServerURL, platform)

	// Создаем HTTP-запрос с авторизацией
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("не удалось создать запрос: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+bu.Token)

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер вернул ошибку: %d", resp.StatusCode)
	}

	// Создаем директорию, если не существует
	destDir := filepath.Dir(destPath)
	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		return fmt.Errorf("не удалось создать директорию: %v", err)
	}

	// Создаем файл для записи
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("не удалось создать файл: %v", err)
	}
	defer out.Close()

	// Копируем содержимое ответа в файл
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка при записи файла: %v", err)
	}

	// Устанавливаем права на выполнение для Unix-систем
	if runtime.GOOS != "windows" {
		err = os.Chmod(destPath, 0755)
		if err != nil {
			return fmt.Errorf("не удалось установить права на выполнение: %v", err)
		}
	}

	log.Printf("Monitor успешно скачан в %s", destPath)
	return nil
}

func (bu *BinaryUpdater) CheckAndDownloadMonitorIfNeeded(monitorPath string) error {
	// Проверяем, существует ли файл
	if _, err := os.Stat(monitorPath); os.IsNotExist(err) {
		log.Println("Monitor не найден, начинаем скачивание...")
		return bu.DownloadMonitor(monitorPath)
	}

	log.Println("Monitor уже существует")
	return nil
}