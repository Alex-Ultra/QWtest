package monitor

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

type Controller struct {
	BinaryPath string
	ConfigPath string
	LogFile    string
	cmd        *exec.Cmd
}

func NewController(binaryPath, configPath, logFile string) *Controller {
	return &Controller{
		BinaryPath: binaryPath,
		ConfigPath: configPath,
		LogFile:    logFile,
	}
}

func (m *Controller) Start() error {
	// Проверяем, запущен ли уже процесс
	if m.cmd != nil {
		if m.IsRunning() {
			return fmt.Errorf("monitor уже запущен")
		}
	}

	// Запускаем monitor с конфигурацией
	args := []string{"-c", m.ConfigPath}
	
	m.cmd = exec.Command(m.BinaryPath, args...)
	
	// Перенаправляем вывод в лог-файл
	logFile, err := openLogFile(m.LogFile)
	if err != nil {
		return err
	}
	defer logFile.Close()
	
	m.cmd.Stdout = logFile
	m.cmd.Stderr = logFile
	
	err = m.cmd.Start()
	if err != nil {
		return fmt.Errorf("не удалось запустить monitor: %v", err)
	}

	log.Println("Monitor успешно запущен")
	return nil
}

func (m *Controller) Stop() error {
	if m.cmd == nil || !m.IsRunning() {
		return fmt.Errorf("monitor не запущен")
	}

	// Завершаем процесс в зависимости от ОС
	if runtime.GOOS == "windows" {
		err := m.cmd.Process.Kill()
		if err != nil {
			return fmt.Errorf("не удалось завершить процесс: %v", err)
		}
	} else {
		err := m.cmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			return fmt.Errorf("не удалось отправить сигнал завершения: %v", err)
		}
		
		// Ждем завершения процесса
		done := make(chan error, 1)
		go func() {
			done <- m.cmd.Wait()
		}()
		
		select {
		case <-time.After(10 * time.Second):
			// Процесс не завершился за 10 секунд, принудительно убиваем
			err := m.cmd.Process.Kill()
			if err != nil {
				return fmt.Errorf("не удалось принудительно завершить процесс: %v", err)
			}
		case <-done:
			// Процесс успешно завершен
		}
	}

	m.cmd = nil
	log.Println("Monitor успешно остановлен")
	return nil
}

func (m *Controller) IsRunning() bool {
	if m.cmd == nil {
		return false
	}

	err := m.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

func (m *Controller) Restart() error {
	err := m.Stop()
	if err != nil {
		return fmt.Errorf("ошибка при остановке monitor: %v", err)
	}

	// Небольшая задержка перед перезапуском
	time.Sleep(2 * time.Second)

	err = m.Start()
	if err != nil {
		return fmt.Errorf("ошибка при запуске monitor: %v", err)
	}

	return nil
}

func (m *Controller) UpdateConfig(configPath string) error {
	// Проверяем, запущен ли процесс
	isRunning := m.IsRunning()
	
	// Если запущен, останавливаем
	if isRunning {
		err := m.Stop()
		if err != nil {
			return fmt.Errorf("не удалось остановить monitor для обновления конфига: %v", err)
		}
	}

	// Обновляем путь к конфигу
	m.ConfigPath = configPath

	// Если был запущен, запускаем снова
	if isRunning {
		err := m.Start()
		if err != nil {
			return fmt.Errorf("не удалось перезапустить monitor после обновления конфига: %v", err)
		}
	}

	log.Println("Конфигурация monitor обновлена")
	return nil
}

// Вспомогательная функция для открытия лог-файла
func openLogFile(logFile string) (*os.File, error) {
	// Создаем директорию, если не существует
	dir := filepath.Dir(logFile)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	// Открываем или создаем лог-файл
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}