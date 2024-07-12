package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/st3iny/nextcloud-status-command/internal/command"
)

//go:generate go run ../../scripts/generateEmojis.go

func main() {
	cmd := ""
	if len(os.Args) >= 2 {
		cmd = os.Args[1]
	}
	if strings.HasPrefix(cmd, "-") {
		cmd = ""
	}

	var err error
	switch cmd {
	case "":
		err = command.RunUpdate()
	case "auth":
		err = command.RunAuth()
	case "clear":
		err = command.RunClear()
	case "get":
		err = command.RunGet()
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
