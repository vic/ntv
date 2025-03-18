package main

import (
	"log"
	"os"

	"github.com/vic/nix-versions/packages/app"
)

func main() {
	app := app.App()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
