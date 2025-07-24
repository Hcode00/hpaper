package backends

import (
	"fmt"
	"hpaper/config"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type SwayBGBackend struct{}

func NewSwayBGBackend() *SwayBGBackend {
	return &SwayBGBackend{}
}

func (s *SwayBGBackend) SetWallpaper(imagePath string, conf config.Config) error {
	fmt.Printf("Setting wallpaper to: %s\n", imagePath)

	cmdFindPIDs := exec.Command("pgrep", "swaybg")
	pidOutput, err := cmdFindPIDs.CombinedOutput()
	if err == nil {
		pidsStr := strings.Fields(strings.TrimSpace(string(pidOutput)))
		if len(pidsStr) > 0 {
			fmt.Printf("Found %d existing swaybg processes. Killing them...\n", len(pidsStr))
			for _, pidStr := range pidsStr {
				pid, parseErr := strconv.Atoi(pidStr)
				if parseErr != nil {
					log.Printf("Warning: Could not parse swaybg PID '%s': %v", pidStr, parseErr)
					continue
				}

				process, procErr := os.FindProcess(pid)
				if procErr != nil {
					log.Printf("Warning: Could not find process with PID %d: %v", pid, procErr)
					continue
				}
				err := process.Signal(syscall.SIGTERM)
				if err != nil {
					log.Printf("Warning: Failed to send SIGTERM to swaybg (PID %d): %v. Trying SIGKILL...", pid, err)
					err = process.Signal(syscall.SIGKILL)
					if err != nil {
						log.Printf("Error: Failed to send SIGKILL to swaybg (PID %d): %v. This process might be orphaned.", pid, err)
					} else {
						fmt.Printf("Successfully SIGKILLed swaybg (PID %d).\n", pid)
					}
				} else {
					fmt.Printf("Successfully SIGTERMd swaybg (PID %d).\n", pid)
				}
			}
			time.Sleep(100 * time.Millisecond)
		} else {
			fmt.Println("No existing swaybg process found.")
		}
	} else {
		fmt.Println("No existing swaybg process found (pgrep outputted no PIDs or failed):", err)
	}

	var cmdNewSwaybg *exec.Cmd
	if conf.SwaybgMode == "" && conf.SwaybgOutput == "" {
		cmdNewSwaybg = exec.Command("swaybg", "-i", imagePath)
	} else if conf.SwaybgMode != "" && conf.SwaybgOutput == "" {
		cmdNewSwaybg = exec.Command("swaybg", "-m", conf.SwaybgMode, "-i", imagePath)
	} else if conf.SwaybgMode == "" && conf.SwaybgOutput != "" {
		cmdNewSwaybg = exec.Command("swaybg", "-o", conf.SwaybgOutput, "-i", imagePath)
	} else {
		cmdNewSwaybg = exec.Command("swaybg", "-m", conf.SwaybgMode, "-o", conf.SwaybgOutput, "-i", imagePath)
	}

	cmdNewSwaybg.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmdNewSwaybg.Stdout = nil
	cmdNewSwaybg.Stderr = nil

	err = cmdNewSwaybg.Start()
	if err != nil {
		return fmt.Errorf("error launching new swaybg process: %v", err)
	}
	fmt.Printf("New swaybg process launched (PID: %d) with updated wallpaper.\n", cmdNewSwaybg.Process.Pid)

	if conf.PywalEnabled {
		parts := strings.Fields(conf.PywalCommand)

		if len(parts) == 0 {
			log.Println("Warning: pywal_command in config is empty. Pywal will not run.")
			return nil
		}

		pywalExecutable := parts[0]
		pywalArgs := parts[1:]

		pywalArgs = append(pywalArgs, "-i", imagePath)

		log.Printf("Attempting to run pywal: %s %s", pywalExecutable, strings.Join(pywalArgs, " "))
		walCmd := exec.Command(pywalExecutable, pywalArgs...)

		walOutput, walErr := walCmd.CombinedOutput()
		if walErr != nil {
			return fmt.Errorf("error running wal: %v\nOutput: %s", walErr, walOutput)
		}
		fmt.Println("Pywal updated successfully.")
	} else {
		fmt.Println("Pywal integration skipped.")
	}

	return nil
}
