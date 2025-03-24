package main

import (
	"log"
	"os"
	"os/exec"
	"path"
	"slices"

	"github.com/vic/ntv/packages/app"
)

func main() {
	name := path.Base(os.Args[0])
	args := slices.Concat([]string{name}, os.Args[1:])
	if len(os.Args) < 2 {
		app.HelpDict.PrintHelpAndExit(app.Help, args, 1)
	}
	err := app.NewAppArgs().ParseAndRun(args)
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			log.Fatal(string(ee.Stderr))
			os.Exit(ee.ExitCode())
		}
		log.Fatal(err)
		os.Exit(2)
	}
}
