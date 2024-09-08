package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	u "github.com/Hcode00/hpaper/utils"
	"github.com/sevlyar/go-daemon"
)

const USAGE = `Usage:-
  hpaper start [directory] [duration in seconds] [maximum number of pictures to preload]
  hpaper start [image file]
  hpaper [next|prev|status|quit]`

var (
	nextChan   = make(chan struct{}, 1)
	statusChan = make(chan struct{}, 1)
	prevChan   = make(chan struct{}, 1)
	quitChan   = make(chan struct{}, 1)
)

type Hpaper struct {
	Path       string
	List       []string
	MaxToLoad  uint
	CurrentIdx uint
	Interval   time.Duration
	IsDir      bool
}

// Function to handle service startup
func (hpaper *Hpaper) StartService() error {
	println(0)
	log.Println("Starting service...")

	println(1)
	u.StartHyprpaper()
	time.Sleep(500 * time.Millisecond)
	println("hyprpaper is installed")
	isDir, err := u.IsDir(u.AbsPath(hpaper.Path))
	if err != nil {
		panic(err.Error())
	}
	if !isDir {

		hpaper.IsDir = false
		err := u.SetWallpaper(hpaper.Path)
		if err != nil {
			return err
		}
	} else {
		hpaper.Path = os.Args[2]
		hpaper.IsDir = true
		files := u.ListFiles(hpaper.Path)
		if len(files) == 0 {
			return errors.New("[ERROR]: This Directory does not contain any of the supported image format")
		}
		hpaper.List = make([]string, 0, len(files))
		println("MaxToLoad", hpaper.MaxToLoad)
		for _, file := range files {
			f := fmt.Sprintf("%s%s", hpaper.Path, file)
			hpaper.List = append(hpaper.List, f)

		}
		_ = hpaper.Status()
		go WaitAndSet(hpaper.Interval, hpaper.List, hpaper.MaxToLoad, &hpaper.CurrentIdx)
	}
	return nil
}

// Function to handle next wallpaper action
func (h *Hpaper) Next() error {
	if !h.IsDir {
		return errors.New("You Preloaded a file not a directory")
	}
	log.Println("Changing to next wallpaper")
	// Implement logic to change to next wallpaper
	return nil
}

// Function to handle previous wallpaper action
func (h *Hpaper) Previous() error {
	if !h.IsDir {
		return errors.New("You Preloaded a file not a directory")
	}
	log.Println("Changing to previous wallpaper")
	// Implement logic to change to previous wallpaper
	return nil
}

// Function to get wallpaper status
func (h *Hpaper) Status() error {
	// Implement logic to retrieve wallpaper status
	u.ListLoaded()
	u.ListActive()
	return nil
}

func TermHandler(sig os.Signal) error {
	if sig == syscall.SIGQUIT || sig == syscall.SIGTERM {
		log.Println("terminating ...")
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

func StatusHandler(sig os.Signal) error {
	if sig == syscall.Signal(100) {
		statusChan <- struct{}{}
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
			println("------------------- after index = ", (curr+i)%uint(len(list)), " ----------------------")
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
			log.Println("Next signal triggered")
			moveToNext(list, &curr, &bufferIndex, buffer, maxCap)
			resetTicker()
		case <-prevChan:
			log.Println("Prev signal triggered")
			moveToPrev(list, &curr, &bufferIndex, buffer, maxCap)
			resetTicker()
		case <-statusChan:
			u.ListLoaded()
			return
		case <-quitChan:
			log.Println("Quitting wallpaper rotation...")
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
		log.Printf("Failed to preload %s: %v", list[nextToLoad], err)
	}

	// Set new wallpaper
	err = u.SetOneWallpaper(list[nextIndex])
	if err != nil {
		log.Printf("Failed to set wallpaper %s: %v", list[nextIndex], err)
	}

	// Unload previous wallpaper
	prevToUnload := (*curr - maxCap/2 + uint(len(list))) % uint(len(list))
	toUnload := list[prevToUnload]
	err = u.Unload(toUnload)
	if err != nil {
		log.Printf("Failed to unload %s: %v", toUnload, err)
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
		log.Printf("Failed to preload %s: %v", list[prevToLoad], err)
	}

	// Set new wallpaper
	err = u.SetOneWallpaper(list[prevIndex])
	if err != nil {
		log.Printf("Failed to set wallpaper %s: %v", list[prevIndex], err)
	}

	// Small delay to avoid potential race conditions
	time.Sleep(100 * time.Millisecond)

	// Unload current wallpaper

	toUnload := list[((*curr)%uint(len(list))+maxCap/2)%uint(len(list))]
	err = u.Unload(toUnload)
	if err != nil {
		log.Printf("Failed to unload %s: %v", toUnload, err)
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
	daemon.AddCommand(daemon.StringFlag(&command, "status"), syscall.Signal(100), StatusHandler)
}
