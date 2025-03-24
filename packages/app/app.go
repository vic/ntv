package app

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/vic/ntv/packages/app/help"
	"github.com/vic/ntv/packages/app/list"
	"github.com/vic/ntv/packages/app/new"
)

//go:embed HELP
var HELP string

//go:embed VERSION
var AppVersion string

//go:embed REVISION
var AppRevision string

var Help = help.CmdHelp{
	HelpTxt: HELP,
	HelpCtx: func(name string) any {
		return map[string]interface{}{
			"Cmd":     name,
			"Version": Version(),
		}
	},
}

var HelpDict = help.HelpDict{
	"init": new.Help,
	"list": list.Help,
}

type AppArgs struct {
	OnHelp    func() `long:"help" short:"h"`
	OnVersion func() `long:"version"`
	help      bool
	version   bool
}

func NewAppArgs() *AppArgs {
	var a AppArgs
	a.OnHelp = func() {
		a.help = true
	}
	a.OnVersion = func() {
		a.version = true
	}
	return &a
}

func Version() string {
	out := &bytes.Buffer{}
	fmt.Fprint(out, strings.TrimSpace(AppVersion))
	revision := strings.TrimSpace(AppRevision)
	if revision != "" {
		fmt.Fprintf(out, " (%s)", revision)
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

	if a.help {
		HelpDict.PrintHelpAndExit(Help, args, 0)
		return nil
	}

	var cmd string
	if len(extra) > 0 {
		cmd = extra[0]
	}

	if cmd == "init" {
		return new.NewInitArgs().ParseAndRun(extra[1:])
	}

	if cmd == "list" {
		return list.NewListArgs().ParseAndRun(extra[1:])
	}

	// // Default action is search.
	// return NewSearchArgs().ParseAndRun(extra)
	return nil
}
