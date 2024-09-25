package swaybg

import (
	"errors"
	"os"
	"os/exec"
	"strconv"

	u "github.com/Hcode00/hpaper/utils"
)

func GetCurrentProcess() int {
	cmd := exec.Command("pgrep", "swaybg")
	output, err := cmd.Output()
	if err != nil {
		u.LOG.Error("No previous wallpaper process")
		return 0
	}
	pid, err := praseStrToUint(string(output))
	if err != nil {
		u.LOG.Error(err.Error())
		return 0
	}
	return int(pid)
}

func UnloadAll() {
	u.Ex("killall", "swaybg")
}

func SetWallpaper(path string) error {
	u.LOG.Debug("Trying to Set Wallpaper at" + path)
	// Check if the process is running
	cmd := exec.Command("swaybg", "-m", "fill", "-i", path)
	err := cmd.Start()
	if err != nil {
		u.LOG.Error("failed to start swaybg -> " + err.Error())
		return err
	}
	// Detach the process
	err = cmd.Process.Release()
	if err != nil {
		u.LOG.Error("Failed to release process")
		return err
	}
	u.LOG.Debug("swaybg started successfully in the background.")
	return nil
}

func praseStrToUint(pid string) (uint, error) {
	if len(pid) == 0 {
		return 0, errors.New("No Other Instances of swaybg")
	}
	// Remove the newline character
	pid = pid[:len(pid)-1]
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		errMsg := "can't parse string" + err.Error()
		u.LOG.Warn(errMsg)
		return 0, errors.New(errMsg)
	}
	return uint(pidInt), nil
}

func OptionalKill(pid int) {
	u.LOG.Debug("Trying to kill previous wallpaper process")
	// get the process
	process, err := os.FindProcess(pid)
	err = process.Kill()
	if err != nil {
		u.LOG.Warn("Failed to kill previous wallpaper process -> " + err.Error())
	}
	_, err = process.Wait()
	if err != nil {
		u.LOG.Warn("Failed to kill previous wallpaper process -> " + err.Error())
	}
}
