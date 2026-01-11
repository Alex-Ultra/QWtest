package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

type BinaryUpdater struct {
	serverURL string
	token     string
}

func NewBinaryUpdater(serverURL, token string) *BinaryUpdater {
	return &BinaryUpdater{
		serverURL: serverURL,
		token:     token,
	}
}

func (bu *BinaryUpdater) DownloadMonitor(binaryPath string) error {
	// Determine platform
	platform := runtime.GOOS
	var downloadURL string
	
	switch platform {
	case "windows":
		downloadURL = fmt.Sprintf("%s/api/v1/download/monitor/windows", strings.TrimSuffix(bu.serverURL, "/"))
	case "linux":
		downloadURL = fmt.Sprintf("%s/api/v1/download/monitor/linux", strings.TrimSuffix(bu.serverURL, "/"))
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bu.token))
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}
	
	// Create the destination file
	out, err := os.Create(binaryPath)
	if err != nil {
		return err
	}
	defer out.Close()
	
	// Copy response body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	
	// Make executable on Unix systems
	if platform != "windows" {
		err = os.Chmod(binaryPath, 0755)
		if err != nil {
			return err
		}
	}
	
	return nil
}

func (bu *BinaryUpdater) CheckAndDownloadMonitor(binaryPath string) error {
	// Check if binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		fmt.Printf("Monitor binary not found, downloading from server...\n")
		return bu.DownloadMonitor(binaryPath)
	}
	
	return nil
}