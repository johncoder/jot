package main

import (
	"os"

	"github.com/johncoder/jot/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
