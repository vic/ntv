package main

import (
	"log"
	"os"

	"github.com/vic/nix-versions/packages/app"
)

func main() {
	err := app.NewAppArgs().ParseAndRun(os.Args[1:])
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
}
