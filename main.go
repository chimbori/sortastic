package main

import (
	"fmt"
	"os"
	"path/filepath"

	"go.chimbori.app/sortastic/conf"
	"go.chimbori.app/sortastic/web"
)

var commands = map[string]func([]string){
	"web": web.Web,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(1)
	}

	cmd, ok := commands[os.Args[1]]
	if !ok {
		fmt.Fprintln(os.Stderr, usage())
		os.Exit(1)
	}

	conf.Config = conf.ReadConfig()
	cmd(os.Args[2:])
}

func usage() string {
	s := fmt.Sprintf("Usage: %s <command> [options]\nAvailable commands:\n", filepath.Base(os.Args[0]))
	for k := range commands {
		s += " - " + k + "\n"
	}
	return s
}
