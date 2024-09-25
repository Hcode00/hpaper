package main

import (
	"os"

	d "github.com/Hcode00/hpaper/daemon"
	s "github.com/Hcode00/hpaper/service"
)

func main() {
	var service *s.Hpaper
	if len(os.Args) < 2 {
		print(s.USAGE)
		return
	}
	command := os.Args[1]
	switch command {
	case "--help", "help":
		s.Help()
	case "start":
		d.StartApp(command, service)
	case "next", "prev":
		d.HandleExternalCommand(d.Cntxt, command, service)
	case "download":
		d.Download()
	case "quit":
		d.SendQuit()
	default:
		println("hpaper", command, "-> Unknown command")
		s.Help()
	}
}
