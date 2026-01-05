package backends

import (
	"fmt"
	"hpaper/config"
	"log"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

type KDEBackend struct{}

func NewKDEBackend() *KDEBackend {
	return &KDEBackend{}
}

func (k *KDEBackend) SetWallpaper(imagePath string, conf config.Config) error {
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		return fmt.Errorf("failed to make wallpaper path absolute: %w", err)
	}

	wallpaperURL := (&url.URL{Scheme: "file", Path: absPath}).String()
	script := fmt.Sprintf(`const allDesktops = desktops();
for (let i = 0; i < allDesktops.length; i++) {
  const d = allDesktops[i];
  d.wallpaperPlugin = "org.kde.image";
  d.currentConfigGroup = Array("Wallpaper", "org.kde.image", "General");
  d.writeConfig("Image", %q);
}
`, wallpaperURL)

	if err := runPlasmaEvaluateScript(script); err != nil {
		fallbackErr := runPlasmaApplyWallpaperImage(absPath)
		if fallbackErr != nil {
			return fmt.Errorf("kde wallpaper set failed via qdbus and plasma-apply-wallpaperimage: %v; fallback: %w", err, fallbackErr)
		}
	}

	if conf.PywalEnabled {
		parts := strings.Fields(conf.PywalCommand)
		if len(parts) == 0 {
			log.Println("Warning: pywal_command in config is empty. Pywal will not run.")
			return nil
		}

		pywalExecutable := parts[0]
		pywalArgs := append(parts[1:], "-i", absPath)

		log.Printf("Attempting to run pywal: %s %s", pywalExecutable, strings.Join(pywalArgs, " "))
		walCmd := exec.Command(pywalExecutable, pywalArgs...)
		walOutput, walErr := walCmd.CombinedOutput()
		if walErr != nil {
			return fmt.Errorf("error running wal with KDE backend: %w\nOutput: %s", walErr, strings.TrimSpace(string(walOutput)))
		}
	}

	return nil
}

func runPlasmaEvaluateScript(script string) error {
	qdbusBinary, err := firstAvailableBinary("qdbus", "qdbus-qt6", "qdbus6")
	if err != nil {
		return err
	}

	cmd := exec.Command(qdbusBinary, "org.kde.plasmashell", "/PlasmaShell", "org.kde.PlasmaShell.evaluateScript", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("qdbus evaluateScript failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func runPlasmaApplyWallpaperImage(absPath string) error {
	plasmaApplyBinary, err := firstAvailableBinary("plasma-apply-wallpaperimage")
	if err != nil {
		return err
	}

	cmd := exec.Command(plasmaApplyBinary, absPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("plasma-apply-wallpaperimage failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func firstAvailableBinary(candidates ...string) (string, error) {
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		if p, err := exec.LookPath(candidate); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("none of these commands were found in PATH: %s", strings.Join(candidates, ", "))
}
