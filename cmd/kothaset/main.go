package main

import (
	"os"

	"github.com/shantoislamdev/kothaset/internal/cli"
	"github.com/shantoislamdev/kothaset/internal/log"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Error("command failed", "error", err)
		os.Exit(1)
	}
}
