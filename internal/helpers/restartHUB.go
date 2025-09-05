package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func RestartHUB(exePath string, tries int) error {
	const maxTries = 5

	var killCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		killCmd = exec.Command("taskkill", "/f", "/im", "r5apex_ds.exe")
	} else {
		fmt.Printf("[DISCORD] Unsupported OS for restarting HUB\n")
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	if out, err := killCmd.CombinedOutput(); err != nil {
		fmt.Printf("[DISCORD] failed to kill r5apex_ds: %v - %s\n", err, string(out))
	}

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("Restarting HUB at %s\n", currentTime)

	hubExePath := filepath.Join(exePath, "r5apex_ds.exe")
	if _, err := os.Stat(hubExePath); err != nil {
		return fmt.Errorf("r5apex_ds.exe not found at %s: %w", hubExePath, err)
	}

	for attempts := tries; attempts <= maxTries; attempts++ {
		startCmd := exec.Command(hubExePath)
		if err := startCmd.Start(); err != nil {
			if attempts == maxTries {
				return fmt.Errorf("failed to start r5apex_ds after %d attempts: %w", attempts, err)
			}
			time.Sleep(60 * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("failed to start r5apex_ds after %d attempts", maxTries)
}
