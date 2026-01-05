package backends

import (
	"fmt"
	"hpaper/config"
	"os"
	"strings"
)

type BackendSelection struct {
	Name    string
	Backend WallpaperBackend
}

func ResolveBackend(name string, cfg config.Config) (BackendSelection, error) {
	selected := strings.ToLower(strings.TrimSpace(name))
	if selected == "" {
		selected = "auto"
	}

	if selected == "auto" {
		xdgDesktop := strings.ToUpper(os.Getenv("XDG_CURRENT_DESKTOP"))
		if strings.Contains(xdgDesktop, "KDE") {
			selected = "kde"
		} else {
			selected = "swaybg"
		}
	}

	switch selected {
	case "swaybg":
		return BackendSelection{Name: "swaybg", Backend: NewSwayBGBackend()}, nil
	case "hyprpaper":
		return BackendSelection{Name: "hyprpaper", Backend: NewHyprpaperBackend(cfg.MonitorName)}, nil
	case "kde":
		return BackendSelection{Name: "kde", Backend: NewKDEBackend()}, nil
	default:
		return BackendSelection{}, fmt.Errorf("unknown backend %q (supported: auto, swaybg, hyprpaper, kde)", name)
	}
}
