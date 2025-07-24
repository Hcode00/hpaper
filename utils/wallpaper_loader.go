package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadWallpapers(wallpaperDir string, randomize bool) ([]string, error) {
	files, err := os.ReadDir(wallpaperDir)
	if err != nil {
		return nil, fmt.Errorf("could not read wallpaper directory %s: %w", wallpaperDir, err)
	}

	var wallpapers []string
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
				wallpapers = append(wallpapers, filepath.Join(wallpaperDir, name))
			}
		}
	}

	if len(wallpapers) == 0 {
		return nil, fmt.Errorf("no wallpaper images found in %s", wallpaperDir)
	}
	if randomize {
		ShuffleStrings(wallpapers)
	}
	return wallpapers, nil
}
