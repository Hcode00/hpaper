package daemon

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"hpaper/backends"
	"hpaper/config"
	"hpaper/utils"
)

type IPCMessage struct {
	Action       backends.WallpaperAction
	ResponseChan chan string
}

type WallpaperManager struct {
	wallpapers        []string
	currentIndex      int
	mu                sync.Mutex
	stopAutoRotate    chan struct{}
	commandChan       chan IPCMessage
	config            config.Config
	configPath        string
	startWallpaperDir string

	wallpaperBackend backends.WallpaperBackend
	backendName      string

	currentWallpaperPath string
	ticker               *time.Ticker
	tickerCh             <-chan time.Time
}

func StartDaemon(configPath string, wallpaperDir string, config config.Config, backendName string, backend backends.WallpaperBackend) {
	wallpapers, err := utils.LoadWallpapers(wallpaperDir, config.Randomize)
	if err != nil {
		log.Fatalf("Failed to load wallpapers: %v", err)
	}

	if len(wallpapers) == 0 {
		log.Fatalf("No wallpapers found in %s after loading.", wallpaperDir)
	}

	manager := &WallpaperManager{
		wallpapers:           wallpapers,
		currentIndex:         0,
		stopAutoRotate:       make(chan struct{}),
		commandChan:          make(chan IPCMessage),
		wallpaperBackend:     backend,
		backendName:          backendName,
		config:               config,
		configPath:           configPath,
		startWallpaperDir:    wallpaperDir,
		currentWallpaperPath: "",
	}

	if err := manager.listenForCommands(); err != nil {
		log.Fatalf("Failed to set up command listener: %v", err)
	}

	manager.RunDaemon()
}

func (wm *WallpaperManager) RunDaemon() {
	log.Println("hpaper daemon started.")

	if err := wm.SetWallpaper(wm.wallpapers[wm.currentIndex]); err != nil {
		log.Printf("Error setting initial wallpaper: %v", err)
	}

	wm.resetTicker()
	defer wm.stopTicker()

	for {
		select {
		case msg := <-wm.commandChan:
			switch msg.Action {
			case backends.ActionNext:
				wm.setNextWallpaper()
				msg.ResponseChan <- "OK"
			case backends.ActionPrev:
				wm.setPrevWallpaper()
				msg.ResponseChan <- "OK"
			case backends.ActionQuit:
				log.Println("Quit command received. Shutting down.")
				msg.ResponseChan <- "OK"
				close(wm.stopAutoRotate)
				return
			case backends.ActionCurrent:
				wm.mu.Lock()
				path := wm.currentWallpaperPath
				wm.mu.Unlock()
				msg.ResponseChan <- path
			case backends.ActionReload:
				if err := wm.handleReload(); err != nil {
					msg.ResponseChan <- fmt.Sprintf("Error: %v", err)
				} else {
					msg.ResponseChan <- "OK"
				}
			default:
				msg.ResponseChan <- "Error: Unknown command"
			}
		case <-wm.stopAutoRotate:
			log.Println("Daemon stop signal received.")
			return
		case <-wm.tickerCh:
			wm.setNextWallpaper()
		}
	}
}

func (wm *WallpaperManager) listenForCommands() error {
	sockPath := "/tmp/hpaper_daemon.sock"

	if err := os.RemoveAll(sockPath); err != nil {
		return fmt.Errorf("failed to remove old socket: %w", err)
	}

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %w", err)
	}
	log.Printf("Listening for commands on %s", sockPath)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				continue
			}
			go wm.handleClientConnection(conn)
		}
	}()
	return nil
}

func (wm *WallpaperManager) resetTicker() {
	wm.stopTicker()
	if wm.config.RotationInterval == 0 {
		log.Println("Auto-rotation is disabled.")
		wm.tickerCh = nil
		return
	}

	interval := wm.config.GetRotationDuration()
	if interval <= 0 {
		log.Println("Invalid non-positive rotation interval; disabling auto-rotation.")
		wm.tickerCh = nil
		return
	}

	log.Printf("Auto-rotation enabled every %s.", interval)
	wm.ticker = time.NewTicker(interval)
	wm.tickerCh = wm.ticker.C
}

func (wm *WallpaperManager) stopTicker() {
	if wm.ticker != nil {
		wm.ticker.Stop()
		wm.ticker = nil
	}
	wm.tickerCh = nil
}

func (wm *WallpaperManager) handleClientConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error reading from client: %v", err)
		return
	}
	cmdStr := strings.TrimSpace(string(buffer[:n]))
	action := backends.WallpaperAction(cmdStr)

	respChan := make(chan string)
	wm.commandChan <- IPCMessage{Action: action, ResponseChan: respChan}

	response := <-respChan
	if _, err := conn.Write([]byte(response)); err != nil {
		log.Printf("Error writing response to client: %v", err)
	}
}

