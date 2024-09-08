package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	d "github.com/Hcode00/hpaper/daemon"
	s "github.com/Hcode00/hpaper/service"
	u "github.com/Hcode00/hpaper/utils"
)

var service *s.Hpaper
var (
	nextChan = make(chan struct{})
	prevChan = make(chan struct{})
	quitChan = make(chan struct{})
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(s.USAGE)
		return
	}
	command := os.Args[1]

	switch command {
	case "--help", "help":
		help()
	case "start":

		path := os.Args[2]
		if len(os.Args) < 3 {
			help()
			return
		}
		isDir, _ := u.IsDir(u.AbsPath(path))
		if isDir {
			seconds := os.Args[3]
			maxLoaded := os.Args[4]
			max, err := strconv.Atoi(maxLoaded)
			if err != nil {
				println(maxLoaded, "not a valid number")
				return
			}
			sec, err := strconv.Atoi(seconds)
			if err != nil {
				println(seconds, "not a valid number")
				return
			}
			s.HandleSignals(command)

			service = &s.Hpaper{
				MaxToLoad:  uint(max),
				CurrentIdx: 0,
				Interval:   time.Duration(sec) * time.Second,
				Path:       path,
			}
			ctx, err := d.StartDaemon(d.Cntxt, service)
			if err != nil {
				log.Fatal("[ERROR]: Unable to run -> ", err)
			}
			defer ctx.Release()
			defer d.RemovePidFile()
			defer log.Println("Service Ended.")
		} else {
			if u.IsValidPicture(path) {
				u.SetWallpaper(path)
			}
		}

	case "next", "prev", "status":
		d.HandleExternalCommand(d.Cntxt, command, service)
	case "quit":
		d.SendQuit()
	default:
		println("hpaper", command, "-> unknown command")
		help()
	}
}

func help() {
	print(s.USAGE)
	println(`
      next -> set next wallpaper in the list
      prev -> set previous wallpaper in the list
      status -> show current wallpaper name
      and preloaded wallpapers
      help -> show helpful information
      quit -> stop rotaing wallpapers`)
}
