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

const TIMEOUT = 5000 * time.Millisecond // Milliseconds

func ex(name string, args ...string) (string, error) {
	c := exec.Command(name, args...)
	out, err := c.Output()
	if err != nil {
		errMsg := err.Error()
		// if exit status 1
		if len(errMsg) > 0 && errMsg[len(errMsg)-1] == '1' {
			return fmt.Sprintf("%s\n", errMsg), nil
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
		fmt.Printf("%s is not running. Starting it...\n", processName)

		cmd := exec.Command(processName)

		err := cmd.Start()
		if err != nil {
			fmt.Printf("Error starting %s: %v\n", processName, err)
			return
		}

		// Detach the process
		cmd.Process.Release()
		fmt.Printf("%s started successfully in the background.\n", processName)
	} else {
		fmt.Printf("%s is already running.\n", processName)
	}
}

func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("[ERROR]: %s does not exist\n", path)
		} else {
			return false, fmt.Errorf("[ERROR]: %s\n", err.Error())
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
			println("appending : ", file.Name())
			list = append(list, file.Name())
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
			panic("[ERROR]: Can't Find Home Directory")
		}
		return fmt.Sprintf("%s%s", HOME, path[1:])
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
		println("[ERROR]: ", err.Error())
		return err
	}
	_, err = ex("hyprctl", "hyprpaper", "wallpaper", ",", path)
	if err != nil {
		print("[ERROR]: ")
		println(err.Error())
		return err
	}
	return nil
}

func PreloadBulk(names []string, path string) error {
	for _, name := range names {
		p := fmt.Sprintf("%s%s", path, name)
		_, err := ex("hyprctl", "hyprpaper", "preload", p)
		if err != nil {
			println("[ERROR]: ", err.Error())
			return err
		}
	}
	return nil
}

func UnloadAll() error {
	_, err := ex("hyprctl", "hyprpaper", "unload", "all")
	if err != nil {
		println("[ERROR] from connectAndUnloadAll: ", err.Error())
	}
	return nil
}

func ListLoaded() error {
	println("\n################# Loaded #################")
	op, err := ex("hyprctl", "hyprpaper", "listloaded")
	if err != nil {
		println("[ERROR] from ListLoaded: ", err.Error())
	}
	print(op)
	println("##########################################\n")
	return nil
}

func ListActive() error {
	println("\n################# Active Wallpaper #################")
	op, err := ex("hyprctl", "hyprpaper", "listactive")
	if err != nil {
		println("[ERROR] from ListActive: ", err.Error())
	}
	print(op)
	println("\n####################################################")
	return nil
}

func Unload(filePath string) error {
	op, err := ex("hyprctl", "hyprpaper", "unload", filePath)
	if err != nil {
		println("[ERROR]: ", err.Error())
	}
	log.Print("Trying to Unload -|> ", filePath, " | status: ", op)
	return nil
}

func Preload(filePath string) error {
	log.Print("Trying to Preload -|> ", filePath)
	op, err := ex("hyprctl", "hyprpaper", "preload", filePath)
	if err != nil {
		return err
	}

	log.Println(" | status: ", op)
	return nil
}

func SetOneWallpaper(filePath string) error {
	log.Print("Trying to Set Wallpaper {", filePath)
	op, err := ex("hyprctl", "hyprpaper", "wallpaper", ",", filePath)
	if err != nil {
		return err
	}
	log.Print(filePath, "} | status: ", op)
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