func (wm *WallpaperManager) SetWallpaper(imagePath string) error {
	if err := wm.wallpaperBackend.SetWallpaper(imagePath, wm.config); err != nil {
		return err
	}
	wm.currentWallpaperPath = imagePath
	return nil
}

func (wm *WallpaperManager) handleReload() error {
	newCfg, err := config.LoadConfig(wm.configPath, "")
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// Resolve effective wallpaper directory: config overrides start argument.
	effectiveDir := strings.TrimSpace(newCfg.WallpaperDir)
	if effectiveDir == "" {
		effectiveDir = wm.startWallpaperDir
	}
	if effectiveDir == "" {
		return fmt.Errorf("wallpaper directory is empty after reload")
	}

	selection, err := backends.ResolveBackend(newCfg.Backend, *newCfg)
	if err != nil {
		return fmt.Errorf("failed to resolve backend: %w", err)
	}

	newWallpapers, err := utils.LoadWallpapers(effectiveDir, newCfg.Randomize)
	if err != nil {
		return fmt.Errorf("failed to reload wallpapers: %w", err)
	}
	if len(newWallpapers) == 0 {
		return fmt.Errorf("no wallpapers found in %s after reload", effectiveDir)
	}

	wm.mu.Lock()
	oldBackend := wm.wallpaperBackend
	oldBackendName := wm.backendName
	oldCurrent := wm.currentWallpaperPath
	wm.config = *newCfg
	wm.startWallpaperDir = effectiveDir
	wm.wallpapers = newWallpapers

	// Choose wallpaper to apply: keep current if still present.
	chosen := ""
	chosenIndex := 0
	if oldCurrent != "" {
		for i, p := range newWallpapers {
			if p == oldCurrent {
				chosen = p
				chosenIndex = i
				break
			}
		}
	}
	if chosen == "" {
		chosen = newWallpapers[0]
		chosenIndex = 0
	}
	wm.currentIndex = chosenIndex
	wm.mu.Unlock()

	// If backend changed, stop the old one right before applying with the new one.
	if selection.Name != oldBackendName {
		if stoppable, ok := oldBackend.(backends.StoppableBackend); ok {
			if stopErr := stoppable.Stop(); stopErr != nil {
				log.Printf("Warning: failed stopping old backend %s: %v", oldBackendName, stopErr)
			}
		}
		wm.mu.Lock()
		wm.wallpaperBackend = selection.Backend
		wm.backendName = selection.Name
		wm.mu.Unlock()
		log.Printf("Reload switched backend: %s -> %s", oldBackendName, selection.Name)
	}

	log.Printf("Reload applying wallpaper: %s", chosen)
	if err := wm.SetWallpaper(chosen); err != nil {
		return fmt.Errorf("failed to set wallpaper after reload: %w", err)
	}

	wm.resetTicker()
	return nil
}

func (wm *WallpaperManager) setNextWallpaper() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.currentIndex = (wm.currentIndex + 1) % len(wm.wallpapers)
	nextWallpaper := wm.wallpapers[wm.currentIndex]
	log.Printf("Setting next wallpaper: %s", nextWallpaper)

	if err := wm.SetWallpaper(nextWallpaper); err != nil {
		log.Printf("Error setting next wallpaper: %v", err)
	}
}

func (wm *WallpaperManager) setPrevWallpaper() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.currentIndex--
	if wm.currentIndex < 0 {
		wm.currentIndex = len(wm.wallpapers) - 1
	}
	prevWallpaper := wm.wallpapers[wm.currentIndex]
	log.Printf("Setting previous wallpaper: %s", prevWallpaper)

	if err := wm.SetWallpaper(prevWallpaper); err != nil {
		log.Printf("Error setting previous wallpaper: %v", err)
	}
}

func (wm *WallpaperManager) reloadWallpapers() error {
	newList, err := utils.LoadWallpapers(wm.config.WallpaperDir, wm.config.Randomize)
	if err != nil {
		return fmt.Errorf("failed to reload wallpapers: %w", err)
	}
	if len(newList) == 0 {
		return fmt.Errorf("no wallpapers found in %s after reload", wm.config.WallpaperDir)
	}

	wm.mu.Lock()
	oldCurrent := wm.currentWallpaperPath
	wm.wallpapers = newList
	found := -1
	for i, p := range newList {
		if p == oldCurrent {
			found = i
			break
		}
	}
	if found >= 0 {
		wm.currentIndex = found
		wm.mu.Unlock()
		log.Printf("Reloaded wallpapers (%d files). Current wallpaper preserved.", len(newList))
		return nil
	}
	wm.currentIndex = 0
	newCurrent := newList[0]
	wm.currentWallpaperPath = newCurrent
	wm.mu.Unlock()

	log.Printf("Reloaded wallpapers (%d files). Current wallpaper not found; switching to %s", len(newList), newCurrent)
	if err := wm.wallpaperBackend.SetWallpaper(newCurrent, wm.config); err != nil {
		return fmt.Errorf("failed to set wallpaper after reload: %w", err)
	}
	return nil
}
