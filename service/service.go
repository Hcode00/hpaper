package service

import (
	"errors"
	"os"
	"strconv"
	"syscall"
	"time"

	u "github.com/Hcode00/hpaper/utils"
	sway "github.com/Hcode00/hpaper/utils/backends/swaybg"
	"github.com/sevlyar/go-daemon"
)

const USAGE = `Usage:
  hpaper help                     Show help information

  hpaper start [directory] [duration in seconds] [maximum number of pictures to preload] [-r]
    -r (optional)                 Randomize wallpaper list

  hpaper start [image file]       Set a single image as wallpaper

  hpaper download [directory] [number of pictures] [width] [height] [-w]
    -w (optional)                              Download in .webp extension

  Other Commands:
  hpaper [next|prev|status|quit]  Control running hpaper instance

  next -> set next wallpaper in the list

  prev -> set previous wallpaper in the list

  status -> show current wallpaper name and preloaded wallpapers

  quit -> stop rotaing wallpapers

Examples:
  hpaper start /path/to/wallpapers 300 3
  hpaper start /path/to/wallpapers 300 3 -r
  hpaper start /path/to/image.jpg
  hpaper download /path/to/save 5 1920 1080
  hpaper next

Notes:
  - For the download option, the number of pictures must be between 1 and 20.
  - Duration is specified in seconds.
  - Maximum number of pictures to preload must be a positive integer.
`

func Help() {
	print(USAGE)
}

var (
	nextChan = make(chan struct{}, 1)
	prevChan = make(chan struct{}, 1)
	quitChan = make(chan struct{}, 1)
)

type Hpaper struct {
	Path       string
	List       []string
	MaxToLoad  uint
	CurrentIdx uint
	Interval   time.Duration
	Randomize  bool
}

func (hpaper *Hpaper) StartSwaybgService() error {
	u.LOG.Debug("Starting service...")
	hpaper.Path = os.Args[2]

	files := u.ListFiles(hpaper.Path)
	if len(files) == 0 {
		return errors.New("This Directory does not contain any of the supported image formats")
	}
	hpaper.List = make([]string, 0, len(files))
	u.LOG.Debug("MaxToLoad = " + strconv.Itoa(int(hpaper.MaxToLoad)))
	if hpaper.Path[len(hpaper.Path)-1] != '/' {
		hpaper.Path += "/"
	}
	if hpaper.Randomize {
		randomList := u.RandomizeFileNames(files)
		for _, file := range randomList {
			hpaper.List = append(hpaper.List, hpaper.Path+file)
		}
	} else {
		for _, file := range files {
			hpaper.List = append(hpaper.List, hpaper.Path+file)
		}
	}

	go WaitAndSetSway(hpaper.Interval, hpaper.List, &hpaper.CurrentIdx)
	return nil
}

func TermHandler(sig os.Signal) error {
	if sig == syscall.SIGQUIT || sig == syscall.SIGTERM {
		u.LOG.Warn("terminating ...")
		quitChan <- struct{}{}
	}
	return daemon.ErrStop
}

func NextHandler(sig os.Signal) error {
	if sig == syscall.SIGUSR1 {
		nextChan <- struct{}{}
	}
	return nil
}

func PrevHandler(sig os.Signal) error {
	if sig == syscall.SIGUSR2 {
		prevChan <- struct{}{}
	}
	return nil
}

func WaitAndSetSway(t time.Duration, list []string, Index *uint) {
	if len(list) == 0 {
		return
	}
	sway.UnloadAll()
	curr := *Index % uint(len(list))

	err := sway.SetWallpaper(list[curr])
	if err != nil {
		u.LOG.Error("Failed to Set wallpaper" + list[curr] + "->" + err.Error())
	}
	ticker := time.NewTicker(t)
	defer ticker.Stop()
	resetTicker := func() {
		ticker.Stop()
		ticker = time.NewTicker(t)
	}
	for {
		select {
		case <-ticker.C:
			moveToNextSway(list, &curr)
		case <-nextChan:
			u.LOG.Debug("Next signal triggered")
			moveToNextSway(list, &curr)
			resetTicker()
		case <-prevChan:
			u.LOG.Debug("Prev signal triggered")
			moveToPrevSway(list, &curr)
			resetTicker()
		case <-quitChan:
			u.LOG.Warn("Quitting wallpaper rotation...")
			return
		}
	}
}

func moveToNextSway(list []string, curr *uint) {
	nextIndex := (*curr + 1) % uint(len(list))
	currentProcess := sway.GetCurrentProcess()
	sway.SetWallpaper(list[nextIndex])
	if currentProcess != 0 {
		sway.OptionalKill(currentProcess)
	}
	*curr = nextIndex
}

func moveToPrevSway(list []string, curr *uint) {
	prevIndex := (*curr - 1 + uint(len(list))) % uint(len(list))
	currentProcess := sway.GetCurrentProcess()
	sway.SetWallpaper(list[prevIndex])
	if currentProcess != 0 {
		sway.OptionalKill(currentProcess)
	}
	*curr = prevIndex
}

func HandleSignals(command string) {
	daemon.AddCommand(daemon.StringFlag(&command, "quit"), syscall.SIGQUIT, TermHandler)
	daemon.AddCommand(daemon.StringFlag(&command, "quit"), syscall.SIGTERM, TermHandler)
	daemon.AddCommand(daemon.StringFlag(&command, "next"), syscall.SIGUSR1, NextHandler)
	daemon.AddCommand(daemon.StringFlag(&command, "prev"), syscall.SIGUSR2, PrevHandler)
}
