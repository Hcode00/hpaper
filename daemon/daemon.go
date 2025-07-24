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
	wallpapers           []string
	currentIndex         int
	mu                   sync.Mutex
	stopAutoRotate       chan struct{}
	commandChan          chan IPCMessage
	config               config.Config
	wallpaperBackend     backends.WallpaperBackend
	currentWallpaperPath string
}

func StartDaemon(wallpaperDir string, config config.Config, backend backends.WallpaperBackend) {
	wallpapers, err := utils.LoadWallpapers(wallpaperDir, config.Randomize)
	if err != nil {
		log.Fatalf("Failed to load wallpapers: %v", err)
	}

	if len(wallpapers) == 0 {
		log.Fatalf("No wallpapers found in %s after loading.", wallpaperDir)
	}

	manager := &WallpaperManager{
		wallpapers:       wallpapers,
		currentIndex:     0,
		stopAutoRotate:   make(chan struct{}),
		commandChan:      make(chan IPCMessage),
		wallpaperBackend: backend,
		config:           config,
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

	ticker := time.NewTicker(wm.config.GetRotationDuration())
	if wm.config.RotationInterval == 0 {
		ticker.Stop()
		log.Println("Auto-rotation is disabled.")
	} else {
		log.Printf("Auto-rotation enabled every %s.", wm.config.GetRotationDuration())
	}

	defer ticker.Stop()

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
			default:
				msg.ResponseChan <- "Error: Unknown command"
			}
		case <-ticker.C:
			wm.setNextWallpaper()
		case <-wm.stopAutoRotate:
			log.Println("Daemon stop signal received.")
			return
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
	wm.currentWallpaperPath = imagePath
	return wm.wallpaperBackend.SetWallpaper(imagePath, wm.config)
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
