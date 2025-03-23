package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/vic/ntv/packages/app"
)

func main() {
	if len(os.Args) < 2 {
		app.HelpAndExit(1)
	}
	err := app.NewAppArgs().ParseAndRun(os.Args)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			log.Fatal(string(ee.Stderr))
			os.Exit(ee.ExitCode())
		}
		log.Fatal(err)
		os.Exit(2)
	}
}
