package utils

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

/*
Level 0 NO LOGS -
Level 1 ERRORS -
Level 2 WARNINGS -
Level 3 Logs -
*/
type Logger struct {
	WarningsMsgs uint
	ErrorsMsgs   uint
	DebugMsgs    uint
	Level        uint
}

func (l *Logger) Debug(msg string) {
	l.DebugMsgs += 1
	if l.Level > 2 {
		log.Println(msg)
	}
}

func (l *Logger) Warn(msg string) {
	l.WarningsMsgs += 1
	if l.Level > 1 {
		log.Println("[WARNING]:", msg)
	}
}
func (l *Logger) Log(msg string) {
	l.WarningsMsgs += 1
	if l.Level > 1 {
		println(msg)
	}
}

func (l *Logger) Error(msg string) {
	l.ErrorsMsgs += 1
	if l.Level > 0 {
		log.Println("[ERROR]:", msg)
	}
}

func (l *Logger) Panic(msg string) {
	l.ErrorsMsgs += 1
	panic("[ERROR]:" + msg)
}

func (l Logger) Status() {
	fmt.Printf("You Have %d Errors, %d Warnings and %d Debug Messages\n", l.ErrorsMsgs, l.WarningsMsgs, l.DebugMsgs)
}

var LOG = Logger{
	Level: 1,
}

const TIMEOUT = 5000 * time.Millisecond // Milliseconds

func ex(name string, args ...string) (string, error) {
	c := exec.Command(name, args...)
	out, err := c.Output()
	if err != nil {
		errMsg := err.Error()
		// if exit status 1
		if len(errMsg) > 0 && errMsg[len(errMsg)-1] == '1' {
			return errMsg + "\n", nil
		} else {
			return "", fmt.Errorf("command exited with status 1: %s %s", name, strings.Join(args, " "))
		}
	}
	return string(out), nil
}

func StartHyprpaper() {
	processName := "hyprpaper"

	// Check if the process is running
	cmd := exec.Command("pgrep", processName)
	output, err := cmd.Output()

	if err != nil && !strings.Contains(string(output), processName) {
		// Process is not running, start it
		LOG.Warn(processName + " is not running. Starting it..")

		cmd := exec.Command(processName)

		err := cmd.Start()
		if err != nil {
			LOG.Error("starting " + processName + " -> " + err.Error())
			return
		}

		// Detach the process
		cmd.Process.Release()
		LOG.Debug(processName + " started successfully in the background.")
	} else {
		LOG.Debug(processName + " is already running.\n")
	}
}

func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf(path + " does not exist")
		} else {
			return false, fmt.Errorf(err.Error())
		}
	}

	return fileInfo.IsDir(), nil
}

func ListFiles(path string) []string {
	files, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}
	list := make([]string, 0, len(files))
	for _, file := range files {
		if IsValidPicture(file.Name()) {
			list = append(list, file.Name())
			LOG.Debug("Added : " + file.Name())
		}
	}
	return list
}

func IsValidPicture(file string) bool {
	parts := strings.Split(file, ".")
	ext := strings.Trim(parts[len(parts)-1], " ")
	validExtensions := []string{"png", "jpg", "jpeg", "webp"}
	for _, e := range validExtensions {
		if ext == e {
			return true
		}
	}
	return false
}

func AbsPath(path string) string {
	if strings.Trim(path, " ")[0] == '~' {
		HOME, err := os.UserHomeDir()
		if err != nil {
			LOG.Panic("Can't Find Home Directory")
		}
		return HOME + path[1:]
	}
	return path
}

func SetWallpaper(path string) error {
	_, err := ex("hyprctl", "hyprpaper", "preload", path)
	if err != nil {
		err = UnloadAll()
		if err != nil {
			return err
		}
	}
	parts := strings.Split(path, "/")
	fileName := strings.Trim(parts[len(parts)-1], " ")
	if !IsValidPicture(fileName) {
		return errors.New("Not a valid picture extenstion")
	}
	_, err = ex("hyprctl", "hyprpaper", "preload", path)
	if err != nil {
		return err
	}
	_, err = ex("hyprctl", "hyprpaper", "wallpaper", ",", path)
	if err != nil {
		return err
	}
	return nil
}

func UnloadAll() error {
	_, err := ex("hyprctl", "hyprpaper", "unload", "all")
	if err != nil {
		return err
	}
	return nil
}

func ListLoaded() error {
	LOG.Log("\n----------------- Loaded Wallpapers -----------------")
	op, err := ex("hyprctl", "hyprpaper", "listloaded")
	if err != nil {
		return errors.New("[ERROR] from ListLoaded: " + err.Error())
	}
	print(op)
	LOG.Log("-----------------------------------------------------")
	return nil
}

func ListActive() error {
	LOG.Log("\n----------------- Active Wallpapers -----------------")
	op, err := ex("hyprctl", "hyprpaper", "listactive")
	if err != nil {
		return errors.New("[ERROR] from ListActive: " + err.Error())
	}
	print(op)
	LOG.Log("-----------------------------------------------------")
	return nil
}

func Unload(filePath string) error {
	LOG.Debug("Trying to Unload -|>" + filePath)
	op, err := ex("hyprctl", "hyprpaper", "unload", filePath)
	if err != nil {
		return err
	}
	LOG.Debug("status:" + op)
	LOG.Log("-----------------------------------------------------")
	return nil
}

func Preload(filePath string) error {
	LOG.Debug("Trying to Preload -|>" + filePath)
	op, err := ex("hyprctl", "hyprpaper", "preload", filePath)
	if err != nil {
		return err
	}
	LOG.Debug("status:" + op)
	return nil
}

func SetOneWallpaper(filePath string) error {
	LOG.Debug("Trying to Set Wallpaper at" + filePath)
	op, err := ex("hyprctl", "hyprpaper", "wallpaper", ",", filePath)
	if err != nil {
		return err
	}
	LOG.Debug("status:" + op)
	return nil
}

func RandomizeFileNames(files []string) []string {
	// Create a copy of the input slice to avoid modifying the original
	randomized := make([]string, len(files))
	copy(randomized, files)

	for i := len(randomized) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		randomized[i], randomized[j] = randomized[j], randomized[i]
	}
	return randomized
}
