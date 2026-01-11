package proxy

import (
	"fmt"
	"os/exec"
	"sync"
)

type Controller struct {
	binaryPath    string
	configPath    string
	isRunning     bool
	mutex         sync.RWMutex
	cmd           *exec.Cmd
}

func NewController(binaryPath, configPath string) *Controller {
	return &Controller{
		binaryPath: binaryPath,
		configPath: configPath,
		isRunning:  false,
	}
}

func (pc *Controller) Start() error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if pc.isRunning {
		return fmt.Errorf("proxy is already running")
	}
	
	// Start the proxy process
	cmd := exec.Command(pc.binaryPath, "-c", pc.configPath)
	err := cmd.Start()
	if err != nil {
		return err
	}
	
	pc.cmd = cmd
	pc.isRunning = true
	
	// Wait for the process in a goroutine to handle completion
	go func() {
		_ = cmd.Wait()
		pc.mutex.Lock()
		pc.isRunning = false
		pc.cmd = nil
		pc.mutex.Unlock()
	}()
	
	return nil
}

func (pc *Controller) Stop() error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if !pc.isRunning || pc.cmd == nil {
		return fmt.Errorf("proxy is not running")
	}
	
	err := pc.cmd.Process.Kill()
	if err != nil {
		return err
	}
	
	pc.isRunning = false
	pc.cmd = nil
	
	return nil
}

func (pc *Controller) Restart() error {
	err := pc.Stop()
	if err != nil {
		return err
	}
	
	return pc.Start()
}

func (pc *Controller) IsRunning() bool {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	
	return pc.isRunning
}

func (pc *Controller) Status() map[string]interface{} {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	
	return map[string]interface{}{
		"is_running": pc.isRunning,
		"binary_path": pc.binaryPath,
		"config_path": pc.configPath,
	}
}