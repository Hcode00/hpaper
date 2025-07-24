package backends

import (
	"fmt"
	"hpaper/config"
	"log"
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
		fmt.Println("Hyprpaper is not running, starting it now...")
		startCmd := exec.Command("hyprpaper")
		startCmd.Stdout = nil
		startCmd.Stderr = nil
		if err := startCmd.Start(); err != nil {
			return fmt.Errorf("failed to start hyprpaper: %v", err)
		}
	}
	time.Sleep(1 * time.Second)
	fmt.Println("Hyprpaper Started.")
	fmt.Printf("Setting wallpaper with Hyprpaper to: %s\n", imagePath)

	monitorPart := ""
	if h.monitorName != "" && h.monitorName != "all" {
		monitorPart = h.monitorName + ","
	} else {
		monitorPart = ","
	}

	cmd = exec.Command("hyprctl", "hyprpaper", "reload", fmt.Sprintf("%s%s", monitorPart, imagePath))
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
