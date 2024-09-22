package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"

	s "github.com/Hcode00/hpaper/service"
	u "github.com/Hcode00/hpaper/utils"
	w "github.com/Hcode00/hpaper/wallpapers"
)

var (
	HOME, e = os.UserHomeDir()
	pidFile = fmt.Sprintf("%s/.hpaper/hpaper.pid", HOME)
)

var Cntxt = &daemon.Context{
	PidFileName: pidFile,
	PidFilePerm: 0o644,
	WorkDir:     "./",
	Umask:       0o27,
	Args:        []string{"[hpaper]"},
}

func StartDaemon(cntxt *daemon.Context, service *s.Hpaper) (*daemon.Context, error) {
	if e != nil {
		u.LOG.Panic("can't find home directory")
	}
	u.LOG.Debug("Starting daemon...")

	dir := filepath.Dir(Cntxt.PidFileName)

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	d, err := cntxt.Reborn()
	if err != nil {
		return nil, err
	}

	u.LOG.Debug("hpaper daemon started")
	u.LOG.Debug("daemon PID: " + strconv.Itoa(d.Pid))
	u.LOG.Debug("process PID: " + strconv.Itoa(syscall.Getpid()))
	WritePIDFile(syscall.Getpid())

	// Initialize your service
	if err := service.StartService(); err != nil {
		return nil, err
	}
	err = daemon.ServeSignals()
	if err != nil {
		return nil, err
	}

	return cntxt, nil
}

func Download() {
	if len(os.Args) < 6 {
		s.Help()
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
}

func StartApp(command string, service *s.Hpaper) {
	arg2 := os.Args[2]
	if len(os.Args) < 3 {
		s.Help()
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
		ctx, err := StartDaemon(Cntxt, service)
		if err != nil {
			u.LOG.Panic("Unable to run ->" + err.Error())
		}
		defer u.LOG.Debug("Service Ended.")
		defer ctx.Release()
		defer RemovePidFile()
	} else if u.IsValidPicture(arg2) {
		u.StartHyprpaper()
		// give hyprpaper time to launch
		time.Sleep(1000 * time.Millisecond)
		u.SetWallpaper(arg2)
	} else {
		u.LOG.Panic("Invalid Command")
	}
}

func HandleExternalCommand(cntxt *daemon.Context, command string, service *s.Hpaper) {
	d, err := cntxt.Search()
	if err != nil {
		u.LOG.Error("Unable to send signal to the daemon\nMake sure the app is running using [hpaper start]")
	}

	switch command {
	case "next":
		err := d.Signal(syscall.SIGUSR1)
		if err != nil {
			u.LOG.Error("Failed to send next signal:" + err.Error())
		}
	case "prev":
		err := d.Signal(syscall.SIGUSR2)
		if err != nil {
			u.LOG.Error("Failed to send prev signal:" + err.Error())
		}
	case "status":
		err = u.ListLoaded()
		if err != nil {
			u.LOG.Panic("Failed to reach hyprpaper\nare you sure hyprpaper is running?")
		}
		err := u.ListActive()
		if err != nil {
			u.LOG.Panic("Failed to reach hyprpaper\nare you sure hyprpaper is running?")
		}
	case "quit":
		err := d.Signal(syscall.SIGTERM)
		if err != nil {
			u.LOG.Error("Failed to send quit signal:" + err.Error())
		}
	default:
		u.LOG.Error("Unknown command:" + command)
	}
}

func WritePIDFile(pid int) {
	file, err := os.OpenFile(Cntxt.PidFileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		u.LOG.Error(err.Error())
	}
	defer file.Close()

	pidString := strconv.Itoa(pid)
	_, err = file.WriteString(pidString)
	if err != nil {
		u.LOG.Error(err.Error())
	}
}

func ReadPID() (int, error) {
	pidStr, err := os.ReadFile(Cntxt.PidFileName)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.ParseInt(string(pidStr), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(pid), nil
}

func RemovePidFile() {
	err := os.Remove(Cntxt.PidFileName)
	if err != nil {
		u.LOG.Error(err.Error())
	}
}

func SendQuit() {
	pid, err := ReadPID()
	if err != nil {
		u.LOG.Error("Cannot terminate" + err.Error())
		return
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		u.LOG.Error("Process not found" + err.Error())
	}
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		u.LOG.Error("Sending signal" + err.Error())
	}
	u.LOG.Error("Quiting ...")
}
