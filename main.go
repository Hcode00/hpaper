package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"hpaper/backends"
	"hpaper/config"
	"hpaper/daemon"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [arguments]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  start <wallpaper_directory> [flags]\n")
	fmt.Fprintf(os.Stderr, "    Arguments:\n")
	fmt.Fprintf(os.Stderr, "      <wallpaper_directory> : Path to the directory containing wallpapers.\n")
	fmt.Fprintf(os.Stderr, "                              (Overrides 'wallpaper_directory' in config file if provided).\n")
	fmt.Fprintf(os.Stderr, "\n    Flags for 'start':\n")
	fmt.Fprintf(os.Stderr, "      -config <path> : Path to a configuration file. Default: ~/.config/hpaper/hpaper.conf\n")
	fmt.Fprintf(os.Stderr, "\n  next\n")
	fmt.Fprintf(os.Stderr, "    Switches to the next wallpaper.\n")
	fmt.Fprintf(os.Stderr, "\n  prev\n")
	fmt.Fprintf(os.Stderr, "    Switches to the previous wallpaper.\n")
	fmt.Fprintf(os.Stderr, "\n  current\n")
	fmt.Fprintf(os.Stderr, "    Returns the path to the currently applied wallpaper.\n")
	fmt.Fprintf(os.Stderr, "\n  quit\n")
	fmt.Fprintf(os.Stderr, "    Sends a signal to the running daemon to quit.\n")
	fmt.Fprintf(os.Stderr, "\nExample:\n")
	fmt.Fprintf(os.Stderr, "  %s start ~/my_wallpapers/ \n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s next\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s current\n", os.Args[0])
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start":
		startCmd := flag.NewFlagSet("start", flag.ExitOnError)
		configPathFlag := startCmd.String("config", "", "Path to a configuration file.")

		if err := startCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("Error parsing 'start' flags: %v", err)
		}

		args := startCmd.Args()
		wallpaperDirFromCmd := ""
		if len(args) > 0 {
			wallpaperDirFromCmd = args[0]
		}

		var cfgPath string
		if *configPathFlag != "" {
			cfgPath = *configPathFlag
		} else {
			defaultPath, err := config.DefaultConfigPath()
			if err != nil {
				log.Fatalf("Error getting default config path: %v", err)
			}
			cfgPath = defaultPath
		}

		cfg, err := config.LoadConfig(cfgPath, wallpaperDirFromCmd)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}

		if wallpaperDirFromCmd != "" {
			cfg.WallpaperDir = wallpaperDirFromCmd
		}

		if cfg.WallpaperDir == "" {
			log.Fatalf("Error: Wallpaper directory is not set. Please specify it via command line (e.g., '%s start /path/to/wallpapers/') or in your config file '%s'.", os.Args[0], cfgPath)
		}

		var backend backends.WallpaperBackend
		switch strings.ToLower(strings.TrimSpace(cfg.Backend)) {
		case "swaybg":
			backend = backends.NewSwayBGBackend()
			fmt.Println("Using backend: swaybg")
		case "hyprpaper":
			backend = backends.NewHyprpaperBackend(cfg.MonitorName)
			fmt.Printf("Using backend: hyprpaper (Monitor: %s)\n", cfg.MonitorName)
		default:
			log.Fatalf("Error: Unknown backend '%s' specified in config. Supported backends: swaybg, hyprpaper.", cfg.Backend)
		}

		daemon.StartDaemon(cfg.WallpaperDir, *cfg, backend)

	case string(backends.ActionNext):
		daemon.RunClientMode(backends.ActionNext)
	case string(backends.ActionPrev):
		daemon.RunClientMode(backends.ActionPrev)
	case string(backends.ActionQuit):
		daemon.RunClientMode(backends.ActionQuit)
	case "current":
		daemon.RunClientMode(backends.ActionCurrent)
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Printf("Error: Unknown command '%s'\n", command)
		usage()
		os.Exit(1)
	}
}
