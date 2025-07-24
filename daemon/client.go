package daemon

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"hpaper/backends"
)

const (
	sockPath = "/tmp/hpaper_daemon.sock"
)

func RunClientMode(action backends.WallpaperAction) {
	conn, err := net.DialTimeout("unix", sockPath, 5*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not connect to hpaper daemon. Is it running? %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte(string(action) + "\n")); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to send command to daemon: %v\n", err)
		os.Exit(1)
	}

	response, err := io.ReadAll(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read response from daemon: %v\n", err)
		os.Exit(1)
	}

	respStr := strings.TrimSpace(string(response))

	if respStr == "OK" {
		fmt.Printf("Command '%s' sent successfully.\n", action)
	} else if strings.HasPrefix(respStr, "Error:") {
		fmt.Fprintf(os.Stderr, "Daemon error: %s\n", respStr)
		os.Exit(1)
	} else {
		fmt.Println(respStr)
	}
}
