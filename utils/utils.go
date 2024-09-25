package utils

import (
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
	l.DebugMsgs += 1
	println(msg)
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
	Level: 2,
}

const TIMEOUT = 5000 * time.Millisecond // Milliseconds

func Ex(name string, args ...string) (string, error) {
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
