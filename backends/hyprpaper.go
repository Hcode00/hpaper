package backends

import (
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

	monitorPart := ""
	if h.monitorName != "" && h.monitorName != "all" {
		monitorPart = h.monitorName + ","
	} else {
		monitorPart = ","
	}

	mode := strings.TrimSpace(conf.HyprpaperMode)
	if mode == "" {
		mode = "cover"
	}

	hyprctlPath := os.Getenv("HPAPER_HYPRCTL")
	if strings.TrimSpace(hyprctlPath) == "" {
		hyprctlPath = "hyprctl"
	}

	// Preload image with retry (hyprpaper may need a moment after spawn).
	var preloadErr error
	for attempt := 1; attempt <= 10; attempt++ {
		preloadCmd := exec.Command(hyprctlPath, "hyprpaper", "preload", imagePath)
		if preloadOut, err2 := preloadCmd.CombinedOutput(); err2 != nil {
			preloadErr = fmt.Errorf("attempt %d preload failed: %v (output: %s)", attempt, err2, strings.TrimSpace(string(preloadOut)))
			time.Sleep(100 * time.Millisecond)
			continue
		} else {
			preloadErr = nil
			break
		}
	}
	if preloadErr != nil {
		return preloadErr
	}

	// Hyprpaper IPC:
	// cover (default) => hyprctl hyprpaper wallpaper "MONITOR,image"
	// tile/contain => hyprctl hyprpaper wallpaper "MONITOR,mode:image"
	// 'all' monitor or empty uses leading comma
	wallpaperArg := ""
	if mode == "cover" {
		wallpaperArg = fmt.Sprintf("%s%s", monitorPart, imagePath)
	} else {
		wallpaperArg = fmt.Sprintf("%s%s:%s", monitorPart, mode, imagePath)
	}
	cmd = exec.Command(hyprctlPath, "hyprpaper", "wallpaper", wallpaperArg)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error setting wallpaper with hyprctl hyprpaper: %v\nOutput: %s", err, output)
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
