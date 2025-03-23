package app

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/vic/ntv/packages/app/list"
	"github.com/vic/ntv/packages/app/new"
)

//go:embed HELP
var HELP string

//go:embed VERSION
var AppVersion string

//go:embed REVISION
var AppRevision string

type AppArgs struct {
	OnHelp    func() `long:"help" short:"h"`
	OnVersion func() `long:"version"`
	Help      bool
	Version   bool
}

func NewAppArgs() *AppArgs {
	var a AppArgs
	a.OnHelp = func() {
		a.Help = true
	}
	a.OnVersion = func() {
		a.Version = true
	}
	return &a
}

func HelpAndExit(exitCode int) {
	t, err := template.New("HELP").Parse(HELP)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	var out = os.Stdout
	if exitCode != 0 {
		out = os.Stderr
	}
	err = t.Execute(out, map[string]interface{}{
		"Cmd":     "nvm",
		"Version": Version(),
	})
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func Version() string {
	out := &bytes.Buffer{}
	fmt.Fprint(out, strings.TrimSpace(AppVersion))
	revision := strings.TrimSpace(AppRevision)
	if revision != "" {
		fmt.Fprint(out, " (%s)", revision)
	}
	return out.String()
}

func VersionAndExit() {
	fmt.Println(Version())
	os.Exit(0)
}

func (a *AppArgs) ParseAndRun(args []string) error {
	parser := flags.NewParser(a, flags.IgnoreUnknown)
	extra, err := parser.ParseArgs(args[1:])
	if err != nil {
		return err
	}

	var cmd string
	if len(extra) > 0 {
		cmd = extra[0]
	}

	if cmd == "new" || cmd == "init" {
		return new.NewInitArgs().ParseAndRun(extra[1:])
	}

	if cmd == "list" {
		list.HelpAndExit("nvm list")
		return list.NewListArgs().ParseAndRun(extra[1:])
	}

	if a.Help {
		HelpAndExit(0)
		return nil
	}

	// // Default action is search.
	// return NewSearchArgs().ParseAndRun(extra)
	return nil
}
