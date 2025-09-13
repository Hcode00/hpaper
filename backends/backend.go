package backends

import (
	"hpaper/config"
)

type WallpaperBackend interface {
	SetWallpaper(imagePath string, config config.Config) error
}

type WallpaperAction string

const (
	ActionNext    WallpaperAction = "next"
	ActionPrev    WallpaperAction = "prev"
	ActionQuit    WallpaperAction = "quit"
	ActionCurrent WallpaperAction = "current"
	ActionReload  WallpaperAction = "reload"
)

const (
	SocketPath = "/tmp/hpaper-go.sock"
)
