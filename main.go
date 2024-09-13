package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	d "github.com/Hcode00/hpaper/daemon"
	s "github.com/Hcode00/hpaper/service"
	u "github.com/Hcode00/hpaper/utils"
)

var service *s.Hpaper

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
			if err != nil || max < 0 {
				u.LOG.Error(maxLoaded + "not a valid or usable number")
				return
			}
			sec, err := strconv.Atoi(seconds)
			if err != nil || sec < 0 {
				u.LOG.Error(maxLoaded + "not a valid or usable number of seconds")
				return
			}
			isRandom := false
			if len(os.Args) > 5 {
				if os.Args[5] == "-r" {
					isRandom = true
				}
			}
			s.HandleSignals(command)

			service = &s.Hpaper{
				MaxToLoad:  uint(max),
				CurrentIdx: 0,
				Interval:   time.Duration(sec) * time.Second,
				Path:       path,
				Randomize:  isRandom,
			}
			ctx, err := d.StartDaemon(d.Cntxt, service)
			if err != nil {
				u.LOG.Panic("Unable to run ->" + err.Error())
			}
			defer u.LOG.Debug("Service Ended.")
			defer ctx.Release()
			defer d.RemovePidFile()
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
