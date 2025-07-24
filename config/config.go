package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	WallpaperDir     string
	RotationInterval int
	Randomize        bool
	Backend          string
	SwaybgMode       string
	SwaybgOutput     string
	MonitorName      string
	PywalEnabled     bool
	PywalCommand     string
}

func DefaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not find user config directory: %w", err)
	}
	return filepath.Join(configDir, "hpaper", "hpaper.conf"), nil
}

func DefaultConfig() *Config {
	return &Config{
		WallpaperDir:     "",
		RotationInterval: 3600,
		Randomize:        true,
		Backend:          "swaybg",
		SwaybgMode:       "fill",
		SwaybgOutput:     "",
		MonitorName:      "",
		PywalEnabled:     false,
		PywalCommand:     "wal --cols16 -n",
	}
}

func saveConfig(cfg *Config, filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create config file %s: %w", filePath, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	effectiveWallpaperDir := cfg.WallpaperDir

	fmt.Fprintf(writer, "# [General]\n\n")

	fmt.Fprintf(writer, "wallpaper_dir = %s\n", effectiveWallpaperDir)
	fmt.Fprintf(writer, "rotation_interval = %d\n", cfg.RotationInterval)
	fmt.Fprintf(writer, "randomize = %t\n", cfg.Randomize)
	fmt.Fprintf(writer, "backend = %s\n", cfg.Backend)
	fmt.Fprintf(writer, "\n# [swaybg settings]\n\n")
	fmt.Fprintf(writer, "# available modes: fill, fit, stretch, center, tile\n")
	fmt.Fprintf(writer, "swaybg_mode = %s\n", cfg.SwaybgMode)
	fmt.Fprintf(writer, "# Select an output to configure. Subsequent appearance options will only apply to this output. The special value * selects all outputs.\n")
	fmt.Fprintf(writer, "swaybg_output = %s\n", cfg.SwaybgOutput)
	fmt.Fprintf(writer, "\n# [hyprpaper settings]\n")
	fmt.Fprintf(writer, "# Example: DP-1 or all\n")

	fmt.Fprintf(writer, "monitor_name = %s\n", cfg.MonitorName)
	fmt.Fprintf(writer, "\n# [pywal settings]\n")
	fmt.Fprintf(writer, "pywal_enabled = %t\n", cfg.PywalEnabled)
	fmt.Fprintf(writer, "pywal_command = %s\n", cfg.PywalCommand)

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to write to config file: %w", err)
	}
	fmt.Printf("Default config created at %s\n", filePath)
	return nil
}

func LoadConfig(filePath string, cliWallpaperDir string) (*Config, error) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		defaultCfg := DefaultConfig()
		if cliWallpaperDir != "" {
			defaultCfg.WallpaperDir = cliWallpaperDir
		}

		fmt.Printf("Config file not found at %s. Creating with default settings.\n", filePath)
		if err := saveConfig(defaultCfg, filePath); err != nil {
			return nil, fmt.Errorf("failed to create default config file: %w", err)
		}
		return defaultCfg, nil
	} else if err != nil {
		return nil, fmt.Errorf("error checking config file %s: %w", filePath, err)
	}

	cfg := DefaultConfig()
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %w", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Warning: Skipping wrong syntax in line %d in config file %s: '%s' (expected 'key = value')\n", lineNum, filePath, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "wallpaper_dir":
			cfg.WallpaperDir = value
		case "rotation_interval":
			if i, err := strconv.Atoi(value); err == nil {
				cfg.RotationInterval = i
			} else {
				fmt.Printf("Warning: Invalid value for rotation_interval in config (line %d): '%s'. Using default (%d).\n", lineNum, value, DefaultConfig().RotationInterval)
			}
		case "randomize":
			if b, err := strconv.ParseBool(value); err == nil {
				cfg.Randomize = b
			} else {
				fmt.Printf("Warning: Invalid value for randomize in config (line %d): '%s'. Using default (%t).\n", lineNum, value, DefaultConfig().Randomize)
			}
		case "enable_pywal", "pywal_enabled":
			if b, err := strconv.ParseBool(value); err == nil {
				cfg.PywalEnabled = b
			} else {
				fmt.Printf("Warning: Invalid value for pywal_enabled in config (line %d): '%s'. Using default (%t).\n", lineNum, value, DefaultConfig().PywalEnabled)
			}
		case "pywal_command":
			cfg.PywalCommand = value
		case "backend":
			cfg.Backend = value
		case "monitor_name":
			cfg.MonitorName = value
		case "swaybg_mode":
			cfg.SwaybgMode = value
		case "swaybg_output":
			cfg.SwaybgOutput = value
		default:
			fmt.Printf("Warning: Unknown config key '%s' on line %d in file %s\n", key, lineNum, filePath)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", filePath, err)
	}

	return cfg, nil
}

func (c *Config) GetRotationDuration() time.Duration {
	return time.Duration(c.RotationInterval) * time.Second
}
