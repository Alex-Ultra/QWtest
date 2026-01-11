package monitor

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Controller struct {
	binaryPath    string
	configPath    string
	logFilePath   string
	isRunning     bool
	mutex         sync.RWMutex
	cmd           *exec.Cmd
}

func NewController(binaryPath, configPath, logFilePath string) *Controller {
	return &Controller{
		binaryPath:  binaryPath,
		configPath:  configPath,
		logFilePath: logFilePath,
		isRunning:   false,
	}
}

func (mc *Controller) Start() error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	if mc.isRunning {
		return fmt.Errorf("monitor is already running")
	}
	
	// Open log file
	logFile, err := os.OpenFile(mc.logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	
	// Start the monitor process
	cmd := exec.Command(mc.binaryPath, "-c", mc.configPath)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	
	err = cmd.Start()
	if err != nil {
		logFile.Close()
		return err
	}
	
	mc.cmd = cmd
	mc.isRunning = true
	
	// Wait for the process in a goroutine to handle completion
	go func() {
		_ = cmd.Wait()
		logFile.Close()
		
		mc.mutex.Lock()
		mc.isRunning = false
		mc.cmd = nil
		mc.mutex.Unlock()
	}()
	
	return nil
}

func (mc *Controller) Stop() error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	if !mc.isRunning || mc.cmd == nil {
		return fmt.Errorf("monitor is not running")
	}
	
	err := mc.cmd.Process.Kill()
	if err != nil {
		return err
	}
	
	mc.isRunning = false
	mc.cmd = nil
	
	return nil
}

func (mc *Controller) Restart() error {
	err := mc.Stop()
	if err != nil {
		return err
	}
	
	time.Sleep(1 * time.Second)
	
	return mc.Start()
}

func (mc *Controller) IsRunning() bool {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	return mc.isRunning
}

func (mc *Controller) UpdateConfig(configPath string) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	// Copy new config to the current config path
	mc.configPath = configPath
	return nil
}