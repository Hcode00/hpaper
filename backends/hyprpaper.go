package backends

import (
	"encoding/json"
	"fmt"
	"hpaper/config"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type HyprpaperBackend struct {
	monitorName string
}

func NewHyprpaperBackend(monitor string) *HyprpaperBackend {
	return &HyprpaperBackend{
		monitorName: monitor,
	}
}

type hyprctlMonitor struct {
	Name string `json:"name"`
}

const (
	hyprpaperRetryCount = 10
	hyprpaperRetryDelay = 100 * time.Millisecond
)

func (h *HyprpaperBackend) SetWallpaper(imagePath string, conf config.Config) error {
	cmd := exec.Command("pidof", "hyprpaper")
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("Hyprpaper not running; starting it...")
		startCmd := exec.Command("hyprpaper")
		startCmd.Stdout = nil
		startCmd.Stderr = nil
		if e := startCmd.Start(); e != nil {
			return fmt.Errorf("failed to start hyprpaper: %v", e)
		}
	} else {
		fmt.Printf("Hyprpaper already running (PIDs: %s)\n", strings.TrimSpace(string(output)))
	}
	fmt.Printf("Setting wallpaper with Hyprpaper to: %s\n", imagePath)

	mode := strings.TrimSpace(conf.HyprpaperMode)
	if mode == "" {
		mode = "cover"
	}

	hyprctlPath := os.Getenv("HPAPER_HYPRCTL")
	if strings.TrimSpace(hyprctlPath) == "" {
		hyprctlPath = "hyprctl"
	}

	// Hyprpaper IPC:
	// hyprctl hyprpaper wallpaper "MONITOR,PATH[,FIT]"
	// FIT is optional and defaults to cover.
	// Empty MONITOR lets hyprpaper apply the default/all outputs.
	var monitorNames []string
	if h.monitorName == "all" {
		names, err := listHyprpaperMonitors(hyprctlPath)
		if err != nil {
			return err
		}
		monitorNames = names
	} else if h.monitorName == "" {
		monitorNames = []string{""}
	} else {
		monitorNames = []string{h.monitorName}
	}

	for _, monitorName := range monitorNames {
		wallpaperArg := buildHyprpaperWallpaperArg(monitorName, imagePath, mode)
		if err := applyHyprpaperWallpaper(hyprctlPath, wallpaperArg); err != nil {
			return err
		}
	}

	fmt.Println("Hyprpaper wallpaper set successfully.")

	if conf.PywalEnabled {
		parts := strings.Fields(conf.PywalCommand)
		if len(parts) == 0 {
			log.Println("Warning: pywal_command in config is empty. Pywal will not run.")
			return nil
		}

		pywalExecutable := parts[0]
		pywalArgs := parts[1:]

		pywalArgs = append(pywalArgs, "-i", imagePath)

		log.Printf("Attempting to run pywal: %s %s", pywalExecutable, strings.Join(pywalArgs, " "))
		walCmd := exec.Command(pywalExecutable, pywalArgs...)

		walOutput, walErr := walCmd.CombinedOutput()
		if walErr != nil {
			return fmt.Errorf("error running wal with Hyprpaper: %v\nOutput: %s", walErr, walOutput)
		}
		fmt.Println("Pywal updated successfully.")
	}

	return nil
}

func listHyprpaperMonitors(hyprctlPath string) ([]string, error) {
	cmd := exec.Command(hyprctlPath, "monitors", "-j")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list hyprctl monitors: %v (output: %s)", err, strings.TrimSpace(string(output)))
	}

	var monitors []hyprctlMonitor
	if err := json.Unmarshal(output, &monitors); err != nil {
		return nil, fmt.Errorf("failed to parse hyprctl monitors output: %v", err)
	}

	names := make([]string, 0, len(monitors))
	for _, monitor := range monitors {
		if monitor.Name != "" {
			names = append(names, monitor.Name)
		}
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("hyprctl monitors returned no monitor names")
	}

	return names, nil
}

func buildHyprpaperWallpaperArg(monitorName, imagePath, mode string) string {
	args := []string{monitorName, imagePath}
	if mode != "cover" {
		args = append(args, mode)
	}
	return strings.Join(args, ",")
}

func applyHyprpaperWallpaper(hyprctlPath, wallpaperArg string) error {
	var lastErr error
	for attempt := 1; attempt <= hyprpaperRetryCount; attempt++ {
		cmd := exec.Command(hyprctlPath, "hyprpaper", "wallpaper", wallpaperArg)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		lastErr = fmt.Errorf("attempt %d failed to set wallpaper %q: %v (output: %s)", attempt, wallpaperArg, err, strings.TrimSpace(string(output)))
		time.Sleep(hyprpaperRetryDelay)
	}
	return lastErr
}
