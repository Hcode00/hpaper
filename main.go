package main

import (
	"os"
	"strconv"
	"time"

	d "github.com/Hcode00/hpaper/daemon"
	s "github.com/Hcode00/hpaper/service"
	u "github.com/Hcode00/hpaper/utils"
	w "github.com/Hcode00/hpaper/wallpapers"
)

var service *s.Hpaper

func main() {
	if len(os.Args) < 2 {
		print(s.USAGE)
		return
	}
	command := os.Args[1]

	switch command {
	case "--help", "help":
		help()
	case "start":
		arg2 := os.Args[2]
		if len(os.Args) < 3 {
			help()
			return
		}
		isDir, _ := u.IsDir(u.AbsPath(arg2))
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
				Path:       arg2,
				Randomize:  isRandom,
			}
			ctx, err := d.StartDaemon(d.Cntxt, service)
			if err != nil {
				u.LOG.Panic("Unable to run ->" + err.Error())
			}
			defer u.LOG.Debug("Service Ended.")
			defer ctx.Release()
			defer d.RemovePidFile()
		} else if u.IsValidPicture(arg2) {
			u.StartHyprpaper()
			// give hyprpaper time to launch
			time.Sleep(1000 * time.Millisecond)
			u.SetWallpaper(arg2)
		} else {
			u.LOG.Panic("Invalid Command")
		}

	case "next", "prev", "status":
		d.HandleExternalCommand(d.Cntxt, command, service)
	case "download":
		if len(os.Args) < 6 {
			help()
			return
		}
		dir := os.Args[2]
		numStr := os.Args[3]
		width := os.Args[4]
		height := os.Args[5]

		isDir, err := u.IsDir(u.AbsPath(dir))
		if err != nil {
			u.LOG.Panic(err.Error())
		}
		if !isDir {
			u.LOG.Panic("Please Specify a directory to save wallpapers in")
		}
		max, err := strconv.Atoi(numStr)
		if err != nil || max < 0 {
			u.LOG.Error(numStr + "not a valid or usable number")
			return
		}
		if max < 1 || max > 20 {
			u.LOG.Panic("Downloaded Range from 1 to 20")
		}
		isWebp := false
		if len(os.Args) > 6 {
			if os.Args[6] == "-w" {
				isWebp = true
			}
		}
		err = w.DownloadFile(u.AbsPath(dir), width, height, uint(max), isWebp)
		if err != nil {
			u.LOG.Panic(err.Error())
		}
	case "quit":
		d.SendQuit()
	default:
		println("hpaper", command, "-> unknown command")
		help()
	}
}

func help() {
	print(s.USAGE)
}
