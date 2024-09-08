package daemon

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/sevlyar/go-daemon"

	s "github.com/Hcode00/hpaper/service"
)

const (
	daemonName = "hpaper"
)

var (
	HOME, e = os.UserHomeDir()
	pidFile = fmt.Sprintf("%s/.hpaper/hpaper.pid", HOME)
	logFile = fmt.Sprintf("%s/.hpaper/hpaper.log", HOME)
)

var Cntxt = &daemon.Context{
	PidFileName: pidFile,
	PidFilePerm: 0o644,
	LogFileName: logFile,
	LogFilePerm: 0o640,
	WorkDir:     "./",
	Umask:       0o27,
	Args:        []string{"[hpaper]"},
}

func StartDaemon(cntxt *daemon.Context, service *s.Hpaper) (*daemon.Context, error) {
	if e != nil {
		panic("can't find home directory")
	}
	log.Println("Starting daemon...")

	dir := filepath.Dir(Cntxt.PidFileName)

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Printf("Error creating directory: %s", err.Error())
		return nil, err
	}

	d, err := cntxt.Reborn()
	if err != nil {
		return nil, err
	}

	log.Println("hpaper daemon started")
	log.Printf("daemon PID: %d", d.Pid)
	log.Printf("process PID: %d", syscall.Getpid())
	WritePIDFile(syscall.Getpid())

	// Initialize your service
	if err := service.StartService(); err != nil {
		return nil, err
	}
	log.Println("- - - - - - - - - - - - - - -")
	log.Println("Service started successfully")
	log.Println("- - - - - - - - - - - - - - -")
	err = daemon.ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}

	return cntxt, nil
}

func HandleExternalCommand(cntxt *daemon.Context, command string, service *s.Hpaper) {
	d, err := cntxt.Search()
	if err != nil {
		log.Fatalf("Unable to send signal to the daemon\nMake sure the app is running using [hpaper start]")
	}

	switch command {
	case "next":
		err := d.Signal(syscall.SIGUSR1)
		if err != nil {
			log.Fatalf("Failed to send next signal: %v", err)
		}
	case "prev":
		err := d.Signal(syscall.SIGUSR2)
		if err != nil {
			log.Fatalf("Failed to send prev signal: %v", err)
		}
	case "status":
		service.Status()
	case "quit":
		err := d.Signal(syscall.SIGTERM)
		if err != nil {
			log.Fatalf("Failed to send quit signal: %v", err)
		}
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func WritePIDFile(pid int) {
	file, err := os.OpenFile(Cntxt.PidFileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
	defer file.Close()

	pidString := strconv.Itoa(pid)
	_, err = file.WriteString(pidString)
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
}

func ReadPID() (int, error) {
	pidStr, err := os.ReadFile(Cntxt.PidFileName)
	if err != nil {
		log.Printf("error reading PID:%s\n", err.Error())
		return 0, err
	}
	pid, err := strconv.ParseInt(string(pidStr), 10, 64)
	if err != nil {
		log.Printf("error reading PID:%s\n", err.Error())
		return 0, err
	}
	return int(pid), nil
}

func RemovePidFile() {
	err := os.Remove(Cntxt.PidFileName)
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
}

func SendQuit() {
	pid, err := ReadPID()
	if err != nil {
		println("[ERROR]: cannot terminate ", err)
		return
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		println("process not found")
	}
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		println("error sending signal,", err)
	}
	println("quiting ...")
}
