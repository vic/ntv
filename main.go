package main

import (
	"log"
	"os"

	"github.com/vic/nix-versions/packages/app"
)

func main() {
	args, err := app.ParseCliArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	if err := app.MainAction(args); err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
}
