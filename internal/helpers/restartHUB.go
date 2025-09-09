package helpers

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func RestartHUB(exePath string) error {

	var killCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		killCmd = exec.Command("taskkill", "/f", "/im", "r5apex_ds.exe")
	} else {
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	killCmd.CombinedOutput()

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("Restarting HUB at %s\n", currentTime)

	hubExePath := filepath.Join(exePath, "r5apex_ds.exe")
	cmd := exec.Command(hubExePath)
	cmd.Dir = exePath

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start r5apex_ds: %s", err)
	}

	return nil
}
