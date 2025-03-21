package main

import (
	"fmt"
	"log"
	"os"

	"github.com/vic/nix-versions/packages/app"
)

func main() {
	args := app.NewAppArgs()
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, app.AppHelp)
		os.Exit(1)
	}
	err := args.ParseAndRun(os.Args[1:])
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
}
