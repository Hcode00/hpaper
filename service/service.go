package service

import (
	"errors"
	"os"
	"strconv"
	"syscall"
	"time"

	u "github.com/Hcode00/hpaper/utils"
	"github.com/sevlyar/go-daemon"
)

const USAGE = `Usage:-
  hpaper start [directory] [duration in seconds] [maximum number of pictures to preload]
  -r at the end -> can be used to randomize wallpaper list
  hpaper start [image file]
  hpaper [next|prev|status|quit]`

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

// Function to handle service startup
func (hpaper *Hpaper) StartService() error {
	u.LOG.Debug("Starting service...")
	u.StartHyprpaper()
	// give hyprpaper time to launch
	time.Sleep(1000 * time.Millisecond)
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

	go WaitAndSet(hpaper.Interval, hpaper.List, hpaper.MaxToLoad, &hpaper.CurrentIdx)
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

func WaitAndSet(t time.Duration, list []string, maxCap uint, Index *uint) {
	if len(list) == 0 || maxCap == 0 {
		return
	}

	u.UnloadAll()
	buffer := make([]string, maxCap)
	bufferIndex := *Index % uint(len(list))
	curr := bufferIndex

	// Preload function
	preload := func(index uint) {
		u.Preload(list[index])
	}

	// Preload initial set
	if maxCap == 1 {
		preload(curr)
	} else if maxCap == 2 {
		preload(curr)
		preload((curr + 1) % uint(len(list)))
	} else if maxCap%2 == 0 { // even
		preload(curr)
		preload((curr - 1 + uint(len(list))) % uint(len(list))) // one before
		for i := uint(1); i <= maxCap/2; i++ {
			preload((curr + i) % uint(len(list))) // two after
		}
	} else { // odd
		preload((curr - 1 + uint(len(list))) % uint(len(list))) // one before
		preload(curr)
		for i := uint(1); i <= maxCap/2; i++ {
			preload((curr + i) % uint(len(list))) // one after
		}
	}

	u.SetOneWallpaper(list[curr])
	u.ListLoaded()

	ticker := time.NewTicker(t)
	defer ticker.Stop()
	resetTicker := func() {
		ticker.Stop()
		ticker = time.NewTicker(t)
	}
	for {
		select {
		case <-ticker.C:
			moveToNext(list, &curr, &bufferIndex, buffer, maxCap)
		case <-nextChan:
			u.LOG.Debug("Next signal triggered")
			moveToNext(list, &curr, &bufferIndex, buffer, maxCap)
			resetTicker()
		case <-prevChan:
			u.LOG.Debug("Prev signal triggered")
			moveToPrev(list, &curr, &bufferIndex, buffer, maxCap)
			resetTicker()
		case <-quitChan:
			u.LOG.Warn("Quitting wallpaper rotation...")
			return
		}
	}
}

func moveToNext(list []string, curr *uint, bufferIndex *uint, buffer []string, maxCap uint) {
	nextIndex := (*curr + 1) % uint(len(list))
	nextToLoad := (nextIndex + maxCap/2) % uint(len(list))

	// Preload next wallpaper
	err := u.Preload(list[nextToLoad])
	if err != nil {
		u.LOG.Warn("Failed to Preload wallpaper" + list[nextToLoad] + "->" + err.Error())
	}

	// Set new wallpaper
	err = u.SetOneWallpaper(list[nextIndex])
	if err != nil {
		u.LOG.Warn("Failed to Set wallpaper" + list[nextIndex] + "->" + err.Error())
	}

	// Unload previous wallpaper
	prevToUnload := (*curr - maxCap/2 + uint(len(list))) % uint(len(list))
	toUnload := list[prevToUnload]
	err = u.Unload(toUnload)
	if err != nil {
		u.LOG.Warn("Failed to unload" + toUnload + "->" + err.Error())
	}

	*curr = nextIndex

	if maxCap > 1 {
		buffer[*bufferIndex] = list[nextToLoad]
		*bufferIndex = (*bufferIndex + 1) % maxCap
	}
}

func moveToPrev(list []string, curr *uint, bufferIndex *uint, buffer []string, maxCap uint) {
	prevIndex := (*curr - 1 + uint(len(list))) % uint(len(list))
	prevToLoad := (prevIndex - maxCap/2 + uint(len(list))) % uint(len(list))

	// Preload previous wallpaper
	err := u.Preload(list[prevToLoad])
	if err != nil {
		u.LOG.Warn("Failed to Preload wallpaper" + list[prevIndex] + "->" + err.Error())
	}

	// Set new wallpaper
	err = u.SetOneWallpaper(list[prevIndex])
	if err != nil {
		u.LOG.Warn("Failed to Set wallpaper" + list[prevIndex] + "->" + err.Error())
	}

	// Small delay to avoid potential race conditions
	time.Sleep(100 * time.Millisecond)

	// Unload current wallpaper

	toUnload := list[((*curr)%uint(len(list))+maxCap/2)%uint(len(list))]
	err = u.Unload(toUnload)
	if err != nil {
		u.LOG.Warn("Failed to unload" + toUnload + "->" + err.Error())
	}
	*curr = prevIndex

	if maxCap > 1 {
		*bufferIndex = (*bufferIndex - 1 + maxCap) % maxCap
		buffer[*bufferIndex] = list[prevToLoad]
	}
}

func HandleSignals(command string) {
	daemon.AddCommand(daemon.StringFlag(&command, "quit"), syscall.SIGQUIT, TermHandler)
	daemon.AddCommand(daemon.StringFlag(&command, "quit"), syscall.SIGTERM, TermHandler)
	daemon.AddCommand(daemon.StringFlag(&command, "next"), syscall.SIGUSR1, NextHandler)
	daemon.AddCommand(daemon.StringFlag(&command, "prev"), syscall.SIGUSR2, PrevHandler)
}
